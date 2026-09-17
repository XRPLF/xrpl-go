package builder

import (
	"errors"
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/flag"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
)

// issuanceState holds the MPTokenIssuance fields the builders preflight against.
type issuanceState struct {
	// hasDomain records an issuance that authorizes through a permissioned domain. The
	// credentials that path accepts live outside the issuance, so a holder can be authorized
	// without lsfMPTAuthorized and the flag-only auth check must stand down.
	hasDomain  bool
	issuerKey  string
	auditorKey string
	flags      uint32
	// transferFee is rejected outright for a confidential send. MPTokenIssuanceCreate and
	// MPTokenIssuanceSet already forbid combining it with confidential balances, so a
	// non-zero value here means the issuance predates that rule or was built by hand.
	transferFee uint32
	// confidentialOutstanding is ConfidentialOutstandingAmount, the confidential supply.
	// It bounds every holder's confidential balance, so a convert-back cannot exceed it and
	// a clawback never searches for a balance past it.
	confidentialOutstanding uint64
}

// canTransfer reports whether non-issuers may move this issuance between accounts.
func (s issuanceState) canTransfer() bool {
	return flag.Contains(s.flags, ledgerentries.LsfMPTCanTransfer)
}

// canClawback reports whether the issuer may claw back value from individual holders.
func (s issuanceState) canClawback() bool {
	return flag.Contains(s.flags, ledgerentries.LsfMPTCanClawback)
}

// isLocked reports whether the issuer locked every balance of the issuance.
func (s issuanceState) isLocked() bool {
	return flag.Contains(s.flags, ledgerentries.LsfMPTLocked)
}

// requiresAuth reports whether holders must be authorized before they can hold the issuance.
func (s issuanceState) requiresAuth() bool {
	return flag.Contains(s.flags, ledgerentries.LsfMPTRequireAuth)
}

// hasAuditor reports whether the issuance registered an auditor key. The transactors treat
// an issuance-level auditor key as implying an AuditorEncryptedBalance on every MPToken of
// that issuance, so holders carry the field only when it is set.
func (s issuanceState) hasAuditor() bool {
	return s.auditorKey != ""
}

// readIssuance fetches the MPTokenIssuance fields every confidential builder preflights and
// rejects an issuance that cannot hold confidential balances, which the transactors reject
// with tecNO_PERMISSION. The issuer key is reported as found rather than required, so a
// builder that encrypts nothing to it can accept an issuance that has none.
func readIssuance(snapshot *ledgerSnapshot, issuanceID string) (issuanceState, error) {
	var state issuanceState

	index, err := xrplhash.MPTokenIssuance(issuanceID)
	if err != nil {
		return state, fmt.Errorf("%w: %w", ErrInvalidIssuanceID, err)
	}

	resp, err := snapshot.readEntry(index)
	if err != nil {
		return state, classifyEntryError(err, ErrIssuanceNotFound)
	}
	if err := validateLedgerEntryType(resp, ledgerentries.MPTokenIssuanceEntry); err != nil {
		return state, err
	}

	if state.flags, err = optionalUint32(resp.Node, "Flags"); err != nil {
		return issuanceState{}, err
	}
	if !flag.Contains(state.flags, ledgerentries.LsfMPTCanHoldConfidentialBalance) {
		return issuanceState{}, ErrConfidentialDisabled
	}
	if state.transferFee, err = optionalUint32(resp.Node, "TransferFee"); err != nil {
		return issuanceState{}, err
	}
	if state.confidentialOutstanding, err = optionalUint64(resp.Node, "ConfidentialOutstandingAmount"); err != nil {
		return issuanceState{}, err
	}
	if state.issuerKey, err = optionalString(resp.Node, "IssuerEncryptionKey"); err != nil {
		return issuanceState{}, err
	}
	if state.auditorKey, err = optionalString(resp.Node, "AuditorEncryptionKey"); err != nil {
		return issuanceState{}, err
	}
	_, state.hasDomain = resp.Node["DomainID"]
	return state, nil
}

