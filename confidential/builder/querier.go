package builder

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const ledgerEntryNotFound = "entryNotFound"

// LedgerQuerier is the minimal interface needed to query ledger state.
// Both rpc.Client and websocket.Client satisfy this interface.
type LedgerQuerier interface {
	GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error)
	GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error)
}

// issuanceState holds the MPTokenIssuance fields the builders preflight against.
type issuanceState struct {
	issuerKey  string
	auditorKey string
	flags      uint32
	// transferFee is rejected outright for a confidential send. MPTokenIssuanceCreate and
	// MPTokenIssuanceSet already forbid combining it with confidential balances, so a
	// non-zero value here means the issuance predates that rule or was built by hand.
	transferFee uint32
	// confidentialOutstanding is ConfidentialOutstandingAmount, the confidential supply.
	// It bounds every holder's confidential balance, so a clawback cannot exceed it.
	confidentialOutstanding uint64
}

// canTransfer reports whether non-issuers may move this issuance between accounts.
func (s issuanceState) canTransfer() bool {
	return s.flags&ledgerentries.LsfMPTCanTransfer != 0
}

// getSequence fetches the account sequence number.
func getSequence(q LedgerQuerier, addr string) (uint32, error) {
	resp, err := q.GetAccountInfo(&account.InfoRequest{
		Account: types.Address(addr),
	})
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrLedgerQuery, err)
	}
	return resp.AccountData.Sequence, nil
}

// getIssuance fetches the MPTokenIssuance fields every confidential builder needs and
// rejects an issuance that cannot hold confidential balances, which the transactors
// reject with tecNO_PERMISSION.
func getIssuance(q LedgerQuerier, issuanceID string) (issuanceState, error) {
	var state issuanceState

	index, err := xrplhash.MPTokenIssuance(issuanceID)
	if err != nil {
		return state, fmt.Errorf("%w: %w", ErrInvalidIssuanceID, err)
	}

	resp, err := q.GetLedgerEntry(&ledger.EntryRequest{Index: index})
	if err != nil {
		if hasExactErrorMessage(err, ledgerEntryNotFound) {
			return state, fmt.Errorf("%w: %w", ErrIssuanceNotFound, err)
		}
		return state, fmt.Errorf("%w: %w", ErrLedgerQuery, err)
	}

	if value, present := resp.Node["Flags"]; present {
		state.flags, err = parseLedgerUint32("Flags", value)
		if err != nil {
			return issuanceState{}, err
		}
	}
	if state.flags&ledgerentries.LsfMPTCanHoldConfidentialBalance == 0 {
		return issuanceState{}, ErrConfidentialDisabled
	}

	if value, present := resp.Node["TransferFee"]; present {
		state.transferFee, err = parseLedgerUint32("TransferFee", value)
		if err != nil {
			return issuanceState{}, err
		}
	}
	if value, present := resp.Node["ConfidentialOutstandingAmount"]; present {
		state.confidentialOutstanding, err = parseLedgerUint64("ConfidentialOutstandingAmount", value)
		if err != nil {
			return issuanceState{}, err
		}
	}

	if v, ok := resp.Node["IssuerEncryptionKey"].(string); ok {
		state.issuerKey = v
	}
	if state.issuerKey == "" {
		return issuanceState{}, ErrEncryptionKeyNotSet
	}
	if v, ok := resp.Node["AuditorEncryptionKey"].(string); ok {
		state.auditorKey = v
	}
	return state, nil
}

// getMPTokenState fetches MPToken fields for a holder.
// Returns holderKey, balanceCt, balanceVersion. Returns ErrMPTokenNotFound if the entry does not exist.
func getMPTokenState(q LedgerQuerier, issuanceID, holder string) (holderKey, balanceCt string, balanceVersion uint32, err error) {
	node, err := readMPToken(q, issuanceID, holder)
	if err != nil {
		return "", "", 0, err
	}

	if v, ok := node["HolderEncryptionKey"].(string); ok {
		holderKey = v
	}
	if v, ok := node["ConfidentialBalanceSpending"].(string); ok {
		balanceCt = v
	}
	if value, present := node["ConfidentialBalanceVersion"]; present {
		balanceVersion, err = parseLedgerUint32("ConfidentialBalanceVersion", value)
		if err != nil {
			return "", "", 0, err
		}
	}
	return holderKey, balanceCt, balanceVersion, nil
}

