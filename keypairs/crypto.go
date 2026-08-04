// Package keypairs provides cryptographic key pair generation and management for XRPL.
package keypairs

import (
	"encoding/hex"

	"github.com/Peersyst/xrpl-go/keypairs/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
)

type keyType uint8

const (
	publicKeyType keyType = iota
	privateKeyType

	noKeyPrefix = -1
)

type keyFormat struct {
	keyType keyType
	prefix  int
	length  int
}

// acceptedKeyFormats is the complete key-type-aware format table used for algorithm selection.
// A noKeyPrefix entry matches keys of that length regardless of their first byte (used for
// raw 32-byte secp256k1 private keys). The secp256k1 implementation accepts the 65-byte
// uncompressed public form through secp256k1.ParsePubKey.
var acceptedKeyFormats = map[keyFormat]interfaces.KeypairCryptoAlg{
	{keyType: publicKeyType, prefix: 0xED, length: 33}: crypto.ED25519(),
	{keyType: publicKeyType, prefix: 0x02, length: 33}: crypto.SECP256K1(),
	{keyType: publicKeyType, prefix: 0x03, length: 33}: crypto.SECP256K1(),
	{keyType: publicKeyType, prefix: 0x04, length: 65}: crypto.SECP256K1(),

	{keyType: privateKeyType, prefix: 0xED, length: 33}:        crypto.ED25519(),
	{keyType: privateKeyType, prefix: noKeyPrefix, length: 32}: crypto.SECP256K1(),
	{keyType: privateKeyType, prefix: 0x00, length: 33}:        crypto.SECP256K1(),
}

// getCryptoImplementationFromKey returns the crypto implementation for a key's type,
// prefix, and exact decoded length, along with the decoded key bytes. Invalid hex and
// unsupported formats return the matching public- or private-key error without exposing key material.
func getCryptoImplementationFromKey(key string, keyType keyType) (interfaces.KeypairCryptoAlg, []byte, error) {
	decoded, err := hex.DecodeString(key)
	if err != nil {
		return nil, nil, invalidKeyFormatError(keyType)
	}

	// An exact first-byte prefix match wins, noKeyPrefix rows accept any first byte.
	prefixes := []int{noKeyPrefix}
	if len(decoded) > 0 {
		prefixes = []int{int(decoded[0]), noKeyPrefix}
	}
	for _, prefix := range prefixes {
		if algorithm, ok := acceptedKeyFormats[keyFormat{keyType: keyType, prefix: prefix, length: len(decoded)}]; ok {
			return algorithm, decoded, nil
		}
	}
	return nil, nil, invalidKeyFormatError(keyType)
}

func invalidKeyFormatError(keyType keyType) error {
	if keyType == privateKeyType {
		return ErrInvalidPrivateKeyFormat
	}
	return ErrInvalidPublicKeyFormat
}
