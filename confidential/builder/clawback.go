// Package builder provides transaction builders for confidential MPT operations.
//
// These builders accept an address in either form. xrplhash.MPToken and the account
// query decode classic addresses only, so an address is normalized to its classic
// spelling before it reaches them. The proof layer resolves either form itself and binds
// the decoded AccountID. The self-send and self-clawback guards likewise compare decoded
// AccountIDs rather than the strings the caller supplied.
//
// A tagged X-address is accepted only where the transaction has a companion tag field to
// carry the tag, and rejected there when it would duplicate an explicit tag.
//
// # Tickets
//
// Every builder accepts a TicketSequence. xrpld hashes the sequence proxy into every
// confidential context hash, so a proof commits to the ticket the transaction spends rather
// than to an account sequence, and the Ticket is spent from the transaction Account's own
// account root, never the Delegate's.
//
// One case a Ticket cannot rescue, and no builder can detect: ConfidentialMPTSend and
// ConfidentialMPTConvertBack commit their proofs to the submitter's own
// ConfidentialBalanceVersion, which a send, a convert-back, a merge-inbox, or a clawback
// against that holder bumps. Of several such transactions built against a single reading of
// one MPToken, the first to land bumps the version and the rest fail with tecBAD_PROOF, each
// paying a fee and destroying its Ticket. The collision is per MPToken, so transactions on
// different issuances never collide, and a builder cannot see what else the account has in
// flight. Submit them one at a time on that issuance and wait for validation, and reserve the
// Ticket for the transactions that can run in parallel: clawbacks against different holders,
// merges and converts across different issuances, and anything a repeat convert covers.
package builder

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// BuildClawbackParams holds minimal inputs for BuildClawback.
// The sequence in TxOptions, IssuerPubKey, IssuerCiphertext, and Amount are auto-resolved
// from the ledger.
type BuildClawbackParams struct {
	TxOptions
	Account       string // Issuer
	Holder        string
	IssuanceID    string
	IssuerPrivKey string              // Non-zero secp256k1 scalar below the curve order, also used to decrypt the holder balance
	BalanceRange  elgamal.AmountRange // Inclusive bounds for decrypting the holder's balance
}

// ClawbackParams holds inputs for PrepareClawback. BalanceRange is used only by
// BuildClawback and has no effect when the caller supplies Amount directly.
type ClawbackParams struct {
	BuildClawbackParams
	Amount           uint64 // The holder's complete confidential balance, which a clawback removes in full
	IssuerPubKey     string // Valid 33-byte compressed secp256k1 point from MPTokenIssuance.IssuerEncryptionKey
	IssuerCiphertext string // Two valid compressed secp256k1 points from the holder MPToken's IssuerEncryptedBalance
}

// BuildClawback queries ledger state, decrypts the holder's balance, and builds a
// ConfidentialMPTClawback transaction. The amount is the holder's complete balance,
// decrypted from IssuerEncryptedBalance within BalanceRange's inclusive bounds, capped
// at the issuance's ConfidentialOutstandingAmount.
func BuildClawback(q LedgerQuerier, p BuildClawbackParams) (*transaction.ConfidentialMPTClawback, error) {
	if err := validateClawbackBase(p); err != nil {
		return nil, err
	}
	if err := p.BalanceRange.Validate(); err != nil {
		return nil, err
	}

	resolved, snapshot, err := resolveTxOptions(q, p.Account, p.TxOptions, transaction.ConfidentialMPTClawbackTx)
	if err != nil {
		return nil, err
	}
	p.TxOptions = resolved

	issuance, err := getProvableIssuance(snapshot, p.IssuanceID)
	if err != nil {
		return nil, err
	}
	if !issuance.canClawback() {
		return nil, ErrClawbackDisabled
	}

	issuerCt, err := getIssuerCiphertext(snapshot, p.IssuanceID, p.Holder)
	if err != nil {
		return nil, err
	}

	// The confidential supply bounds every holder's balance, and a clawback above it fails
	// with tecINSUFFICIENT_FUNDS, so it is both a preflight check and a ceiling that keeps
	// the decryption search from scanning further than the balance can possibly reach.
	searchRange := p.BalanceRange
	if searchRange.Low > issuance.confidentialOutstanding {
		// The range cannot contain the holder's balance, which is a problem with the bounds
		// the caller supplied rather than the protocol bound ErrAmountExceedsOutstanding
		// reports, so it is not that sentinel.
		return nil, fmt.Errorf("%w: BalanceRange.Low %d exceeds the issuance confidential outstanding amount %d",
			elgamal.ErrInvalidAmountRange, searchRange.Low, issuance.confidentialOutstanding)
	}
	if searchRange.High > issuance.confidentialOutstanding {
		searchRange.High = issuance.confidentialOutstanding
	}

	amount, err := elgamal.Decrypt(issuerCt, p.IssuerPrivKey, searchRange)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decrypt holder balance: %w", ErrCryptoFailed, err)
	}

	return PrepareClawback(ClawbackParams{
		BuildClawbackParams: p,
		Amount:              amount,
		IssuerPubKey:        issuance.issuerKey,
		IssuerCiphertext:    issuerCt,
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
	if err := validateAmount(p.Amount); err != nil {
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
	proofSequence, err := p.validateForProof(p.Account, transaction.ConfidentialMPTClawbackTx)
	if err != nil {
		return nil, err
	}

	ctxHash, err := proof.ClawbackContextHash(p.Account, p.IssuanceID, proofSequence, p.Holder)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	proofHex, err := proof.GenerateClawbackProof(p.IssuerPrivKey, p.IssuerPubKey, ctxHash, p.Amount, p.IssuerCiphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	tx := &transaction.ConfidentialMPTClawback{
		BaseTx:            baseTx(p.Account, transaction.ConfidentialMPTClawbackTx, p.TxOptions),
		MPTokenIssuanceID: p.IssuanceID,
		Holder:            types.Address(p.Holder),
		MPTAmount:         types.MPTPlainAmount(p.Amount),
		ZKProof:           proofHex,
	}

	if err := validatePreparedTransaction(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func validateClawbackBase(p BuildClawbackParams) error {
	if p.Account == "" {
		return ErrMissingAccount
	}
	account, err := decodeBuilderAddress(p.Account)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	if p.Holder == "" {
		return ErrMissingHolder
	}
	holder, err := decodeBuilderAddress(p.Holder)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidHolder, err)
	}
	// Holder has no companion tag field, so a tagged X-address cannot be used.
	if holder.HasTag {
		return fmt.Errorf("%w: %w", ErrInvalidHolder, transaction.ErrAccountIDTagNotAllowed)
	}
	if account.AccountID == holder.AccountID {
		return ErrSelfClawback
	}
	if p.IssuanceID == "" {
		return ErrMissingIssuanceID
	}
	if err := validateIssuerRole(p.IssuanceID, p.Account); err != nil {
		return err
	}
	if p.IssuerPrivKey == "" {
		return ErrMissingIssuerKey
	}
	if !isValidPrivateKey(p.IssuerPrivKey) {
		return ErrInvalidPrivKey
	}
	return nil
}
