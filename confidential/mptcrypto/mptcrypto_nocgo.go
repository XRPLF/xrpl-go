//go:build !cgo || js || wasip1 || tinygo || gofuzz || !(linux || darwin) || !(amd64 || arm64)

package mptcrypto

import "errors"

// ErrCgoRequired is returned by all crypto functions when built without CGo.
var ErrCgoRequired = errors.New(
	"mptcrypto: CGo is required for confidential MPT operations; " +
		"rebuild with CGO_ENABLED=1 and vendored mpt-crypto libraries",
)

// GenerateKeypair creates a new secp256k1 ElGamal keypair.
// Returns a 32-byte private key and a 33-byte compressed public key.
func GenerateKeypair() (privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, err error) {
	return privkey, pubkey, ErrCgoRequired
}

// GenerateBlindingFactor returns a random 32-byte scalar suitable for ElGamal encryption.
func GenerateBlindingFactor() (bf [BlindingFactorSize]byte, err error) {
	return bf, ErrCgoRequired
}

// EncryptAmount encrypts a uint64 amount under a compressed public key using a blinding factor.
// Returns a 66-byte ciphertext (two compressed EC points: C1 || C2).
func EncryptAmount(amount uint64, pubkey [PubKeySize]byte, bf [BlindingFactorSize]byte) (ct [CiphertextSize]byte, err error) {
	return ct, ErrCgoRequired
}

// DecryptAmount decrypts a 66-byte ElGamal ciphertext using a private key.
// Returns the plaintext uint64 amount.
func DecryptAmount(ciphertext [CiphertextSize]byte, privkey [PrivKeySize]byte) (uint64, error) {
	return 0, ErrCgoRequired
}
