package proofs

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// GenerateClawbackProof generates an equality proof for a ConfidentialMPTClawback transaction.
// Returns 196 hex chars (98-byte proof).
func GenerateClawbackProof(privkeyHex, pubkeyHex, ctxHashHex string, amount uint64, ciphertextHex string) (string, error) {
	privBytes, err := hexutil.DecodeFixedHex(privkeyHex, mptcrypto.PrivKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPrivKeyLength, err)
	}
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptcrypto.PubKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPubKeyLength, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(ciphertextHex, mptcrypto.CiphertextSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidCiphertextLength, err)
	}

	var priv [mptcrypto.PrivKeySize]byte
	var pub [mptcrypto.PubKeySize]byte
	var hash [mptcrypto.HashOutputSize]byte
	var ct [mptcrypto.CiphertextSize]byte
	copy(priv[:], privBytes)
	copy(pub[:], pubBytes)
	copy(hash[:], hashBytes)
	copy(ct[:], ctBytes)

	proof, err := mptcrypto.GenerateClawbackProof(priv, pub, hash, amount, ct)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrProofGenerationFailed, err)
	}
	return hex.EncodeToString(proof[:]), nil
}

// VerifyClawbackProof verifies an equality proof for a ConfidentialMPTClawback transaction.
func VerifyClawbackProof(proofHex string, amount uint64, pubkeyHex, ciphertextHex, ctxHashHex string) error {
	proofBytes, err := hexutil.DecodeFixedHex(proofHex, mptcrypto.EqualityProofSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidProofLength, err)
	}
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptcrypto.PubKeySize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPubKeyLength, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(ciphertextHex, mptcrypto.CiphertextSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCiphertextLength, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var proof [mptcrypto.EqualityProofSize]byte
	var pub [mptcrypto.PubKeySize]byte
	var ct [mptcrypto.CiphertextSize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(proof[:], proofBytes)
	copy(pub[:], pubBytes)
	copy(ct[:], ctBytes)
	copy(hash[:], hashBytes)

	if err := mptcrypto.VerifyClawbackProof(proof, amount, pub, ct, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}
