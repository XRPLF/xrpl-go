// Package crypto provides cryptographic utilities for XRPL key management,
// including Ed25519 and secp256k1 key derivation, signing, and verification.
package crypto

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

const (
	// ed25519 prefix - value is 237
	ed25519Prefix byte = 0xED
	// XRPL Ed25519 keys are encoded as the ED prefix byte followed by the seed or public key bytes.
	ed25519PrivateKeyLength = 1 + ed25519.SeedSize
	ed25519PublicKeyLength  = 1 + ed25519.PublicKeySize
)

var (
	_                       Algorithm = &ED25519CryptoAlgorithm{}
	ed25519FamilySeedPrefix           = []byte{0x01, 0xe1, 0x4b}
)

// ED25519CryptoAlgorithm is the implementation of the ED25519 cryptographic algorithm.
type ED25519CryptoAlgorithm struct {
	prefix byte
}

// ED25519 returns the ED25519 cryptographic algorithm.
func ED25519() ED25519CryptoAlgorithm {
	return ED25519CryptoAlgorithm{
		prefix: ed25519Prefix,
	}
}

// Prefix returns the prefix for the ED25519 cryptographic algorithm.
func (c ED25519CryptoAlgorithm) Prefix() byte {
	return c.prefix
}

// FamilySeedPrefix returns the family seed prefix for the ED25519 cryptographic algorithm.
func (c ED25519CryptoAlgorithm) FamilySeedPrefix() []byte {
	return ed25519FamilySeedPrefix
}

// DeriveKeypair derives a keypair from a seed.
func (c ED25519CryptoAlgorithm) DeriveKeypair(decodedSeed []byte, validator bool) (string, string, error) {
	if validator {
		return "", "", ErrValidatorNotSupported
	}
	rawPriv := Sha512Half(decodedSeed)
	pubKey, privKey, err := ed25519.GenerateKey(bytes.NewBuffer(rawPriv))
	if err != nil {
		return "", "", err
	}
	pubKey = append([]byte{c.prefix}, pubKey...)
	public := hexutil.EncodeToUpperHex(pubKey)
	privKey = append([]byte{c.prefix}, privKey...)
	private := hexutil.EncodeToUpperHex(privKey[:32+len([]byte{c.prefix})])
	return private, public, nil
}

// Sign signs a message using the ED25519 algorithm with the provided private key.
func (c ED25519CryptoAlgorithm) Sign(msg, privKey string) (string, error) {
	b, err := hex.DecodeString(privKey)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPrivateKey, err)
	}
	// The private key is sliced below to remove the ED prefix, so reject malformed keys first.
	if len(b) != ed25519PrivateKeyLength || b[0] != ed25519Prefix {
		return "", ErrInvalidPrivateKey
	}
	rawPriv := ed25519.NewKeyFromSeed(b[1:])
	signedMsg := ed25519.Sign(rawPriv, []byte(msg))
	return hexutil.EncodeToUpperHex(signedMsg), nil
}

// Validate validates a signature for a message with a public key.
func (c ED25519CryptoAlgorithm) Validate(msg, pubkey, sig string) bool {
	bp, err := hex.DecodeString(pubkey)
	if err != nil {
		return false
	}

	bs, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	// Validate the ED prefix and lengths before stripping the prefix or verifying the signature.
	if len(bp) != ed25519PublicKeyLength || bp[0] != ed25519Prefix || len(bs) != ed25519.SignatureSize {
		return false
	}

	return ed25519.Verify(ed25519.PublicKey(bp[1:]), []byte(msg), bs)
}
