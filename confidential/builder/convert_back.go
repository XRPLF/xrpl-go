package builder

import (
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/confidential/commitment"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// BuildConvertBackParams holds minimal inputs for BuildConvertBack.
// Sequence, IssuerPubKey, AuditorPubKey, BalanceVersion, CurrentBalanceCt,
// and CurrentBalance are auto-resolved from the ledger. Balance is decrypted using HolderPrivKey.
type BuildConvertBackParams struct {
	Account       string
	IssuanceID    string
	Amount        uint64
	HolderPrivKey string // 64 hex chars, also used to decrypt balance from ledger
	HolderPubKey  string // 66 hex chars (compressed)
}

// ConvertBackParams holds inputs for PrepareConvertBack.
type ConvertBackParams struct {
	BuildConvertBackParams
	IssuerPubKey     string // 66 hex chars
	AuditorPubKey    string // 66 hex chars, empty if no auditor
	Sequence         uint32
	BalanceVersion   uint32
	CurrentBalance   uint64 // Current spending balance (plaintext)
	CurrentBalanceCt string // 132 hex chars, current ConfidentialBalanceSpending ciphertext
}

func validateConvertBackBase(p BuildConvertBackParams) error {
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
	if p.Amount == 0 {
		return ErrZeroAmount
	}
	if p.HolderPrivKey == "" {
		return ErrMissingHolderKey
	}
	if !transaction.IsValidPrivKey(p.HolderPrivKey) {
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

// BuildConvertBack queries ledger state, decrypts the holder's balance, and builds
// a ConfidentialMPTConvertBack transaction.
func BuildConvertBack(q LedgerQuerier, p BuildConvertBackParams) (*transaction.ConfidentialMPTConvertBack, error) {
	if err := validateConvertBackBase(p); err != nil {
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

	holderLedgerKey, balanceCt, balanceVersion, err := getMPTokenState(q, p.IssuanceID, p.Account)
	if err != nil {
		return nil, err
	}

	if holderLedgerKey != "" && holderLedgerKey != p.HolderPubKey {
		return nil, fmt.Errorf("%w: holder pubkey does not match ledger", ErrCryptoFailed)
	}

	currentBalance, err := elgamal.Decrypt(balanceCt, p.HolderPrivKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decrypt balance: %w", ErrCryptoFailed, err)
	}

	return PrepareConvertBack(ConvertBackParams{
		BuildConvertBackParams: p,
		IssuerPubKey:           issuerKey,
		AuditorPubKey:          auditorKey,
		Sequence:               seq,
		BalanceVersion:         balanceVersion,
		CurrentBalance:         currentBalance,
		CurrentBalanceCt:       balanceCt,
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

	ctxHash, err := proof.ConvertBackContextHash(p.Account, p.IssuanceID, p.Sequence, p.BalanceVersion)
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
		BaseTx: transaction.BaseTx{
			Account:         types.Address(p.Account),
			TransactionType: transaction.ConfidentialMPTConvertBackTx,
			Sequence:        p.Sequence,
		},
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

	return tx, nil
}
