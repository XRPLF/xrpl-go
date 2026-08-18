package builder

import (
	"errors"
	"fmt"

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
	Account        string
	Destination    string
	DestinationTag *uint32
	IssuanceID     string
	Amount         uint64
	SenderPrivKey  string              // Non-zero secp256k1 scalar below the curve order, also used to decrypt balance from ledger
	SenderPubKey   string              // Valid 33-byte compressed secp256k1 point
	BalanceRange   elgamal.AmountRange // Inclusive balance decryption bounds
	CredentialIDs  types.CredentialIDs // Optional
}

// SendParams holds inputs for PrepareSend. BalanceRange is used only by
// BuildSend and has no effect when the caller supplies CurrentBalance directly.
type SendParams struct {
	BuildSendParams
	ReceiverPubKey   string // 66 hex chars (receiver's registered encryption key)
	IssuerPubKey     string // 66 hex chars
	AuditorPubKey    string // 66 hex chars, empty if no auditor
	Sequence         uint32 // Final transaction sequence bound into the proof. It must not change after preparation.
	BalanceVersion   uint32 // From MPToken.ConfidentialBalanceVersion
	CurrentBalance   uint64 // Sender's known plaintext spending balance
	CurrentBalanceCt string // 132 hex chars, current ConfidentialBalanceSpending ciphertext
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

	seq, snapshot, err := beginBuild(q, p.Account)
	if err != nil {
		return nil, err
	}

	issuance, err := getProvableIssuance(snapshot, p.IssuanceID)
	if err != nil {
		return nil, err
	}
	if !issuance.canTransfer() {
		return nil, ErrTransferDisabled
	}
	if issuance.transferFee > 0 {
		return nil, ErrTransferFeeSet
	}

	sender, err := getMPTokenState(snapshot, issuance, p.IssuanceID, p.Account)
	if err != nil {
		return nil, nameParty(err, "sender")
	}

	// Validate sender pubkey matches ledger.
	if !sameEncryptionKey(sender.holderKey, p.SenderPubKey) {
		return nil, fmt.Errorf("%w: sender key", ErrKeyMismatch)
	}

	currentBalance, err := elgamal.Decrypt(sender.balanceCt, p.SenderPrivKey, p.BalanceRange)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decrypt balance: %w", ErrCryptoFailed, err)
	}

	receiverKey, err := getMPTokenReceiverState(snapshot, issuance, p.IssuanceID, p.Destination)
	if err != nil {
		if errors.Is(err, ErrMPTokenNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrReceiverNotOptedIn, err)
		}
		return nil, nameParty(err, "destination")
	}

	return PrepareSend(SendParams{
		BuildSendParams:  p,
		ReceiverPubKey:   receiverKey,
		IssuerPubKey:     issuance.issuerKey,
		AuditorPubKey:    issuance.auditorKey,
		Sequence:         seq,
		BalanceVersion:   sender.balanceVersion,
		CurrentBalance:   currentBalance,
		CurrentBalanceCt: sender.balanceCt,
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
	if p.Sequence == 0 {
		return nil, ErrMissingSequence
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

	balanceParams := proof.Params{
		CommitmentHex:     balanceCommit,
		Amount:            p.CurrentBalance,
		CiphertextHex:     p.CurrentBalanceCt,
		BlindingFactorHex: balanceBF,
	}

	proofHex, err := proof.GenerateSendProof(p.SenderPrivKey, p.SenderPubKey, p.Amount, participants, amountBF, ctxHash, amountCommit, balanceParams)
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
		DestinationTag:             p.DestinationTag,
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

	if err := validatePreparedTransaction(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func validateSendBase(p BuildSendParams) error {
	if p.Account == "" {
		return ErrMissingAccount
	}
	account, err := decodeBuilderAddress(p.Account)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidAccount, err)
	}
	if p.Destination == "" {
		return ErrMissingDestination
	}
	destination, err := decodeBuilderAddress(p.Destination)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDestination, err)
	}
	// Destination has a DestinationTag companion field, so a tagged X-address is
	// allowed as long as it does not duplicate an explicit DestinationTag.
	if destination.HasTag && p.DestinationTag != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDestination, transaction.ErrDuplicateXAddressTag)
	}
	if account.AccountID == destination.AccountID {
		return ErrSelfSend
	}
	if p.IssuanceID == "" {
		return ErrMissingIssuanceID
	}
	if err := validateHolderRole(p.IssuanceID, p.Account); err != nil {
		return err
	}
	if err := validateDestinationNotIssuer(p.IssuanceID, p.Destination); err != nil {
		return err
	}
	if err := validateAmount(p.Amount); err != nil {
		return err
	}
	if p.SenderPrivKey == "" {
		return ErrMissingSenderKey
	}
	if !isValidPrivateKey(p.SenderPrivKey) {
		return ErrInvalidPrivKey
	}
	if p.SenderPubKey == "" {
		return ErrMissingSenderKey
	}
	if !transaction.IsValidCompressedEncryptionKey(p.SenderPubKey) {
		return fmt.Errorf("sender pub key: %w", ErrInvalidPubKey)
	}
	if len(p.CredentialIDs) > 0 && !p.CredentialIDs.IsValid() {
		return ErrInvalidCredentialIDs
	}
	return nil
}
