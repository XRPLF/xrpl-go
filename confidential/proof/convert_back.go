package proof

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
)

// GenerateConvertBackProof generates a compact sigma + range proof for a ConfidentialMPTConvertBack transaction.
// Returns hex-encoded proof string (1632 hex chars = 816 bytes).
func GenerateConvertBackProof(privkeyHex, pubkeyHex, ctxHashHex string, amount uint64, params Params) (string, error) {
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
	pp, err := decodeProofParams(params)
	if err != nil {
		return "", err
	}

	priv := mptcrypto.PrivateKey(privBytes)
	pub := mptcrypto.PublicKey(pubBytes)
	hash := mptcrypto.ContextHash(hashBytes)

	proof, err := mptcrypto.GenerateConvertBackProof(priv, pub, hash, amount, pp)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrProofGenerationFailed, err)
	}
	return hex.EncodeToString(proof[:]), nil
}

// VerifyConvertBackProof verifies a linkage + range proof for a ConfidentialMPTConvertBack transaction.
func VerifyConvertBackProof(proofHex, pubkeyHex, ciphertextHex, balanceCommitHex string, amount uint64, ctxHashHex string) error {
	proofBytes, err := hexutil.DecodeFixedHex(proofHex, mptsizes.ConvertBackProofSize)
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
	commitBytes, err := hexutil.DecodeFixedHex(balanceCommitHex, mptsizes.CommitmentSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitment, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptsizes.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var proof [mptsizes.ConvertBackProofSize]byte
	copy(proof[:], proofBytes)
	pub := mptcrypto.PublicKey(pubBytes)
	ct := mptcrypto.Ciphertext(ctBytes)
	commit := mptcrypto.Commitment(commitBytes)
	hash := mptcrypto.ContextHash(hashBytes)

	if err := mptcrypto.VerifyConvertBackProof(proof, pub, ct, commit, amount, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}
