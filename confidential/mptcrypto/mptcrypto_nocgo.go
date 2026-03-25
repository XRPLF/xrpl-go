//go:build !cgo

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
