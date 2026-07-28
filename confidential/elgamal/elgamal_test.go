//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package elgamal_test

import (
	"math"
	"strings"
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
		name        string
		amount      uint64
		amountRange elgamal.AmountRange
	}{
		{name: "pass - zero", amount: 0, amountRange: elgamal.AmountRange{Low: 0, High: 0}},
		{name: "pass - small value", amount: 42, amountRange: elgamal.AmountRange{Low: 40, High: 50}},
		{name: "pass - one million", amount: 1_000_000, amountRange: elgamal.AmountRange{Low: 1_000_000, High: 1_000_000}},
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

			got, err := elgamal.Decrypt(ct, kp.PrivKeyHex, tt.amountRange)
			require.NoError(t, err)
			require.Equal(t, tt.amount, got)
		})
	}
}

func TestDecryptOutsideRange(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	ciphertext, err := elgamal.Encrypt(42, kp.PubKeyHex, bf)
	require.NoError(t, err)

	_, err = elgamal.Decrypt(ciphertext, kp.PrivKeyHex, elgamal.AmountRange{Low: 0, High: 41})
	require.ErrorIs(t, err, elgamal.ErrDecryptFailed)
}

func TestAmountRangeValidate(t *testing.T) {
	tests := []struct {
		name        string
		amountRange elgamal.AmountRange
		wantErr     error
	}{
		{name: "pass - inclusive range", amountRange: elgamal.AmountRange{Low: 1, High: 2}},
		{name: "pass - single-value range", amountRange: elgamal.AmountRange{Low: 1, High: 1}},
		{name: "fail - low exceeds high", amountRange: elgamal.AmountRange{Low: 2, High: 1}, wantErr: elgamal.ErrInvalidAmountRange},
		{name: "fail - high is max uint64", amountRange: elgamal.AmountRange{Low: 0, High: math.MaxUint64}, wantErr: elgamal.ErrInvalidAmountRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.amountRange.Validate()
			require.ErrorIs(t, err, tt.wantErr)
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
	_, err = elgamal.Decrypt(ct, kp2.PrivKeyHex, elgamal.AmountRange{Low: 0, High: 100})
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
			wantErr: elgamal.ErrInvalidKey,
		},
		{
			name: "fail - encrypt short pubkey",
			fn: func() error {
				_, err := elgamal.Encrypt(1, "0102", bf)
				return err
			},
			wantErr: elgamal.ErrInvalidKey,
		},
		{
			name: "fail - encrypt bad blinding factor",
			fn: func() error {
				_, err := elgamal.Encrypt(1, kp.PubKeyHex, "not-hex")
				return err
			},
			wantErr: elgamal.ErrInvalidBlindingFactor,
		},
		{
			name: "fail - decrypt bad ciphertext",
			fn: func() error {
				_, err := elgamal.Decrypt("zz", kp.PrivKeyHex, elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidCiphertext,
		},
		{
			name: "fail - decrypt short ciphertext",
			fn: func() error {
				_, err := elgamal.Decrypt(strings.Repeat("00", mptcrypto.CiphertextSize-1), kp.PrivKeyHex, elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidCiphertext,
		},
		{
			name: "fail - decrypt long ciphertext",
			fn: func() error {
				_, err := elgamal.Decrypt(strings.Repeat("00", mptcrypto.CiphertextSize+1), kp.PrivKeyHex, elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidCiphertext,
		},
		{
			name: "fail - decrypt bad privkey",
			fn: func() error {
				ct, _ := elgamal.Encrypt(1, kp.PubKeyHex, bf)
				_, err := elgamal.Decrypt(ct, "short", elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidKey,
		},
		{
			name: "fail - decrypt short privkey",
			fn: func() error {
				ct, _ := elgamal.Encrypt(1, kp.PubKeyHex, bf)
				_, err := elgamal.Decrypt(ct, strings.Repeat("00", mptcrypto.PrivKeySize-1), elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidKey,
		},
		{
			name: "fail - decrypt long privkey",
			fn: func() error {
				ct, _ := elgamal.Encrypt(1, kp.PubKeyHex, bf)
				_, err := elgamal.Decrypt(ct, strings.Repeat("00", mptcrypto.PrivKeySize+1), elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidKey,
		},
		{
			name: "fail - decrypt invalid range before decoding inputs",
			fn: func() error {
				_, err := elgamal.Decrypt("zz", "short", elgamal.AmountRange{Low: 2, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidAmountRange,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
