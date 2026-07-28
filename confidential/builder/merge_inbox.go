package builder

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
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

func validateMergeInboxBase(p BuildMergeInboxParams) error {
	if p.Account == "" {
		return ErrMissingAccount
	}
	if !addresscodec.IsValidAddress(p.Account) {
		return ErrInvalidAccount
	}
	if p.IssuanceID == "" {
		return ErrMissingIssuanceID
	}
	if !transaction.IsMPTIssuanceID(p.IssuanceID) {
		return ErrInvalidIssuanceID
	}
	return nil
}

// BuildMergeInbox queries ledger state and builds a ConfidentialMPTMergeInbox transaction.
func BuildMergeInbox(q LedgerQuerier, p BuildMergeInboxParams) (*transaction.ConfidentialMPTMergeInbox, error) {
	if err := validateMergeInboxBase(p); err != nil {
		return nil, err
	}

	seq, err := getSequence(q, p.Account)
	if err != nil {
		return nil, err
	}

	return PrepareMergeInbox(MergeInboxParams{
		BuildMergeInboxParams: p,
		Sequence:              seq,
	})
}

// PrepareMergeInbox builds a ConfidentialMPTMergeInbox transaction.
// No cryptographic operations needed; this is a permissionless inbox-to-spending balance merge.
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

	return tx, nil
}
