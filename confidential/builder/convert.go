package builder

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// BuildConvertParams holds minimal inputs for BuildConvert.
// Sequence, IssuerPubKey, AuditorPubKey, and FirstTime are auto-resolved from the ledger.
type BuildConvertParams struct {
	Account       string
	IssuanceID    string
	Amount        uint64
	HolderPrivKey string // Non-zero secp256k1 scalar below the curve order. Required only for first-time key registration.
	HolderPubKey  string // Valid 33-byte compressed secp256k1 point
}

// ConvertParams holds inputs for PrepareConvert.
type ConvertParams struct {
	BuildConvertParams
	IssuerPubKey  string // 66 hex chars (from MPTokenIssuance.IssuerEncryptionKey)
	AuditorPubKey string // 66 hex chars, empty if no auditor
	// Sequence is bound into the first-time proof and must not change after preparation.
	// A repeat convert carries no proof, so it may be left zero for a later autofill.
	Sequence  uint32
	FirstTime bool // If true, registers key + generates Schnorr proof
}

// BuildConvert queries ledger state and builds a ConfidentialMPTConvert transaction.
func BuildConvert(q LedgerQuerier, p BuildConvertParams) (*transaction.ConfidentialMPTConvert, error) {
	if err := validateConvertBase(p); err != nil {
		return nil, err
	}

	seq, err := getSequence(q, p.Account)
	if err != nil {
		return nil, err
	}

	issuance, err := getIssuance(q, p.IssuanceID)
	if err != nil {
		return nil, err
	}

	// A holder without a registered encryption key uses the first-time proof form.
	holderKey, err := getMPTokenHolderKey(q, p.IssuanceID, p.Account)
	if err != nil {
		return nil, err
	}
	firstTime := holderKey == ""
	if !firstTime && !sameEncryptionKey(holderKey, p.HolderPubKey) {
		return nil, fmt.Errorf("%w: holder key", ErrKeyMismatch)
	}

	return PrepareConvert(ConvertParams{
		BuildConvertParams: p,
		IssuerPubKey:       issuance.issuerKey,
		AuditorPubKey:      issuance.auditorKey,
		Sequence:           seq,
		FirstTime:          firstTime,
	})
}

// PrepareConvert builds a ConfidentialMPTConvert transaction.
//
// Steps:
// 1. Generate a shared blinding factor.
// 2. Encrypt amount under holder and issuer keys (same BF).
// 3. Optionally encrypt under auditor key.
// 4. If first-time: compute context hash, register holder key, generate Schnorr proof.
func PrepareConvert(p ConvertParams) (*transaction.ConfidentialMPTConvert, error) {
	if err := validateConvertBase(p.BuildConvertParams); err != nil {
		return nil, err
	}
	if p.IssuerPubKey == "" {
		return nil, ErrMissingIssuerKey
	}
	if !transaction.IsValidCompressedEncryptionKey(p.IssuerPubKey) {
		return nil, fmt.Errorf("issuer pub key: %w", ErrInvalidPubKey)
	}
	if p.AuditorPubKey != "" && !transaction.IsValidCompressedEncryptionKey(p.AuditorPubKey) {
		return nil, fmt.Errorf("auditor pub key: %w", ErrInvalidPubKey)
	}
	// Only the first-time form carries a proof, and only that proof binds the sequence.
	// A repeat convert is free to leave Sequence to a later autofill.
	if p.FirstTime {
		if p.HolderPrivKey == "" {
			return nil, ErrMissingHolderKey
		}
		if !isValidPrivateKey(p.HolderPrivKey) {
			return nil, ErrInvalidPrivKey
		}
		if p.Sequence == 0 {
			return nil, ErrMissingSequence
		}
	}

	bf, err := elgamal.GenerateBlindingFactor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	// Encrypt amount under each party's key with the same BF.
	holderCt, err := elgamal.Encrypt(p.Amount, p.HolderPubKey, bf)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}
	issuerCt, err := elgamal.Encrypt(p.Amount, p.IssuerPubKey, bf)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	tx := &transaction.ConfidentialMPTConvert{
		BaseTx: transaction.BaseTx{
			Account:         types.Address(p.Account),
			TransactionType: transaction.ConfidentialMPTConvertTx,
			Sequence:        p.Sequence,
		},
		MPTokenIssuanceID:     p.IssuanceID,
		MPTAmount:             types.MPTPlainAmount(p.Amount),
		HolderEncryptedAmount: holderCt,
		IssuerEncryptedAmount: issuerCt,
		BlindingFactor:        bf,
	}

	if p.AuditorPubKey != "" {
		auditorCt, err := elgamal.Encrypt(p.Amount, p.AuditorPubKey, bf)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
		}
		tx.AuditorEncryptedAmount = &auditorCt
	}

	// First-time key registration: decompress key + generate Schnorr proof.
	if p.FirstTime {
		ctxHash, err := proof.ConvertContextHash(p.Account, p.IssuanceID, p.Sequence)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
		}

		tx.HolderEncryptionKey = &p.HolderPubKey

		proofHex, err := proof.GenerateConvertProof(p.HolderPubKey, p.HolderPrivKey, ctxHash)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
		}
		tx.ZKProof = &proofHex
	}

	return tx, nil
}

// validateConvertBase validates common Convert fields. A zero Amount is valid here.
// See validateAmountUpperBound.
func validateConvertBase(p BuildConvertParams) error {
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
	if err := validateAmountUpperBound(p.Amount); err != nil {
		return err
	}
	if p.HolderPrivKey != "" && !isValidPrivateKey(p.HolderPrivKey) {
		return ErrInvalidPrivKey
	}
	if p.HolderPubKey == "" {
		return ErrMissingHolderKey
	}
	if !transaction.IsValidCompressedEncryptionKey(p.HolderPubKey) {
		return fmt.Errorf("holder pub key: %w", ErrInvalidPubKey)
	}
	return nil
}