// getProvableIssuance reads the issuance state a proving builder needs. Every confidential
// proof except the inbox merge encrypts an amount to the issuer key, so an issuance without
// one cannot be built against.
func getProvableIssuance(snapshot *ledgerSnapshot, issuanceID string) (issuanceState, error) {
	state, err := readIssuance(snapshot, issuanceID)
	if err != nil {
		return issuanceState{}, err
	}
	if state.issuerKey == "" {
		return issuanceState{}, ErrEncryptionKeyNotSet
	}
	return state, nil
}

// mptokenState holds the MPToken fields a confidential spend preflights against.
type mptokenState struct {
	holderKey string
	// balanceCt is ConfidentialBalanceSpending, the ciphertext the spend debits.
	balanceCt      string
	balanceVersion uint32
}

// getMPTokenState fetches the confidential spending state a holder must already have.
// ConfidentialMPTSend and ConfidentialMPTConvertBack both reject a holder missing the
// holder key or the spending balance with tecNO_PERMISSION. Returns ErrMPTokenNotFound if
// the entry does not exist.
func getMPTokenState(snapshot *ledgerSnapshot, issuance issuanceState, issuanceID, holder string) (mptokenState, error) {
	resp, err := readUsableMPToken(snapshot, issuance, issuanceID, holder)
	if err != nil {
		return mptokenState{}, err
	}

	var state mptokenState
	if state.holderKey, err = requireSpendField(resp.Node, "HolderEncryptionKey"); err != nil {
		return mptokenState{}, err
	}
	if state.balanceCt, err = requireSpendField(resp.Node, "ConfidentialBalanceSpending"); err != nil {
		return mptokenState{}, err
	}
	// The transactor updates each mirror balance homomorphically alongside the holder's own.
	// Both ConfidentialMPTSend and ConfidentialMPTConvertBack reject a missing
	// IssuerEncryptedBalance with tecNO_PERMISSION, a condition XLS-96 10.4.2 omits for the
	// convert-back path. A missing AuditorEncryptedBalance under an auditing issuance is an
	// invariant the transactors assert rather than a failure they expect, so requiring it
	// here reports a malformed ledger as a builder error instead of letting the transaction
	// pay for a tefINTERNAL.
	for _, field := range mirrorBalanceFields(issuance) {
		if _, err := requireSpendField(resp.Node, field); err != nil {
			return mptokenState{}, err
		}
	}
	// XLS-96 7.5.5 starts the version counter at 0 and 9.3 wraps it back to 0, so an absent
	// field is version 0 and only a present but unreadable value is malformed.
	if state.balanceVersion, err = optionalUint32(resp.Node, "ConfidentialBalanceVersion"); err != nil {
		return mptokenState{}, err
	}
	if err := requireCurrentBalanceVersion(snapshot, resp.Index, state.balanceVersion); err != nil {
		return mptokenState{}, err
	}
	return state, nil
}

// requireCurrentBalanceVersion rejects a build whose proof is already superseded. Both proofs
// that consume a balance version bind it into their context hash, and the transactor rebuilds
// that hash from the version it reads when the transaction applies, so a confidential
// transaction of the holder's own still in flight bumps the version and the proof is rejected
// with tecBAD_PROOF. The validated snapshot cannot see that transaction, which is the point of
// reading the open ledger here: the version is the one field whose staleness is knowable before
// paying for it. Only a version the open ledger reports differently is rejected, so an
// unrelated in-flight transaction that moves the account sequence alone still builds.
func requireCurrentBalanceVersion(snapshot *ledgerSnapshot, index string, validated uint32) error {
	resp, err := snapshot.readCurrentEntry(index)
	if err != nil {
		return classifyEntryError(err, ErrStaleBalanceVersion)
	}
	if err := validateLedgerEntryType(resp, ledgerentries.MPTokenEntry); err != nil {
		return err
	}

	// The comparison only means something if the open ledger actually leads the snapshot.
	// rippled reports an open ledger as ledger_current_index and never as ledger_index, and a
	// pooled or load-balanced endpoint can answer this read from a server whose open ledger
	// trails the validated ledger the snapshot pinned. Rejecting on that would fail a build
	// the network would have accepted, so a read that does not demonstrably lead is skipped
	// rather than trusted: the staleness check saves a fee, and going without it costs at
	// worst the tecBAD_PROOF it exists to avoid.
	if resp.LedgerCurrentIndex <= snapshot.index {
		return nil
	}

	current, err := optionalUint32(resp.Node, "ConfidentialBalanceVersion")
	if err != nil {
		return err
	}
	// The version wraps at 32 bits (XLS-96 9.3), so any difference is a change and no
	// ordering may be read into it.
	if current != validated {
		return fmt.Errorf("%w: validated ledger reports version %d, open ledger reports %d",
			ErrStaleBalanceVersion, validated, current)
	}
	return nil
}

