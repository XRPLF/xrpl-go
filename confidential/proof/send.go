package proof

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// GenerateSendProof generates the full proof (equality + linkage + range) for a ConfidentialMPTSend transaction.
// Returns hex-encoded proof string (variable length depending on participant count).
// The C API limits participants to 255 (uint8); XRPL uses at most 4 (sender, dest, issuer, auditor).
func GenerateSendProof(privkeyHex string, amount uint64, participants []HexParticipant, txBFHex, ctxHashHex string, amountParams, balanceParams HexProofParams) (string, error) {
	privBytes, err := hexutil.DecodeFixedHex(privkeyHex, mptcrypto.PrivKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPrivKeyLength, err)
	}
	parts, err := decodeParticipants(participants)
	if err != nil {
		return "", err
	}
	bfBytes, err := hexutil.DecodeFixedHex(txBFHex, mptcrypto.BlindingFactorSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidBlindingFactor, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}
	ap, err := decodeProofParams(amountParams)
	if err != nil {
		return "", err
	}
	bp, err := decodeProofParams(balanceParams)
	if err != nil {
		return "", err
	}

	var priv [mptcrypto.PrivKeySize]byte
	var bf [mptcrypto.BlindingFactorSize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(priv[:], privBytes)
	copy(bf[:], bfBytes)
	copy(hash[:], hashBytes)

	proof, err := mptcrypto.GenerateSendProof(priv, amount, parts, bf, hash, ap, bp)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrProofGenerationFailed, err)
	}
	return hex.EncodeToString(proof), nil
}

// VerifySendProof verifies the full proof for a ConfidentialMPTSend transaction.
// The C API limits participants to 255 (uint8); XRPL uses at most 4 (sender, dest, issuer, auditor).
func VerifySendProof(proofHex string, participants []HexParticipant, senderCtHex, amountCommitHex, balanceCommitHex, ctxHashHex string) error {
	proofBytes, err := hex.DecodeString(proofHex)
	if err != nil {
		return fmt.Errorf("%w: invalid hex: %w", ErrInvalidProofLength, err)
	}
	parts, err := decodeParticipants(participants)
	if err != nil {
		return err
	}
	senderCtBytes, err := hexutil.DecodeFixedHex(senderCtHex, mptcrypto.CiphertextSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCiphertextLength, err)
	}
	amountCommitBytes, err := hexutil.DecodeFixedHex(amountCommitHex, mptcrypto.CommitmentSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitmentLength, err)
	}
	balanceCommitBytes, err := hexutil.DecodeFixedHex(balanceCommitHex, mptcrypto.CommitmentSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitmentLength, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var senderCt [mptcrypto.CiphertextSize]byte
	var amountCommit [mptcrypto.CommitmentSize]byte
	var balanceCommit [mptcrypto.CommitmentSize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(senderCt[:], senderCtBytes)
	copy(amountCommit[:], amountCommitBytes)
	copy(balanceCommit[:], balanceCommitBytes)
	copy(hash[:], hashBytes)

	if err := mptcrypto.VerifySendProof(proofBytes, parts, senderCt, amountCommit, balanceCommit, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}
