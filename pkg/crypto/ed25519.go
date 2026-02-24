package crypto

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"strings"
)

const (
	// ed25519 prefix - value is 237
	ed25519Prefix byte = 0xED
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
	public := strings.ToUpper(hex.EncodeToString(pubKey))
	privKey = append([]byte{c.prefix}, privKey...)
	private := strings.ToUpper(hex.EncodeToString(privKey[:32+len([]byte{c.prefix})]))
	return private, public, nil
}

// Sign signs a message using the ED25519 algorithm with the provided private key.
func (c ED25519CryptoAlgorithm) Sign(msg, privKey string) (string, error) {
	b, err := hex.DecodeString(privKey)
	if err != nil {
		return "", err
	}
	rawPriv := ed25519.NewKeyFromSeed(b[1:])
	signedMsg := ed25519.Sign(rawPriv, []byte(msg))
	return strings.ToUpper(hex.EncodeToString(signedMsg)), nil
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

	return ed25519.Verify(ed25519.PublicKey(bp[1:]), []byte(msg), bs)
}