// mirrorBalanceFields lists the mirror balances the transactor keeps on a holder's MPToken
// alongside the holder's own. The auditor mirror exists only once the issuance registers an
// auditor key, which XLS-96 8.4 treats as implying the field on every holder.
func mirrorBalanceFields(issuance issuanceState) []string {
	if issuance.hasAuditor() {
		return []string{"IssuerEncryptedBalance", "AuditorEncryptedBalance"}
	}
	return []string{"IssuerEncryptedBalance"}
}

// readUsableMPToken reads a holder's MPToken and rejects a holder the transactor would reject
// before it looks at any confidential state. Every confidential MPT reader goes through this
// except getIssuerCiphertext, which reads the entry directly because the clawback deliberately
// runs neither check.
func readUsableMPToken(snapshot *ledgerSnapshot, issuance issuanceState, issuanceID, holder string) (*ledger.EntryResponse, error) {
	resp, err := getMPTokenEntry(snapshot, issuanceID, holder)
	if err != nil {
		return nil, err
	}
	if err := requireHolderUsable(resp.Node, issuance); err != nil {
		return nil, err
	}
	return resp, nil
}

// requireHolderUsable rejects a holder the transactor would reject before it looks at any
// confidential state. Every confidential MPT transactor except the clawback ends its preclaim
// with checkFrozen and requireAuth, so a locked or unauthorized holder costs a fee for a
// tecLOCKED or a tecNO_AUTH that the entries already in hand decide for free. The clawback
// runs neither check on purpose, because an issuer must be able to claw back from a holder it
// has locked.
//
// Two conditions rippled also rejects are deliberately not decided here, because both need
// ledger entries no builder reads and rejecting on a guess would deny a valid build: a holder
// that is a vault pseudo-account, and the permissioned-domain credentials an issuance with a
// DomainID accepts in place of lsfMPTAuthorized.
func requireHolderUsable(node map[string]any, issuance issuanceState) error {
	// An MPToken with no flag set omits the field entirely.
	flags, err := optionalUint32(node, "Flags")
	if err != nil {
		return err
	}
	if issuance.isLocked() {
		return ErrIssuanceLocked
	}
	if flag.Contains(flags, ledgerentries.LsfMPTLocked) {
		return ErrHolderLocked
	}
	if issuance.requiresAuth() && !issuance.hasDomain && !flag.Contains(flags, ledgerentries.LsfMPTAuthorized) {
		return ErrHolderNotAuthorized
	}
	return nil
}

// nameParty labels the holder conditions that belong to one participant, so a send between
// two accounts reports which side blocked it. Both parties run through requireHolderUsable
// and would otherwise return the same sentinel. ErrIssuanceLocked is deliberately left
// unlabelled, because a locked issuance is a property of the issuance and belongs to
// neither participant.
func nameParty(err error, party string) error {
	if errors.Is(err, ErrHolderLocked) || errors.Is(err, ErrHolderNotAuthorized) {
		return fmt.Errorf("%s: %w", party, err)
	}
	return err
}

// requireSpendField reads a field no confidential spend can proceed without.
func requireSpendField(node map[string]any, field string) (string, error) {
	value, err := requiredString(node, field)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrMissingSenderState, err)
	}
	return value, nil
}

