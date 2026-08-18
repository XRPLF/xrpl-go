package proof

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
)

// GenerateConvertProof generates a Schnorr proof of knowledge for a ConfidentialMPTConvert transaction.
// pubkeyHex: 66 hex chars, privkeyHex: 64 hex chars, ctxHashHex: 64 hex chars.
// Returns 128 hex chars (64-byte proof).
func GenerateConvertProof(pubkeyHex, privkeyHex, ctxHashHex string) (string, error) {
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptsizes.PubKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPubKey, err)
	}
	privBytes, err := hexutil.DecodeFixedHex(privkeyHex, mptsizes.PrivKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPrivKey, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptsizes.HashOutputSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	pub := mptcrypto.PublicKey(pubBytes)
	priv := mptcrypto.PrivateKey(privBytes)
	hash := mptcrypto.ContextHash(hashBytes)

	proof, err := mptcrypto.GenerateConvertProof(pub, priv, hash)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrProofGenerationFailed, err)
	}
	return hex.EncodeToString(proof[:]), nil
}

// VerifyConvertProof verifies a Schnorr proof for a ConfidentialMPTConvert transaction.
// proofHex: 128 hex chars, pubkeyHex: 66 hex chars, ctxHashHex: 64 hex chars.
func VerifyConvertProof(proofHex, pubkeyHex, ctxHashHex string) error {
	proofBytes, err := hexutil.DecodeFixedHex(proofHex, mptsizes.SchnorrProofSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidProof, err)
	}
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptsizes.PubKeySize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPubKey, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptsizes.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var proof [mptsizes.SchnorrProofSize]byte
	copy(proof[:], proofBytes)
	pub := mptcrypto.PublicKey(pubBytes)
	hash := mptcrypto.ContextHash(hashBytes)

	if err := mptcrypto.VerifyConvertProof(proof, pub, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}
