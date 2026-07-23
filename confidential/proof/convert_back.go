package proof

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// GenerateConvertBackProof generates a compact sigma + range proof for a ConfidentialMPTConvertBack transaction.
// Returns hex-encoded proof string (1632 hex chars = 816 bytes).
func GenerateConvertBackProof(privkeyHex, pubkeyHex, ctxHashHex string, amount uint64, params Params) (string, error) {
	privBytes, err := hexutil.DecodeFixedHex(privkeyHex, mptcrypto.PrivKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPrivKey, err)
	}
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptcrypto.PubKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPubKey, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}
	pp, err := decodeProofParams(params)
	if err != nil {
		return "", err
	}

	var priv [mptcrypto.PrivKeySize]byte
	var pub [mptcrypto.PubKeySize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(priv[:], privBytes)
	copy(pub[:], pubBytes)
	copy(hash[:], hashBytes)

	proof, err := mptcrypto.GenerateConvertBackProof(priv, pub, hash, amount, pp)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrProofGenerationFailed, err)
	}
	return hex.EncodeToString(proof[:]), nil
}

// VerifyConvertBackProof verifies a linkage + range proof for a ConfidentialMPTConvertBack transaction.
func VerifyConvertBackProof(proofHex, pubkeyHex, ciphertextHex, balanceCommitHex string, amount uint64, ctxHashHex string) error {
	proofBytes, err := hexutil.DecodeFixedHex(proofHex, mptcrypto.ConvertBackProofSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidProof, err)
	}
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptcrypto.PubKeySize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPubKey, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(ciphertextHex, mptcrypto.CiphertextSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}
	commitBytes, err := hexutil.DecodeFixedHex(balanceCommitHex, mptcrypto.CommitmentSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitment, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var proof [mptcrypto.ConvertBackProofSize]byte
	var pub [mptcrypto.PubKeySize]byte
	var ct [mptcrypto.CiphertextSize]byte
	var commit [mptcrypto.CommitmentSize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(proof[:], proofBytes)
	copy(pub[:], pubBytes)
	copy(ct[:], ctBytes)
	copy(commit[:], commitBytes)
	copy(hash[:], hashBytes)

	if err := mptcrypto.VerifyConvertBackProof(proof, pub, ct, commit, amount, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}