// getMPTokenReceiverState checks that a destination is initialized to receive a
// confidential send and returns the key the transferred amount is encrypted under.
// XLS-96 8.4 credits the destination's inbox and updates its mirror balances, so each one
// must already exist.
func getMPTokenReceiverState(snapshot *ledgerSnapshot, issuance issuanceState, issuanceID, destination string) (string, error) {
	resp, err := readUsableMPToken(snapshot, issuance, issuanceID, destination)
	if err != nil {
		return "", err
	}

	holderKey, err := requiredString(resp.Node, "HolderEncryptionKey")
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrReceiverNotOptedIn, err)
	}
	for _, field := range append([]string{"ConfidentialBalanceInbox"}, mirrorBalanceFields(issuance)...) {
		if _, err := requiredString(resp.Node, field); err != nil {
			return "", fmt.Errorf("%w: %w", ErrReceiverNotOptedIn, err)
		}
	}
	return holderKey, nil
}

// getMPTokenMergeState checks that a holder's MPToken is initialized for an inbox merge.
// ConfidentialMPTMergeInbox rejects with tecNO_PERMISSION a holder missing either confidential
// balance or the holder encryption key, and requires nothing else of the MPToken. XLS-96
// 9.2.1.2 lists only the two balances, so the key is taken from the transactor. The merge
// moves a balance the ledger already knows, so no field value is read here.
func getMPTokenMergeState(snapshot *ledgerSnapshot, issuance issuanceState, issuanceID, holder string) error {
	resp, err := readUsableMPToken(snapshot, issuance, issuanceID, holder)
	if err != nil {
		return err
	}
	for _, field := range []string{"ConfidentialBalanceInbox", "ConfidentialBalanceSpending", "HolderEncryptionKey"} {
		if _, err := requireSpendField(resp.Node, field); err != nil {
			return err
		}
	}
	return nil
}

// convertState holds the MPToken fields a confidential convert preflights against.
type convertState struct {
	// holderKey is optional. An absent key means an existing MPToken has not registered for
	// confidential transfers, which is what selects the first-time proof form.
	holderKey string
	// publicAmount is MPTAmount, the public balance the convert debits. An MPToken holding
	// nothing omits the field, so an absent value is zero.
	publicAmount uint64
}

// getMPTokenConvertState fetches the state ConfidentialMPTConvert preflights against. The
// balance version is deliberately not read, so a convert never fails on a version it does
// not consume.
func getMPTokenConvertState(snapshot *ledgerSnapshot, issuance issuanceState, issuanceID, holder string) (convertState, error) {
	resp, err := readUsableMPToken(snapshot, issuance, issuanceID, holder)
	if err != nil {
		return convertState{}, err
	}

	var state convertState
	if state.holderKey, err = optionalString(resp.Node, "HolderEncryptionKey"); err != nil {
		return convertState{}, err
	}
	if state.publicAmount, err = optionalUint64(resp.Node, "MPTAmount"); err != nil {
		return convertState{}, err
	}
	return state, nil
}

// getIssuerCiphertext fetches the IssuerEncryptedBalance the clawback equality proof consumes.
// XLS-96 11.3.2 rejects a holder without it, and ConfidentialMPTClawback rejects a holder
// missing the holder encryption key alongside it. The clawback is the one confidential
// transactor that runs neither checkFrozen nor requireAuth, because an issuer must be able to
// claw back from a holder it has locked, so this reads the entry directly rather than through
// readUsableMPToken.
func getIssuerCiphertext(snapshot *ledgerSnapshot, issuanceID, holder string) (string, error) {
	resp, err := getMPTokenEntry(snapshot, issuanceID, holder)
	if err != nil {
		return "", err
	}
	// Both fields are ordinary holder state rather than a malformed response: a holder that
	// never opted into confidential balances simply has neither. requireSpendField reports
	// that as ErrMissingSenderState, the same sentinel every other reader uses for it.
	if _, err := requireSpendField(resp.Node, "HolderEncryptionKey"); err != nil {
		return "", err
	}
	return requireSpendField(resp.Node, "IssuerEncryptedBalance")
}
