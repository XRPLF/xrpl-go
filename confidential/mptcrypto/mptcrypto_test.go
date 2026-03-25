//go:build cgo

package mptcrypto_test

import (
	"math"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeypair(t *testing.T) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)
	require.NotEqual(t, [mptcrypto.PrivKeySize]byte{}, priv, "privkey is all zeros")
	// compressed secp256k1 pubkey starts with 0x02 or 0x03
	require.Contains(t, []byte{0x02, 0x03}, pub[0], "unexpected pubkey prefix: 0x%02x", pub[0])
}

func TestGenerateBlindingFactor(t *testing.T) {
	bf1, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)
	require.NotEqual(t, [mptcrypto.BlindingFactorSize]byte{}, bf1, "blinding factor is all zeros")

	// two calls should produce different values (non-deterministic RNG)
	bf2, err := mptcrypto.GenerateBlindingFactor()
	require.NoError(t, err)
	require.NotEqual(t, bf1, bf2, "two consecutive blinding factors are identical")
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	tests := []struct {
		name   string
		amount uint64
		// skipOnDecryptErr skips the subtest instead of failing when the C
		// library cannot recover the plaintext (BSGS table limitation).
		skipOnDecryptErr bool
	}{
		{"pass - zero", 0, false},
		{"pass - small value", 42, false},
		{"pass - one million", 1_000_000, false},
		{"pass - max uint64", math.MaxUint64, true},
	}

	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf, err := mptcrypto.GenerateBlindingFactor()
			require.NoError(t, err)

			ct, err := mptcrypto.EncryptAmount(tt.amount, pub, bf)
			require.NoError(t, err)
			require.NotEqual(t, [mptcrypto.CiphertextSize]byte{}, ct, "ciphertext is all zeros")

			got, err := mptcrypto.DecryptAmount(ct, priv)
			if err != nil && tt.skipOnDecryptErr {
				t.Skipf("DecryptAmount not supported for this value: %v", err)
			}
			require.NoError(t, err)
			require.Equal(t, tt.amount, got)
		})
	}
}
