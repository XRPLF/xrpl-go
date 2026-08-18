package proof

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
)

// GenerateClawbackProof generates a compact sigma proof for a ConfidentialMPTClawback transaction.
// Returns 128 hex chars (64-byte proof).
func GenerateClawbackProof(privkeyHex, pubkeyHex, ctxHashHex string, amount uint64, ciphertextHex string) (string, error) {
	privBytes, err := hexutil.DecodeFixedHex(privkeyHex, mptsizes.PrivKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPrivKey, err)
	}
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptsizes.PubKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPubKey, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptsizes.HashOutputSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(ciphertextHex, mptsizes.CiphertextSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}

	priv := mptcrypto.PrivateKey(privBytes)
	pub := mptcrypto.PublicKey(pubBytes)
	hash := mptcrypto.ContextHash(hashBytes)
	ct := mptcrypto.Ciphertext(ctBytes)

	proof, err := mptcrypto.GenerateClawbackProof(priv, pub, hash, amount, ct)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrProofGenerationFailed, err)
	}
	return hex.EncodeToString(proof[:]), nil
}

// VerifyClawbackProof verifies an equality proof for a ConfidentialMPTClawback transaction.
func VerifyClawbackProof(proofHex string, amount uint64, pubkeyHex, ciphertextHex, ctxHashHex string) error {
	proofBytes, err := hexutil.DecodeFixedHex(proofHex, mptsizes.CompactClawbackProofSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidProof, err)
	}
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptsizes.PubKeySize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPubKey, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(ciphertextHex, mptsizes.CiphertextSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptsizes.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var proof [mptsizes.CompactClawbackProofSize]byte
	copy(proof[:], proofBytes)
	pub := mptcrypto.PublicKey(pubBytes)
	ct := mptcrypto.Ciphertext(ctBytes)
	hash := mptcrypto.ContextHash(hashBytes)

	if err := mptcrypto.VerifyClawbackProof(proof, amount, pub, ct, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}
