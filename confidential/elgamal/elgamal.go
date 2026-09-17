// Package elgamal provides a hex-string API for ElGamal keypair generation,
// encryption, and decryption. It wraps the byte-array functions in mptcrypto
// with hex encoding/decoding for use with XRPL transaction fields.
package elgamal

import (
	"encoding/hex"
	"fmt"
	"math"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
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
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptsizes.PubKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidKey, err)
	}
	bfBytes, err := hexutil.DecodeFixedHex(bfHex, mptsizes.BlindingFactorSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidBlindingFactor, err)
	}

	pub := mptcrypto.PublicKey(pubBytes)
	bf := mptcrypto.BlindingFactor(bfBytes)

	ct, err := mptcrypto.EncryptAmount(amount, pub, bf)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrEncryptFailed, err)
	}
	return hex.EncodeToString(ct[:]), nil
}

// Decrypt decrypts a ciphertext using a private key by searching amountRange.
// ciphertextHex: 132 hex chars (66 bytes), privateKeyHex: 64 hex chars (32 bytes).
// The amount range bounds are inclusive and the search cost is linear.
func Decrypt(ciphertextHex, privateKeyHex string, amountRange AmountRange) (uint64, error) {
	if err := amountRange.Validate(); err != nil {
		return 0, err
	}

	ctBytes, err := hexutil.DecodeFixedHex(ciphertextHex, mptsizes.CiphertextSize)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}
	privBytes, err := hexutil.DecodeFixedHex(privateKeyHex, mptsizes.PrivKeySize)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidKey, err)
	}

	ct := mptcrypto.Ciphertext(ctBytes)
	privateKey := mptcrypto.PrivateKey(privBytes)

	result, err := mptcrypto.DecryptAmount(ct, privateKey, amountRange.Low, amountRange.High)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrDecryptFailed, err)
	}
	return result, nil
}

// Add homomorphically adds two ciphertexts encrypted under the same public key, returning an
// encryption of the sum of their plaintexts.
func Add(firstHex, secondHex string) (string, error) {
	return combine(firstHex, secondHex, mptcrypto.AddCiphertexts)
}

// Subtract homomorphically subtracts the second ciphertext from the first, both encrypted under
// the same public key, returning an encryption of the difference of their plaintexts.
func Subtract(firstHex, secondHex string) (string, error) {
	return combine(firstHex, secondHex, mptcrypto.SubtractCiphertexts)
}

// combine decodes both hex operands, applies one of the native group operations, and re-encodes
// the result.
func combine(firstHex, secondHex string, op func(a, b mptcrypto.Ciphertext) (mptcrypto.Ciphertext, error)) (string, error) {
	first, err := hexutil.DecodeFixedHex(firstHex, mptsizes.CiphertextSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}
	second, err := hexutil.DecodeFixedHex(secondHex, mptsizes.CiphertextSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}

	result, err := op(mptcrypto.Ciphertext(first), mptcrypto.Ciphertext(second))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCiphertextArithmetic, err)
	}
	return hex.EncodeToString(result[:]), nil
}

// EncryptCanonicalZero returns the deterministic encryption of zero that xrpld stores when a
// confidential transactor initializes or resets a balance, such as the inbox a merge clears.
// pubkeyHex: 66 hex chars (33 bytes), the key the balance is encrypted under. account is the
// holder as a classic or X-address; both forms produce the same result. issuanceIDHex: 48 hex
// chars (24 bytes). Returns 132 hex chars (66-byte ciphertext).
func EncryptCanonicalZero(pubkeyHex, account, issuanceIDHex string) (string, error) {
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptsizes.PubKeySize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidKey, err)
	}
	decoded, err := addresscodec.DecodeAddress(account)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidAddress, err)
	}
	issuanceBytes, err := hexutil.DecodeFixedHex(issuanceIDHex, mptsizes.IssuanceIDSize)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidIssuanceID, err)
	}

	var accountID [mptsizes.AccountIDSize]byte
	copy(accountID[:], decoded.AccountID[:])

	ct, err := mptcrypto.CanonicalEncryptedZero(mptcrypto.PublicKey(pubBytes), accountID, [mptsizes.IssuanceIDSize]byte(issuanceBytes))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrEncryptFailed, err)
	}
	return hex.EncodeToString(ct[:]), nil
}
