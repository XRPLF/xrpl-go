package builder

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/commitment"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// BuildConvertBackParams holds minimal inputs for BuildConvertBack.
// The sequence in TxOptions, IssuerPubKey, AuditorPubKey, BalanceVersion, CurrentBalanceCt,
// and CurrentBalance are auto-resolved from the ledger. Balance is decrypted using HolderPrivKey
// within BalanceRange's inclusive bounds.
type BuildConvertBackParams struct {
	TxOptions
	Account       string
	IssuanceID    string
	Amount        uint64
	HolderPrivKey string              // Non-zero secp256k1 scalar below the curve order, also used to decrypt balance from ledger
	HolderPubKey  string              // Valid 33-byte compressed secp256k1 point
	BalanceRange  elgamal.AmountRange // Inclusive balance decryption bounds
}

// ConvertBackParams holds inputs for PrepareConvertBack.
type ConvertBackParams struct {
	BuildConvertBackParams
	IssuerPubKey     string // 66 hex chars
	AuditorPubKey    string // 66 hex chars, empty if no auditor
	BalanceVersion   uint32
	CurrentBalance   uint64 // Current spending balance (plaintext)
	CurrentBalanceCt string // 132 hex chars, current ConfidentialBalanceSpending ciphertext
}

// BuildConvertBack queries ledger state, decrypts the holder's balance, and builds
// a ConfidentialMPTConvertBack transaction.
func BuildConvertBack(q LedgerQuerier, p BuildConvertBackParams) (*transaction.ConfidentialMPTConvertBack, error) {
	if err := validateConvertBackBase(p); err != nil {
		return nil, err
	}
	if err := p.BalanceRange.Validate(); err != nil {
		return nil, err
	}

	resolved, snapshot, err := resolveTxOptions(q, p.Account, p.TxOptions, transaction.ConfidentialMPTConvertBackTx)
	if err != nil {
		return nil, err
	}
	p.TxOptions = resolved

	issuance, err := getProvableIssuance(snapshot, p.IssuanceID)
	if err != nil {
		return nil, err
	}

	// Converting back more than the confidential supply fails with tecINSUFFICIENT_FUNDS,
	// and the issuance is already in hand, so reject it before reading the holder's MPToken.
	if p.Amount > issuance.confidentialOutstanding {
		return nil, ErrAmountExceedsOutstanding
	}

	holder, err := getMPTokenState(snapshot, issuance, p.IssuanceID, p.Account)
	if err != nil {
		return nil, err
	}

	if !sameEncryptionKey(holder.holderKey, p.HolderPubKey) {
		return nil, fmt.Errorf("%w: holder key", ErrKeyMismatch)
	}

	currentBalance, err := elgamal.Decrypt(holder.balanceCt, p.HolderPrivKey, p.BalanceRange)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decrypt balance: %w", ErrCryptoFailed, err)
	}

	return PrepareConvertBack(ConvertBackParams{
		BuildConvertBackParams: p,
		IssuerPubKey:           issuance.issuerKey,
		AuditorPubKey:          issuance.auditorKey,
		BalanceVersion:         holder.balanceVersion,
		CurrentBalance:         currentBalance,
		CurrentBalanceCt:       holder.balanceCt,
	})
}

// PrepareConvertBack builds a ConfidentialMPTConvertBack transaction.
//
// Steps:
// 1. Generate a shared blinding factor for the withdrawal amount.
// 2. Encrypt amount under holder, issuer (and optionally auditor) keys.
// 3. Generate a fresh blinding factor for the balance commitment.
// 4. Create Pedersen commitment for the current balance.
// 5. Compute convert-back context hash (account, issuance, seq, version).
// 6. Generate ZK proof linking balance commitment to on-ledger ciphertext.
func PrepareConvertBack(p ConvertBackParams) (*transaction.ConfidentialMPTConvertBack, error) {
	if err := validateConvertBackBase(p.BuildConvertBackParams); err != nil {
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
	if p.CurrentBalanceCt == "" {
		return nil, ErrMissingSenderState
	}
	if !transaction.IsValidCiphertext(p.CurrentBalanceCt) {
		return nil, ErrInvalidCiphertext
	}
	if p.Amount > p.CurrentBalance {
		return nil, ErrInsufficientBalance
	}
	proofSequence, err := p.validateForProof(p.Account, transaction.ConfidentialMPTConvertBackTx)
	if err != nil {
		return nil, err
	}

	bf, err := elgamal.GenerateBlindingFactor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	// Encrypt withdrawal amount under holder, issuer, and optionally auditor keys.
	holderCt, err := elgamal.Encrypt(p.Amount, p.HolderPubKey, bf)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}
	issuerCt, err := elgamal.Encrypt(p.Amount, p.IssuerPubKey, bf)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	var auditorCt string
	if p.AuditorPubKey != "" {
		auditorCt, err = elgamal.Encrypt(p.Amount, p.AuditorPubKey, bf)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
		}
	}

	// Generate a fresh blinding factor for balance commitment.
	// This ensures correctness even after MergeInbox operations.
	balanceBF, err := elgamal.GenerateBlindingFactor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	// Create balance commitment from current balance state with fresh BF.
	balanceCommit, err := commitment.Create(p.CurrentBalance, balanceBF)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	ctxHash, err := proof.ConvertBackContextHash(p.Account, p.IssuanceID, proofSequence, p.BalanceVersion)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	balanceProofParams := proof.Params{
		CommitmentHex:     balanceCommit,
		Amount:            p.CurrentBalance,
		CiphertextHex:     p.CurrentBalanceCt,
		BlindingFactorHex: balanceBF,
	}

	proofHex, err := proof.GenerateConvertBackProof(p.HolderPrivKey, p.HolderPubKey, ctxHash, p.Amount, balanceProofParams)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	tx := &transaction.ConfidentialMPTConvertBack{
		BaseTx:                baseTx(p.Account, transaction.ConfidentialMPTConvertBackTx, p.TxOptions),
		MPTokenIssuanceID:     p.IssuanceID,
		MPTAmount:             types.MPTPlainAmount(p.Amount),
		HolderEncryptedAmount: holderCt,
		IssuerEncryptedAmount: issuerCt,
		BlindingFactor:        bf,
		BalanceCommitment:     balanceCommit,
		ZKProof:               proofHex,
	}

	if auditorCt != "" {
		tx.AuditorEncryptedAmount = &auditorCt
	}

	if err := validatePreparedTransaction(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func validateConvertBackBase(p BuildConvertBackParams) error {
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
	if err := validateAmount(p.Amount); err != nil {
		return err
	}
	if p.HolderPrivKey == "" {
		return ErrMissingHolderKey
	}
	if !isValidPrivateKey(p.HolderPrivKey) {
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
