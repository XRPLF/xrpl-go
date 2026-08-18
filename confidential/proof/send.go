package proof

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
)

type sendProofInputs struct {
	privateKey   mptcrypto.PrivateKey
	publicKey    mptcrypto.PublicKey
	participants []mptcrypto.Participant
	blinding     mptcrypto.BlindingFactor
	contextHash  mptcrypto.ContextHash
}

// GenerateSendProof generates the full proof (equality + linkage + range) for a ConfidentialMPTSend
// transaction from the amount commitment and balance proof parameters.
// Participants must contain the sender, destination, and issuer in that order, followed by the
// optional auditor. The generated proof is verified before it is returned.
// Returns a fixed-size proof string (946 bytes, 1892 hex chars).
func GenerateSendProof(privkeyHex string, pubkeyHex string, amount uint64, participants []Participant, txBFHex, ctxHashHex, amountCommitmentHex string, balanceParams Params) (string, error) {
	inputs, err := decodeSendProofInputs(privkeyHex, pubkeyHex, participants, txBFHex, ctxHashHex)
	if err != nil {
		return "", err
	}
	amountCommitmentBytes, err := hexutil.DecodeFixedHex(amountCommitmentHex, mptsizes.CommitmentSize)
	if err != nil {
		return "", fmt.Errorf("amount commitment: %w: %w", ErrInvalidCommitment, err)
	}
	decodedBalanceParams, err := decodeProofParams(balanceParams)
	if err != nil {
		return "", fmt.Errorf("balance params: %w", err)
	}
	amountCommitment := mptcrypto.Commitment(amountCommitmentBytes)

	generatedProof, err := mptcrypto.GenerateSendProof(inputs.privateKey, inputs.publicKey, amount, inputs.participants, inputs.blinding, inputs.contextHash, amountCommitment, decodedBalanceParams)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrProofGenerationFailed, err)
	}
	if err := mptcrypto.VerifySendProof(generatedProof, inputs.participants, decodedBalanceParams.Ciphertext, amountCommitment, decodedBalanceParams.Commitment, inputs.contextHash); err != nil {
		return "", fmt.Errorf("%w: generated proof verification failed: %w", ErrProofGenerationFailed, err)
	}
	return hex.EncodeToString(generatedProof), nil
}

func decodeSendProofInputs(privkeyHex, pubkeyHex string, participants []Participant, txBFHex, ctxHashHex string) (sendProofInputs, error) {
	var inputs sendProofInputs
	privBytes, err := hexutil.DecodeFixedHex(privkeyHex, mptsizes.PrivKeySize)
	if err != nil {
		return inputs, fmt.Errorf("%w: %w", ErrInvalidPrivKey, err)
	}
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptsizes.PubKeySize)
	if err != nil {
		return inputs, fmt.Errorf("%w: %w", ErrInvalidPubKey, err)
	}
	parts, err := decodeSendParticipants(participants)
	if err != nil {
		return inputs, err
	}
	bfBytes, err := hexutil.DecodeFixedHex(txBFHex, mptsizes.BlindingFactorSize)
	if err != nil {
		return inputs, fmt.Errorf("%w: %w", ErrInvalidBlindingFactor, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptsizes.HashOutputSize)
	if err != nil {
		return inputs, fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}
	inputs.privateKey = mptcrypto.PrivateKey(privBytes)
	inputs.publicKey = mptcrypto.PublicKey(pubBytes)
	inputs.participants = parts
	inputs.blinding = mptcrypto.BlindingFactor(bfBytes)
	inputs.contextHash = mptcrypto.ContextHash(hashBytes)
	return inputs, nil
}

// VerifySendProof verifies the full proof for a ConfidentialMPTSend transaction.
// Participants must contain the sender, destination, and issuer in that order, followed by the optional auditor.
// balanceCtHex is the sender's on-ledger ConfidentialBalanceSpending ciphertext, which the proof
// links the balance commitment to. It is not the transaction's SenderEncryptedAmount.
func VerifySendProof(proofHex string, participants []Participant, balanceCtHex, amountCommitHex, balanceCommitHex, ctxHashHex string) error {
	proofBytes, err := hexutil.DecodeFixedHex(proofHex, mptsizes.SendProofSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidProof, err)
	}
	parts, err := decodeSendParticipants(participants)
	if err != nil {
		return err
	}
	balanceCtBytes, err := hexutil.DecodeFixedHex(balanceCtHex, mptsizes.CiphertextSize)
	if err != nil {
		return fmt.Errorf("balance ciphertext: %w: %w", ErrInvalidCiphertext, err)
	}
	amountCommitBytes, err := hexutil.DecodeFixedHex(amountCommitHex, mptsizes.CommitmentSize)
	if err != nil {
		return fmt.Errorf("amount commitment: %w: %w", ErrInvalidCommitment, err)
	}
	balanceCommitBytes, err := hexutil.DecodeFixedHex(balanceCommitHex, mptsizes.CommitmentSize)
	if err != nil {
		return fmt.Errorf("balance commitment: %w: %w", ErrInvalidCommitment, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptsizes.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	balanceCt := mptcrypto.Ciphertext(balanceCtBytes)
	amountCommit := mptcrypto.Commitment(amountCommitBytes)
	balanceCommit := mptcrypto.Commitment(balanceCommitBytes)
	hash := mptcrypto.ContextHash(hashBytes)

	if err := mptcrypto.VerifySendProof(proofBytes, parts, balanceCt, amountCommit, balanceCommit, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}

// decodeSendParticipants converts Participant values to mptcrypto.Participant values. A send
// proof binds the sender, destination, and issuer, followed by the optional auditor, so any
// other count is rejected.
func decodeSendParticipants(hps []Participant) ([]mptcrypto.Participant, error) {
	if len(hps) != 3 && len(hps) != 4 {
		return nil, ErrInvalidParticipantCount
	}
	parts := make([]mptcrypto.Participant, len(hps))
	for i, hp := range hps {
		p, err := decodeParticipant(hp)
		if err != nil {
			return nil, err
		}
		parts[i] = p
	}
	return parts, nil
}
