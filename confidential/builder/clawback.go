// Package builder provides transaction builders for confidential MPT operations.
package builder

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// BuildClawbackParams holds minimal inputs for BuildClawback.
// Sequence, IssuerPubKey, and IssuerCiphertext are auto-resolved from the ledger.
type BuildClawbackParams struct {
	Account       string // Issuer
	Holder        string
	IssuanceID    string
	Amount        uint64
	IssuerPrivKey string // 64 hex chars
}

// ClawbackParams holds inputs for PrepareClawback.
type ClawbackParams struct {
	BuildClawbackParams
	IssuerPubKey     string // 66 hex chars (from MPTokenIssuance.IssuerEncryptionKey)
	IssuerCiphertext string // 132 hex chars, IssuerEncryptedBalance from holder's MPToken
	Sequence         uint32
}

func validateClawbackBase(p BuildClawbackParams) error {
	if p.Account == "" {
		return ErrMissingAccount
	}
	if p.Holder == "" {
		return ErrMissingHolder
	}
	if p.Account == p.Holder {
		return ErrSelfClawback
	}
	if p.IssuanceID == "" {
		return ErrMissingIssuanceID
	}
	if p.Amount == 0 {
		return ErrZeroAmount
	}
	if p.IssuerPrivKey == "" {
		return ErrMissingIssuerKey
	}
	return nil
}

// BuildClawback queries ledger state and builds a ConfidentialMPTClawback transaction.
func BuildClawback(q LedgerQuerier, p BuildClawbackParams) (*transaction.ConfidentialMPTClawback, error) {
	if err := validateClawbackBase(p); err != nil {
		return nil, err
	}

	seq, err := getSequence(q, p.Account)
	if err != nil {
		return nil, err
	}

	issuerKey, _, err := getIssuanceKeys(q, p.IssuanceID)
	if err != nil {
		return nil, err
	}

	issuerCt, err := getIssuerCiphertext(q, p.IssuanceID, p.Holder)
	if err != nil {
		return nil, err
	}

	return PrepareClawback(ClawbackParams{
		BuildClawbackParams: p,
		IssuerPubKey:        issuerKey,
		IssuerCiphertext:    issuerCt,
		Sequence:            seq,
	})
}

// PrepareClawback builds a ConfidentialMPTClawback transaction.
//
// Steps:
// 1. Compute clawback context hash (issuer, issuance, seq, holder).
// 2. Generate equality proof proving the clawback amount matches the issuer's ciphertext.
func PrepareClawback(p ClawbackParams) (*transaction.ConfidentialMPTClawback, error) {
	if err := validateClawbackBase(p.BuildClawbackParams); err != nil {
		return nil, err
	}
	if p.IssuerPubKey == "" {
		return nil, ErrMissingIssuerKey
	}
	if p.IssuerCiphertext == "" {
		return nil, ErrMissingCiphertext
	}

	ctxHash, err := proof.ClawbackContextHash(p.Account, p.IssuanceID, p.Sequence, p.Holder)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	proofHex, err := proof.GenerateClawbackProof(p.IssuerPrivKey, p.IssuerPubKey, ctxHash, p.Amount, p.IssuerCiphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	tx := &transaction.ConfidentialMPTClawback{
		BaseTx: transaction.BaseTx{
			Account:         types.Address(p.Account),
			TransactionType: transaction.ConfidentialMPTClawbackTx,
			Sequence:        p.Sequence,
		},
		MPTokenIssuanceID: p.IssuanceID,
		Holder:            types.Address(p.Holder),
		MPTAmount:         types.MPTPlainAmount(p.Amount),
		ZKProof:           proofHex,
	}

	return tx, nil
}
