// Package elgamal provides a hex-string API for ElGamal keypair generation,
// encryption, and decryption. It wraps the byte-array functions in mptcrypto
// with hex encoding/decoding for use with XRPL transaction fields.
package elgamal

import (
	"encoding/hex"
	"fmt"
	"math"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// Keypair holds a hex-encoded ElGamal keypair.
type Keypair struct {
	PrivKeyHex string // 64 hex chars (32 bytes)
	PubKeyHex  string // 66 hex chars (33 bytes, compressed)
}

// AmountRange defines inclusive bounds for a decryption search.
type AmountRange struct {
	Low  uint64
	High uint64
}

// Validate checks that the inclusive decryption range can be searched safely.
func (r AmountRange) Validate() error {
	if r.Low > r.High {
		return fmt.Errorf("%w: low %d exceeds high %d", ErrInvalidAmountRange, r.Low, r.High)
	}
	if r.High == math.MaxUint64 {
		return fmt.Errorf("%w: high must be less than %d", ErrInvalidAmountRange, uint64(math.MaxUint64))
	}
	return nil
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
		return "", fmt.Errorf("%w: %w", ErrInvalidKey, err)
	}
	bfBytes, err := hexutil.DecodeFixedHex(bfHex, mptcrypto.BlindingFactorSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidBlindingFactor, err)
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

// Decrypt decrypts a ciphertext using a private key by searching amountRange.
// ciphertextHex: 132 hex chars (66 bytes), privkeyHex: 64 hex chars (32 bytes).
// The amount range bounds are inclusive and the search cost is linear.
func Decrypt(ciphertextHex, privateKeyHex string, amountRange AmountRange) (uint64, error) {
	if err := amountRange.Validate(); err != nil {
		return 0, err
	}

	ctBytes, err := hexutil.DecodeFixedHex(ciphertextHex, mptcrypto.CiphertextSize)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}
	privBytes, err := hexutil.DecodeFixedHex(privateKeyHex, mptcrypto.PrivKeySize)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidKey, err)
	}

	var ct [mptcrypto.CiphertextSize]byte
	var priv [mptcrypto.PrivKeySize]byte
	copy(ct[:], ctBytes)
	copy(priv[:], privBytes)

	result, err := mptcrypto.DecryptAmount(ct, priv, amountRange.Low, amountRange.High)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrDecryptFailed, err)
	}
	return result, nil
}
