// Package elgamal provides a hex-string API for ElGamal keypair generation,
// encryption, and decryption. It wraps the byte-array functions in mptcrypto
// with hex encoding/decoding for use with XRPL transaction fields.
package elgamal

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// Keypair holds a hex-encoded ElGamal keypair.
type Keypair struct {
	PrivKeyHex string // 64 hex chars (32 bytes)
	PubKeyHex  string // 66 hex chars (33 bytes, compressed)
}

// GenerateKeypair creates a new secp256k1 ElGamal keypair with hex-encoded keys.
func GenerateKeypair() (Keypair, error) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	if err != nil {
		return Keypair{}, err
	}
	return Keypair{
		PrivKeyHex: hex.EncodeToString(priv[:]),
		PubKeyHex:  hex.EncodeToString(pub[:]),
	}, nil
}

// GenerateBlindingFactor returns a random 32-byte scalar as a 64-char hex string.
func GenerateBlindingFactor() (string, error) {
	bf, err := mptcrypto.GenerateBlindingFactor()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bf[:]), nil
}

// Encrypt encrypts an amount under a compressed public key with a blinding factor.
// pubkeyHex: 66 hex chars (33 bytes), bfHex: 64 hex chars (32 bytes).
// Returns 132 hex chars (66-byte ciphertext).
func Encrypt(amount uint64, pubkeyHex, bfHex string) (string, error) {
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptcrypto.PubKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidKeyLength, err)
	}
	bfBytes, err := hexutil.DecodeFixedHex(bfHex, mptcrypto.BlindingFactorSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidBlindingFactorLength, err)
	}

	var pub [mptcrypto.PubKeySize]byte
	var bf [mptcrypto.BlindingFactorSize]byte
	copy(pub[:], pubBytes)
	copy(bf[:], bfBytes)

	ct, err := mptcrypto.EncryptAmount(amount, pub, bf)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrEncryptFailed, err)
	}
	return hex.EncodeToString(ct[:]), nil
}

// Decrypt decrypts a ciphertext using a private key.
// ciphertextHex: 132 hex chars (66 bytes), privkeyHex: 64 hex chars (32 bytes).
// Returns the plaintext uint64 amount.
func Decrypt(ciphertextHex, privkeyHex string) (uint64, error) {
	ctBytes, err := hexutil.DecodeFixedHex(ciphertextHex, mptcrypto.CiphertextSize)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidCiphertextLength, err)
	}
	privBytes, err := hexutil.DecodeFixedHex(privkeyHex, mptcrypto.PrivKeySize)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidKeyLength, err)
	}

	var ct [mptcrypto.CiphertextSize]byte
	var priv [mptcrypto.PrivKeySize]byte
	copy(ct[:], ctBytes)
	copy(priv[:], privBytes)

	result, err := mptcrypto.DecryptAmount(ct, priv)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrDecryptFailed, err)
	}
	return result, nil
}
