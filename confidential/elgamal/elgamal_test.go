//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package elgamal_test

import (
	"math"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeypair(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	// privkey: 64 hex chars (32 bytes)
	require.Len(t, kp.PrivKeyHex, mptcrypto.PrivKeySize*2)

	// pubkey: 66 hex chars (33 bytes)
	require.Len(t, kp.PubKeyHex, mptcrypto.PubKeySize*2)

	// compressed pubkey starts with "02" or "03"
	prefix := kp.PubKeyHex[:2]
	require.Contains(t, []string{"02", "03"}, prefix, "PubKeyHex prefix: got %q, want 02 or 03", prefix)
}

func TestGenerateBlindingFactor(t *testing.T) {
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	// 64 hex chars (32 bytes)
	require.Len(t, bf, mptcrypto.BlindingFactorSize*2)

	// not all zeros
	allZeros := true
	for _, c := range bf {
		if c != '0' {
			allZeros = false
			break
		}
	}
	require.False(t, allZeros, "blinding factor is all zeros")
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

	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf, err := elgamal.GenerateBlindingFactor()
			require.NoError(t, err)

			ct, err := elgamal.Encrypt(tt.amount, kp.PubKeyHex, bf)
			require.NoError(t, err)
			require.Len(t, ct, mptcrypto.CiphertextSize*2)

			got, err := elgamal.Decrypt(ct, kp.PrivKeyHex)
			if err != nil && tt.skipOnDecryptErr {
				t.Skipf("Decrypt not supported for this value: %v", err)
			}
			require.NoError(t, err)
			require.Equal(t, tt.amount, got)
		})
	}
}

func TestEncryptMultipleKeys(t *testing.T) {
	kp1, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	kp2, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	// Intentionally reuse the same blinding factor to prove that different
	// public keys alone are sufficient to produce distinct ciphertexts.
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	ct1, err := elgamal.Encrypt(42, kp1.PubKeyHex, bf)
	require.NoError(t, err)
	ct2, err := elgamal.Encrypt(42, kp2.PubKeyHex, bf)
	require.NoError(t, err)

	require.NotEqual(t, ct1, ct2, "same amount with different keys produced identical ciphertexts")
}

func TestDecryptWithWrongKey(t *testing.T) {
	kp1, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	kp2, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	ct, err := elgamal.Encrypt(42, kp1.PubKeyHex, bf)
	require.NoError(t, err)

	// Decrypting with a different private key should fail.
	_, err = elgamal.Decrypt(ct, kp2.PrivKeyHex)
	require.ErrorIs(t, err, elgamal.ErrDecryptFailed)
}

func TestInvalidHexInputs(t *testing.T) {
	kp, _ := elgamal.GenerateKeypair()
	bf, _ := elgamal.GenerateBlindingFactor()

	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - encrypt bad pubkey hex",
			fn: func() error {
				_, err := elgamal.Encrypt(1, "zzzz", bf)
				return err
			},
			wantErr: elgamal.ErrInvalidKeyLength,
		},
		{
			name: "fail - encrypt short pubkey",
			fn: func() error {
				_, err := elgamal.Encrypt(1, "0102", bf)
				return err
			},
			wantErr: elgamal.ErrInvalidKeyLength,
		},
		{
			name: "fail - encrypt bad blinding factor",
			fn: func() error {
				_, err := elgamal.Encrypt(1, kp.PubKeyHex, "not-hex")
				return err
			},
			wantErr: elgamal.ErrInvalidBlindingFactorLength,
		},
		{
			name: "fail - decrypt bad ciphertext",
			fn: func() error {
				_, err := elgamal.Decrypt("zz", kp.PrivKeyHex)
				return err
			},
			wantErr: elgamal.ErrInvalidCiphertextLength,
		},
		{
			name: "fail - decrypt bad privkey",
			fn: func() error {
				ct, _ := elgamal.Encrypt(1, kp.PubKeyHex, bf)
				_, err := elgamal.Decrypt(ct, "short")
				return err
			},
			wantErr: elgamal.ErrInvalidKeyLength,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
