package builder

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// BuildMergeInboxParams holds minimal inputs for BuildMergeInbox.
// Sequence is auto-resolved from the ledger.
type BuildMergeInboxParams struct {
	Account    string
	IssuanceID string
}

// MergeInboxParams holds inputs for PrepareMergeInbox.
type MergeInboxParams struct {
	BuildMergeInboxParams
	Sequence uint32
}

// BuildMergeInbox queries ledger state and builds a ConfidentialMPTMergeInbox transaction.
func BuildMergeInbox(q LedgerQuerier, p BuildMergeInboxParams) (*transaction.ConfidentialMPTMergeInbox, error) {
	if err := validateMergeInboxBase(p); err != nil {
		return nil, err
	}

	seq, snapshot, err := beginBuild(q, p.Account)
	if err != nil {
		return nil, err
	}

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

	return PrepareMergeInbox(MergeInboxParams{
		BuildMergeInboxParams: p,
		Sequence:              seq,
	})
}

// PrepareMergeInbox builds a ConfidentialMPTMergeInbox transaction.
// No cryptographic operations are needed. The holder authorizes this inbox-to-spending balance merge.
// Unlike the proof-bearing helpers, no proof binds the sequence here, so a zero Sequence is
// accepted and may be left to a later autofill.
func PrepareMergeInbox(p MergeInboxParams) (*transaction.ConfidentialMPTMergeInbox, error) {
	if err := validateMergeInboxBase(p.BuildMergeInboxParams); err != nil {
		return nil, err
	}

	tx := &transaction.ConfidentialMPTMergeInbox{
		BaseTx: transaction.BaseTx{
			Account:         types.Address(p.Account),
			TransactionType: transaction.ConfidentialMPTMergeInboxTx,
			Sequence:        p.Sequence,
		},
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
