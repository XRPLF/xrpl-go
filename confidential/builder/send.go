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

// BuildSendParams holds minimal inputs for BuildSend.
// Sequence, ReceiverPubKey, IssuerPubKey, AuditorPubKey, BalanceVersion, CurrentBalanceCt,
// and CurrentBalance are auto-resolved from the ledger. Balance is decrypted using SenderPrivKey
// within BalanceRange's inclusive bounds.
type BuildSendParams struct {
	Account       string
	Destination   string
	IssuanceID    string
	Amount        uint64
	SenderPrivKey string              // 64 hex chars, also used to decrypt balance from ledger
	SenderPubKey  string              // 66 hex chars (compressed)
	BalanceRange  elgamal.AmountRange // Inclusive balance decryption bounds
	CredentialIDs []string            // Optional
}

// SendParams holds inputs for PrepareSend.
type SendParams struct {
	BuildSendParams
	ReceiverPubKey   string // 66 hex chars (receiver's registered encryption key)
	IssuerPubKey     string // 66 hex chars
	AuditorPubKey    string // 66 hex chars, empty if no auditor
	Sequence         uint32
	BalanceVersion   uint32 // From MPToken.ConfidentialBalanceVersion
	CurrentBalance   uint64 // Sender's known plaintext spending balance
	CurrentBalanceCt string // 132 hex chars, current ConfidentialBalanceSpending ciphertext
}

