package builder

import (
	"bytes"
	"encoding/hex"

	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// privateKeyHexLength is the hex-encoded length of a private key scalar. Derived from the
// byte length the proof and elgamal decoders enforce, so the two cannot drift.
const privateKeyHexLength = 2 * mptsizes.PrivKeySize

// isValidPrivateKey reports whether key is a usable secp256k1 scalar: a non-zero value
// below the curve order. Private keys are builder inputs only. They are never carried by
// a transaction.
func isValidPrivateKey(key string) bool {
	if len(key) != privateKeyHexLength {
		return false
	}
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return false
	}
	var scalar secp256k1.ModNScalar
	return !scalar.SetByteSlice(keyBytes) && !scalar.IsZero()
}

// sameEncryptionKey compares two hex-encoded encryption keys by their decoded bytes, so
// that case differences in the hex do not read as different keys.
func sameEncryptionKey(first, second string) bool {
	firstBytes, firstErr := hex.DecodeString(first)
	secondBytes, secondErr := hex.DecodeString(second)
	return firstErr == nil && secondErr == nil && bytes.Equal(firstBytes, secondBytes)
}
