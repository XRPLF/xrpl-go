//go:build !cgo || js || wasip1 || tinygo || gofuzz || !(linux || darwin) || !(amd64 || arm64)

package mptcrypto_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/stretchr/testify/require"
)

func TestDecryptAmountWithoutCgo(t *testing.T) {
	_, err := mptcrypto.DecryptAmount(mptcrypto.Ciphertext{}, mptcrypto.PrivateKey{}, 2, 1)
	require.ErrorIs(t, err, mptcrypto.ErrCgoRequired)
}

// TestCiphertextArithmeticWithoutCgo pins that the homomorphic operations report the missing
// native backend rather than returning a zero ciphertext a caller could mistake for a result.
func TestCiphertextArithmeticWithoutCgo(t *testing.T) {
	_, err := mptcrypto.AddCiphertexts(mptcrypto.Ciphertext{}, mptcrypto.Ciphertext{})
	require.ErrorIs(t, err, mptcrypto.ErrCgoRequired)

	_, err = mptcrypto.SubtractCiphertexts(mptcrypto.Ciphertext{}, mptcrypto.Ciphertext{})
	require.ErrorIs(t, err, mptcrypto.ErrCgoRequired)
}

func TestCanonicalEncryptedZeroWithoutCgo(t *testing.T) {
	_, err := mptcrypto.CanonicalEncryptedZero(mptcrypto.PublicKey{}, [20]byte{}, [24]byte{})
	require.ErrorIs(t, err, mptcrypto.ErrCgoRequired)
}
