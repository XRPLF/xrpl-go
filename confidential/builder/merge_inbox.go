package builder

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// BuildMergeInboxParams holds minimal inputs for BuildMergeInbox.
// The sequence in TxOptions is auto-resolved from the ledger when the caller sets no nonce.
type BuildMergeInboxParams struct {
	TxOptions
	Account    string
	IssuanceID string
}

// MergeInboxParams holds inputs for PrepareMergeInbox. A merge needs nothing the build does not,
// so this adds no field of its own. It exists so every Prepare* helper takes the params type
// named after it, and so a field this operation later needs stays additive.
type MergeInboxParams struct {
	BuildMergeInboxParams
}

// BuildMergeInbox queries ledger state and builds a ConfidentialMPTMergeInbox transaction.
func BuildMergeInbox(q LedgerQuerier, p BuildMergeInboxParams) (*transaction.ConfidentialMPTMergeInbox, error) {
	if err := validateMergeInboxBase(p); err != nil {
		return nil, err
	}

	resolved, snapshot, err := resolveTxOptions(q, p.Account, p.TxOptions, transaction.ConfidentialMPTMergeInboxTx)
	if err != nil {
		return nil, err
	}
	p.TxOptions = resolved

	// The merge carries no proof and encrypts nothing to the issuer, so the ledger reads
	// exist purely to preflight a missing or non-confidential issuance,
	// and an MPToken that was never initialized for confidential balances.
	issuance, err := readIssuance(snapshot, p.IssuanceID)
	if err != nil {
		return nil, err
	}
	if err := getMPTokenMergeState(snapshot, issuance, p.IssuanceID, p.Account); err != nil {
		return nil, err
	}

	return PrepareMergeInbox(MergeInboxParams{BuildMergeInboxParams: p})
}

// PrepareMergeInbox builds a ConfidentialMPTMergeInbox transaction.
// No cryptographic operations are needed. The holder authorizes this inbox-to-spending balance merge.
// Unlike the proof-bearing helpers, no proof binds the sequence here, so a zero nonce is
// accepted and may be left to a later autofill.
func PrepareMergeInbox(p MergeInboxParams) (*transaction.ConfidentialMPTMergeInbox, error) {
	if err := validateMergeInboxBase(p.BuildMergeInboxParams); err != nil {
		return nil, err
	}
	if err := p.validate(p.Account, transaction.ConfidentialMPTMergeInboxTx); err != nil {
		return nil, err
	}

	tx := &transaction.ConfidentialMPTMergeInbox{
		BaseTx:            baseTx(p.Account, transaction.ConfidentialMPTMergeInboxTx, p.TxOptions),
		MPTokenIssuanceID: p.IssuanceID,
	}

	if err := validatePreparedTransaction(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func validateMergeInboxBase(p BuildMergeInboxParams) error {
	if p.Account == "" {
		return ErrMissingAccount
	}
	if _, err := decodeBuilderAddress(p.Account); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	if p.IssuanceID == "" {
		return ErrMissingIssuanceID
	}
	if err := validateHolderRole(p.IssuanceID, p.Account); err != nil {
		return err
	}
	return nil
}
