package builder

import (
	"errors"
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
	HolderPrivKey string // 64 hex chars
	HolderPubKey  string // 66 hex chars (compressed)
}

// ConvertParams holds inputs for PrepareConvert.
type ConvertParams struct {
	BuildConvertParams
	IssuerPubKey  string // 66 hex chars (from MPTokenIssuance.IssuerEncryptionKey)
	AuditorPubKey string // 66 hex chars, empty if no auditor
	Sequence      uint32 // Account sequence number
	FirstTime     bool   // If true, registers key + generates Schnorr proof
}

// validateConvertBase validates common Convert fields.
// Note: Amount == 0 is valid per XLS-96 Section 7 (zero-amount convert is the opt-in mechanism
// for registering a holder's encryption key without converting any tokens).
func validateConvertBase(p BuildConvertParams) error {
	if p.Account == "" {
		return ErrMissingAccount
	}
	if p.IssuanceID == "" {
		return ErrMissingIssuanceID
	}
	if p.HolderPrivKey == "" || p.HolderPubKey == "" {
		return ErrMissingHolderKey
	}
	return nil
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

	issuerKey, auditorKey, err := getIssuanceKeys(q, p.IssuanceID)
	if err != nil {
		return nil, err
	}

	// Determine if this is a first-time convert by checking if the MPToken exists
	// and has a HolderEncryptionKey registered.
	firstTime := true
	holderKey, _, _, err := getMPTokenState(q, p.IssuanceID, p.Account)
	if err == nil && holderKey != "" {
		firstTime = false
	} else if err != nil && !errors.Is(err, ErrMPTokenNotFound) {
		return nil, err
	}

	return PrepareConvert(ConvertParams{
		BuildConvertParams: p,
		IssuerPubKey:       issuerKey,
		AuditorPubKey:      auditorKey,
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