// getMPTokenHolderKey fetches only HolderEncryptionKey. Callers that do not consume the
// balance version use this so an unreadable version cannot fail a build that never needed it.
func getMPTokenHolderKey(q LedgerQuerier, issuanceID, holder string) (string, error) {
	node, err := readMPToken(q, issuanceID, holder)
	if err != nil {
		return "", err
	}
	holderKey, _ := node["HolderEncryptionKey"].(string)
	return holderKey, nil
}

// getIssuerCiphertext fetches the IssuerEncryptedBalance from a holder's MPToken.
func getIssuerCiphertext(q LedgerQuerier, issuanceID, holder string) (string, error) {
	node, err := readMPToken(q, issuanceID, holder)
	if err != nil {
		return "", err
	}

	ct, ok := node["IssuerEncryptedBalance"].(string)
	if !ok || ct == "" {
		return "", fmt.Errorf("%w: IssuerEncryptedBalance not found", ErrLedgerQuery)
	}
	return ct, nil
}

// readMPToken fetches a holder's MPToken node, mapping a missing entry to ErrMPTokenNotFound.
func readMPToken(q LedgerQuerier, issuanceID, holder string) (map[string]any, error) {
	index, err := mpTokenIndex(issuanceID, holder)
	if err != nil {
		return nil, err
	}

	resp, err := q.GetLedgerEntry(&ledger.EntryRequest{Index: index})
	if err != nil {
		if hasExactErrorMessage(err, ledgerEntryNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrMPTokenNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", ErrLedgerQuery, err)
	}
	return resp.Node, nil
}

// mpTokenIndex computes the ledger entry index for an MPToken.
func mpTokenIndex(issuanceID, holder string) (string, error) {
	index, err := xrplhash.MPToken(issuanceID, holder)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidIssuanceID, err)
	}
	return index, nil
}

// hasExactErrorMessage reports whether err, or anything it wraps, carries exactly message.
// rippled returns "entryNotFound" as the whole error string for a missing ledger entry, and
// both shipped clients surface it verbatim, so an exact match distinguishes that result from
// a transport failure without treating an unrelated error that merely mentions it as a match.
func hasExactErrorMessage(err error, message string) bool {
	if err == nil {
		return false
	}
	if err.Error() == message {
		return true
	}
	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		for _, cause := range multi.Unwrap() {
			if hasExactErrorMessage(cause, message) {
				return true
			}
		}
		return false
	}
	if single, ok := err.(interface{ Unwrap() error }); ok {
		return hasExactErrorMessage(single.Unwrap(), message)
	}
	return false
}

// parseLedgerUint32 reads a numeric ledger field. rpc decodes with UseNumber and websocket
// decodes into any, so a JSON number arrives as json.Number or float64 and never as an integer.
func parseLedgerUint32(field string, value any) (uint32, error) {
	switch number := value.(type) {
	case json.Number:
		parsed, err := strconv.ParseUint(number.String(), 10, 32)
		if err == nil {
			return uint32(parsed), nil
		}
	case float64:
		if number >= 0 && number <= float64(^uint32(0)) && math.Trunc(number) == number {
			return uint32(number), nil
		}
	}

	return 0, fmt.Errorf("%w: invalid %s %v", ErrLedgerQuery, field, value)
}

// parseLedgerUint64 reads a ledger amount field. MPT amounts are serialized as decimal
// strings, so a float64 cannot represent them exactly and is not accepted.
func parseLedgerUint64(field string, value any) (uint64, error) {
	switch amount := value.(type) {
	case string:
		parsed, err := strconv.ParseUint(amount, 10, 64)
		if err == nil {
			return parsed, nil
		}
	case json.Number:
		parsed, err := strconv.ParseUint(amount.String(), 10, 64)
		if err == nil {
			return parsed, nil
		}
	}

	return 0, fmt.Errorf("%w: invalid %s %v", ErrLedgerQuery, field, value)
}
