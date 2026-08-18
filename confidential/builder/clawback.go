// Package builder provides transaction builders for confidential MPT operations.
//
// Every address these builders accept must be a classic r-address. Unlike the transaction
// models, which the binary codec converts on encode, the builders feed addresses straight
// into xrplhash.MPToken and the proof layer, both of which decode classic addresses only.
// An X-address would surface much later as an unrelated error, so it is rejected up front.
//
// Classic encoding is canonical, so equal accounts always produce equal strings once both
// sides pass that check, which is what lets the self-send and self-clawback guards compare
// addresses with ==.
package builder

import (
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
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
	if !addresscodec.IsValidClassicAddress(p.Account) {
		return ErrInvalidAccount
	}
	if p.Holder == "" {
		return ErrMissingHolder
	}
	if !addresscodec.IsValidClassicAddress(p.Holder) {
		return ErrInvalidHolder
	}
	if p.Account == p.Holder {
		return ErrSelfClawback
	}
	if p.IssuanceID == "" {
		return ErrMissingIssuanceID
	}
	if err := validateIssuerRole(p.IssuanceID, p.Account); err != nil {
		return err
	}
	if err := validateAmount(p.Amount); err != nil {
		return err
	}
	if p.IssuerPrivKey == "" {
		return ErrMissingIssuerKey
	}
	if !isValidPrivKey(p.IssuerPrivKey) {
		return ErrInvalidPrivKey
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
	if !transaction.IsValidCompressedEncryptionKey(p.IssuerPubKey) {
		return nil, ErrInvalidPubKey
	}
	if p.IssuerCiphertext == "" {
		return nil, ErrMissingCiphertext
	}
	if !transaction.IsValidCiphertext(p.IssuerCiphertext) {
		return nil, ErrInvalidCiphertext
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
