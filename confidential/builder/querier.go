package builder

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const ledgerEntryNotFound = "entryNotFound"

// LedgerQuerier is the minimal interface needed to query ledger state.
// Both rpc.Client and websocket.Client satisfy this interface.
//
// Two separate trust decisions apply, and they are not the same one. Trusting the endpoint
// is the caller's responsibility: a false encryption key can disclose a transfer amount
// before rippled rejects the transaction, so builders must be pointed at an endpoint the
// caller trusts. Conformance of the implementation is not assumed: this interface is public,
// so readEntry re-derives snapshot consistency from whatever response it is handed rather
// than taking it on faith.
//
// Custom implementations must return server errors unchanged, or preserve them with error
// wrapping, so builders can classify an entryNotFound result. If submission fails because
// the selected state is stale, the caller must rebuild.
type LedgerQuerier interface {
	GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error)
	GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error)
}

// ledgerSnapshot pins one build to a single validated ledger, and owns the querier it was
// selected against so the two cannot drift apart. beginBuild selects the index and the first
// read binds the snapshot to the hash the server reports for it.
type ledgerSnapshot struct {
	q     LedgerQuerier
	index common.LedgerIndex
	hash  common.LedgerHash
}

// readEntry reads one entry from the snapshot's ledger. Binding is the point of reading
// through a snapshot rather than through the querier: an unbound snapshot requests its
// index and adopts the hash the server reports, and every later read requests that hash,
// so a ledger closing mid-build cannot mix state from two ledgers into one proof.
func (s *ledgerSnapshot) readEntry(index string) (*ledger.EntryResponse, error) {
	if s == nil || s.index == 0 {
		return nil, fmt.Errorf("%w: validated ledger snapshot is missing", ErrInvalidLedgerState)
	}

	req := &ledger.EntryRequest{Index: index}
	if s.hash == "" {
		req.LedgerIndex = s.index
	} else {
		req.LedgerHash = s.hash
	}
	resp, err := s.q.GetLedgerEntry(req)
	if err != nil {
		return nil, err
	}
	if err := requireEntryNode(resp, index); err != nil {
		return nil, err
	}
	if !resp.Validated || resp.LedgerIndex != s.index || resp.LedgerHash == "" {
		return nil, fmt.Errorf("%w: ledger_entry did not return validated ledger %s", ErrInvalidLedgerState, s.index.Ledger())
	}
	if _, err := hexutil.DecodeFixedHex(string(resp.LedgerHash), 32); err != nil {
		return nil, fmt.Errorf("%w: malformed ledger hash: %w", ErrInvalidLedgerState, err)
	}
	if s.hash == "" {
		s.hash = resp.LedgerHash
	} else if !strings.EqualFold(string(resp.LedgerHash), string(s.hash)) {
		return nil, fmt.Errorf("%w: ledger_entry returned hash %q for selected hash %q", ErrInvalidLedgerState, resp.LedgerHash, s.hash)
	}
	return resp, nil
}

// readCurrentEntry reads one entry from the open ledger, deliberately outside the snapshot.
// The snapshot's contract is that every proof input comes from one validated ledger, and a
// caller reaches for this only to answer a question that contract cannot: whether a
// transaction still in flight has already superseded the state the proof would bind.
func (s *ledgerSnapshot) readCurrentEntry(index string) (*ledger.EntryResponse, error) {
	resp, err := s.q.GetLedgerEntry(&ledger.EntryRequest{Index: index, LedgerIndex: common.Current})
	if err != nil {
		return nil, err
	}
	if err := requireEntryNode(resp, index); err != nil {
		return nil, err
	}
	return resp, nil
}

// requireEntryNode rejects a response that carries no usable JSON node, including one that
// answered in binary instead, and one that answered about an entry other than the one
// requested. Both reads verify this. A response the builder did not ask for is as unusable
// on the open ledger as on the validated one, and readCurrentEntry compares the version it
// finds against the snapshot's, so reading the wrong entry there would decide staleness
// from state that belongs to someone else.
func requireEntryNode(resp *ledger.EntryResponse, index string) error {
	if resp == nil || len(resp.Node) == 0 || resp.NodeBinary != "" {
		return fmt.Errorf("%w: ledger_entry did not return one non-empty JSON node", ErrInvalidLedgerState)
	}
	if !strings.EqualFold(resp.Index, index) {
		return fmt.Errorf("%w: ledger_entry returned index %q for requested index %q", ErrInvalidLedgerState, resp.Index, index)
	}
	return nil
}

