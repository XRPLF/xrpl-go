// Package keypairs provides cryptographic key pair generation and management for XRPL.
package keypairs

import (
	"encoding/hex"

	"github.com/Peersyst/xrpl-go/keypairs/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
)

// GetCryptoImplementationFromKey returns the CryptoImplementation based on the key.
// It returns nil if the key does not match any crypto implementation.
// Currently, only ED25519 and SECP256K1 are supported.
func getCryptoImplementationFromKey(k string) interfaces.KeypairCryptoAlg {
	if len(k) < 2 {
		return nil
	}

	prefix, err := hex.DecodeString(k[:2])
	if err != nil {
		return nil
	}

	if ed25519 := crypto.ED25519(); prefix[0] == ed25519.Prefix() {
		return ed25519
	}
	if secp256k1 := crypto.SECP256K1(); prefix[0] == secp256k1.Prefix() {
		return secp256k1
	}
	return nil
}