func validateSendBase(p BuildSendParams) error {
	if p.Account == "" {
		return ErrMissingAccount
	}
	if !addresscodec.IsValidAddress(p.Account) {
		return ErrInvalidAccount
	}
	if p.Destination == "" {
		return ErrMissingDestination
	}
	if !addresscodec.IsValidAddress(p.Destination) {
		return ErrInvalidDestination
	}
	if p.Account == p.Destination {
		return ErrSelfSend
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
	if p.SenderPrivKey == "" {
		return ErrMissingSenderKey
	}
	if !transaction.IsValidPrivKey(p.SenderPrivKey) {
		return ErrInvalidPrivKey
	}
	if p.SenderPubKey == "" {
		return ErrMissingSenderKey
	}
	if !transaction.IsValidCompressedEncryptionKey(p.SenderPubKey) {
		return fmt.Errorf("sender pub key: %w", ErrInvalidPubKey)
	}
	return nil
}

// BuildSend queries ledger state, decrypts the sender's balance, and builds
// a ConfidentialMPTSend transaction.
func BuildSend(q LedgerQuerier, p BuildSendParams) (*transaction.ConfidentialMPTSend, error) {
	if err := validateSendBase(p); err != nil {
		return nil, err
	}
	if err := p.BalanceRange.Validate(); err != nil {
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

	senderLedgerKey, senderBalanceCt, balanceVersion, err := getMPTokenState(q, p.IssuanceID, p.Account)
	if err != nil {
		return nil, err
	}

	// Validate sender pubkey matches ledger.
	if senderLedgerKey != "" && senderLedgerKey != p.SenderPubKey {
		return nil, fmt.Errorf("%w: sender pubkey does not match ledger", ErrCryptoFailed)
	}

	currentBalance, err := elgamal.Decrypt(senderBalanceCt, p.SenderPrivKey, p.BalanceRange)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decrypt balance: %w", ErrCryptoFailed, err)
	}

	receiverKey, _, _, err := getMPTokenState(q, p.IssuanceID, p.Destination)
	if err != nil {
		return nil, ErrReceiverNotOptedIn
	}
	if receiverKey == "" {
		return nil, ErrReceiverNotOptedIn
	}

	return PrepareSend(SendParams{
		BuildSendParams:  p,
		ReceiverPubKey:   receiverKey,
		IssuerPubKey:     issuerKey,
		AuditorPubKey:    auditorKey,
		Sequence:         seq,
		BalanceVersion:   balanceVersion,
		CurrentBalance:   currentBalance,
		CurrentBalanceCt: senderBalanceCt,
	})
}

// PrepareSend builds a ConfidentialMPTSend transaction.
//
// Steps:
// 1. Encrypt amount under sender, receiver, issuer (and optionally auditor) keys with shared BF.
// 2. Create Pedersen commitment for the transfer amount.
// 3. Create Pedersen commitment for the current balance (fresh BF).
// 4. Compute send context hash (account, issuance, seq, dest, version).
// 5. Build participant list and proof params.
// 6. Generate composite ZK proof (range + linkage + equality).
func PrepareSend(p SendParams) (*transaction.ConfidentialMPTSend, error) {
	if err := validateSendBase(p.BuildSendParams); err != nil {
		return nil, err
	}
	if p.ReceiverPubKey == "" {
		return nil, ErrMissingReceiverKey
	}
	if !transaction.IsValidCompressedEncryptionKey(p.ReceiverPubKey) {
		return nil, fmt.Errorf("receiver pub key: %w", ErrInvalidPubKey)
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

	amountBF, err := elgamal.GenerateBlindingFactor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	senderCt, err := elgamal.Encrypt(p.Amount, p.SenderPubKey, amountBF)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}
	destCt, err := elgamal.Encrypt(p.Amount, p.ReceiverPubKey, amountBF)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}
	issuerCt, err := elgamal.Encrypt(p.Amount, p.IssuerPubKey, amountBF)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	var auditorCt string
	if p.AuditorPubKey != "" {
		auditorCt, err = elgamal.Encrypt(p.Amount, p.AuditorPubKey, amountBF)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
		}
	}

	amountCommit, err := commitment.Create(p.Amount, amountBF)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	balanceBF, err := elgamal.GenerateBlindingFactor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}
	balanceCommit, err := commitment.Create(p.CurrentBalance, balanceBF)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	ctxHash, err := proof.SendContextHash(p.Account, p.IssuanceID, p.Sequence, p.Destination, p.BalanceVersion)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	participants := []proof.Participant{
		{PubKeyHex: p.SenderPubKey, CiphertextHex: senderCt},
		{PubKeyHex: p.ReceiverPubKey, CiphertextHex: destCt},
		{PubKeyHex: p.IssuerPubKey, CiphertextHex: issuerCt},
	}
	if p.AuditorPubKey != "" {
		participants = append(participants, proof.Participant{
			PubKeyHex:     p.AuditorPubKey,
			CiphertextHex: auditorCt,
		})
	}

	amountParams := proof.Params{
		CommitmentHex:     amountCommit,
		Amount:            p.Amount,
		CiphertextHex:     senderCt,
		BlindingFactorHex: amountBF,
	}
	balanceParams := proof.Params{
		CommitmentHex:     balanceCommit,
		Amount:            p.CurrentBalance,
		CiphertextHex:     p.CurrentBalanceCt,
		BlindingFactorHex: balanceBF,
	}

	proofHex, err := proof.GenerateSendProof(p.SenderPrivKey, p.SenderPubKey, p.Amount, participants, amountBF, ctxHash, amountParams, balanceParams)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCryptoFailed, err)
	}

	tx := &transaction.ConfidentialMPTSend{
		BaseTx: transaction.BaseTx{
			Account:         types.Address(p.Account),
			TransactionType: transaction.ConfidentialMPTSendTx,
			Sequence:        p.Sequence,
		},
		MPTokenIssuanceID:          p.IssuanceID,
		Destination:                types.Address(p.Destination),
		SenderEncryptedAmount:      senderCt,
		DestinationEncryptedAmount: destCt,
		IssuerEncryptedAmount:      issuerCt,
		ZKProof:                    proofHex,
		AmountCommitment:           amountCommit,
		BalanceCommitment:          balanceCommit,
	}

	if auditorCt != "" {
		tx.AuditorEncryptedAmount = &auditorCt
	}

	if len(p.CredentialIDs) > 0 {
		tx.CredentialIDs = p.CredentialIDs
	}

	return tx, nil
}