// beginBuild opens a build: it fetches the account sequence and selects the validated ledger every later
// state read of the same build is bound to. The two come from different ledgers on purpose:
// the sequence must match the ledger the transaction is applied in, so it is read from the
// open ledger and counts anything the account already submitted, while the state the proofs
// consume is read from a validated ledger so a ledger closing mid-build cannot mix two
// ledgers into one proof.
func beginBuild(q LedgerQuerier, addr string) (uint32, *ledgerSnapshot, error) {
	decoded, err := decodeBuilderAddress(addr)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	classic := types.Address(decoded.Classic)

	validated, err := getAccountInfo(q, classic, common.Validated)
	if err != nil {
		return 0, nil, err
	}
	if !validated.Validated || validated.LedgerIndex == 0 {
		return 0, nil, fmt.Errorf("%w: account_info did not identify a validated ledger", ErrInvalidLedgerState)
	}

	current, err := getAccountInfo(q, classic, common.Current)
	if err != nil {
		return 0, nil, err
	}
	// An account that exists always has a sequence of at least 1, so a zero means account_info
	// answered about something the builder cannot sign for.
	if current.AccountData.Sequence == 0 {
		return 0, nil, fmt.Errorf("%w: account_info reported sequence 0", ErrInvalidLedgerState)
	}
	return current.AccountData.Sequence, &ledgerSnapshot{q: q, index: validated.LedgerIndex}, nil
}

// getAccountInfo reads one account_info response for the given ledger.
func getAccountInfo(q LedgerQuerier, addr types.Address, ledgerIndex common.LedgerSpecifier) (*account.InfoResponse, error) {
	resp, err := q.GetAccountInfo(&account.InfoRequest{Account: addr, LedgerIndex: ledgerIndex})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLedgerQuery, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("%w: account_info returned no result for ledger %s", ErrInvalidLedgerState, ledgerIndex.Ledger())
	}
	return resp, nil
}

// validateLedgerEntryType rejects an entry of the wrong type. An index collision or a
// substituted response would otherwise be read as the object the builder asked for.
func validateLedgerEntryType(resp *ledger.EntryResponse, expected ledgerentries.EntryType) error {
	entryType, err := requiredString(resp.Node, "LedgerEntryType")
	if err != nil {
		return err
	}
	if entryType != string(expected) {
		return fmt.Errorf("%w: expected LedgerEntryType %q, got %q", ErrInvalidLedgerState, expected, entryType)
	}
	return nil
}

// getMPTokenEntry fetches a holder's MPToken entry, mapping a missing entry to
// ErrMPTokenNotFound.
func getMPTokenEntry(snapshot *ledgerSnapshot, issuanceID, holder string) (*ledger.EntryResponse, error) {
	index, err := mpTokenIndex(issuanceID, holder)
	if err != nil {
		return nil, err
	}

	resp, err := snapshot.readEntry(index)
	if err != nil {
		return nil, classifyEntryError(err, ErrMPTokenNotFound)
	}
	if err := validateLedgerEntryType(resp, ledgerentries.MPTokenEntry); err != nil {
		return nil, err
	}
	return resp, nil
}

// classifyEntryError maps a snapshot read failure onto the builder's error surface. The
// three cases are ordered: a server that reports the entry is missing names the entry the
// caller asked for, a snapshot the read itself found inconsistent passes through as it is
// rather than being relabelled a transport failure, and anything left is one.
func classifyEntryError(err, notFound error) error {
	if isEntryNotFound(err) {
		return fmt.Errorf("%w: %w", notFound, err)
	}
	if errors.Is(err, ErrInvalidLedgerState) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrLedgerQuery, err)
}

// mpTokenIndex computes the ledger entry index for an MPToken.
func mpTokenIndex(issuanceID, holder string) (string, error) {
	// The address reaching here is an Account, a Destination, or a Holder depending on the
	// caller, so the error names no field. Each builder validates its own address fields
	// first, which is where a malformed one is reported against the field it came from.
	decoded, err := decodeBuilderAddress(holder)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidAddress, err)
	}
	index, err := xrplhash.MPToken(issuanceID, decoded.Classic)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidIssuanceID, err)
	}
	return index, nil
}

// isEntryNotFound reports whether err, or anything it wraps, is rippled's missing ledger
// entry result. rippled returns "entryNotFound" as the whole error string, and both shipped
// clients surface it verbatim, so an exact match distinguishes that result from a transport
// failure without treating an unrelated error that merely mentions it as a match. The tree
// walk covers the wrapping a custom LedgerQuerier is allowed to add.
func isEntryNotFound(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == ledgerEntryNotFound {
		return true
	}
	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		return slices.ContainsFunc(multi.Unwrap(), isEntryNotFound)
	}
	return isEntryNotFound(errors.Unwrap(err))
}
