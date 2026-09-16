//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package elgamal_test

import (
	"math"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeypair(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	// privkey: 64 hex chars (32 bytes)
	require.Len(t, kp.PrivKeyHex, mptsizes.PrivKeySize*2)

	// pubkey: 66 hex chars (33 bytes)
	require.Len(t, kp.PubKeyHex, mptsizes.PubKeySize*2)

	// compressed pubkey starts with "02" or "03"
	prefix := kp.PubKeyHex[:2]
	require.Contains(t, []string{"02", "03"}, prefix, "PubKeyHex prefix: got %q, want 02 or 03", prefix)
}

func TestGenerateBlindingFactor(t *testing.T) {
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	// 64 hex chars (32 bytes)
	require.Len(t, bf, mptsizes.BlindingFactorSize*2)

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
			require.Len(t, ct, mptsizes.CiphertextSize*2)

			got, err := elgamal.Decrypt(ct, kp.PrivKeyHex, tt.amountRange)
			require.NoError(t, err)
			require.Equal(t, tt.amount, got)
		})
	}
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

func TestDecryptFailures(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	wrongKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	ciphertext, err := elgamal.Encrypt(42, kp.PubKeyHex, bf)
	require.NoError(t, err)

	tests := []struct {
		name        string
		privateKey  string
		amountRange elgamal.AmountRange
	}{
		{name: "fail - amount outside range", privateKey: kp.PrivKeyHex, amountRange: elgamal.AmountRange{Low: 0, High: 41}},
		{name: "fail - wrong private key", privateKey: wrongKP.PrivKeyHex, amountRange: elgamal.AmountRange{Low: 0, High: 100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := elgamal.Decrypt(ciphertext, tt.privateKey, tt.amountRange)
			require.ErrorIs(t, err, elgamal.ErrDecryptFailed)
		})
	}
}

func TestInvalidHexInputs(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	ciphertext, err := elgamal.Encrypt(1, kp.PubKeyHex, bf)
	require.NoError(t, err)

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
				_, err := elgamal.Decrypt(strings.Repeat("00", mptsizes.CiphertextSize-1), kp.PrivKeyHex, elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidCiphertext,
		},
		{
			name: "fail - decrypt long ciphertext",
			fn: func() error {
				_, err := elgamal.Decrypt(strings.Repeat("00", mptsizes.CiphertextSize+1), kp.PrivKeyHex, elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidCiphertext,
		},
		{
			name: "fail - decrypt bad privkey",
			fn: func() error {
				_, err := elgamal.Decrypt(ciphertext, "short", elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidKey,
		},
		{
			name: "fail - decrypt short privkey",
			fn: func() error {
				_, err := elgamal.Decrypt(ciphertext, strings.Repeat("00", mptsizes.PrivKeySize-1), elgamal.AmountRange{Low: 0, High: 1})
				return err
			},
			wantErr: elgamal.ErrInvalidKey,
		},
		{
			name: "fail - decrypt long privkey",
			fn: func() error {
				_, err := elgamal.Decrypt(ciphertext, strings.Repeat("00", mptsizes.PrivKeySize+1), elgamal.AmountRange{Low: 0, High: 1})
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
			require.ErrorIs(t, tc.fn(), tc.wantErr)
		})
	}
}

// TestAddAndSubtract covers the homomorphic arithmetic a client uses to predict the balance a
// confidential transaction leaves behind: the ledger stores ciphertexts, and the transactors
// credit and debit them without ever decrypting, so a prediction has to do the same.
func TestAddAndSubtract(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	encrypt := func(amount uint64) string {
		bf, err := elgamal.GenerateBlindingFactor()
		require.NoError(t, err)
		ciphertext, err := elgamal.Encrypt(amount, kp.PubKeyHex, bf)
		require.NoError(t, err)
		return ciphertext
	}

	tests := []struct {
		name  string
		left  uint64
		right uint64
		want  uint64
		op    func(first, second string) (string, error)
	}{
		{name: "add", left: 40, right: 15, want: 55, op: elgamal.Add},
		{name: "add zero", left: 40, right: 0, want: 40, op: elgamal.Add},
		{name: "subtract", left: 40, right: 15, want: 25, op: elgamal.Subtract},
		{name: "subtract to zero", left: 40, right: 40, want: 0, op: elgamal.Subtract},
		{name: "subtract zero", left: 40, right: 0, want: 40, op: elgamal.Subtract},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.op(encrypt(tt.left), encrypt(tt.right))
			require.NoError(t, err)
			require.Len(t, result, mptsizes.CiphertextSize*2)

			decrypted, err := elgamal.Decrypt(result, kp.PrivKeyHex, elgamal.AmountRange{Low: 0, High: 100})
			require.NoError(t, err)
			require.Equal(t, tt.want, decrypted)
		})
	}
}

// TestAddAndSubtractErrors covers the inputs that have no ciphertext result. Subtracting a
// ciphertext from itself is the one that is not a malformed input: the difference is the
// curve's identity element, which has no encoding as a ciphertext.
func TestAddAndSubtractErrors(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	ciphertext, err := elgamal.Encrypt(7, kp.PubKeyHex, bf)
	require.NoError(t, err)

	tests := []struct {
		name    string
		first   string
		second  string
		op      func(first, second string) (string, error)
		wantErr error
	}{
		{name: "add first not hex", first: "zz", second: ciphertext, op: elgamal.Add, wantErr: elgamal.ErrInvalidCiphertext},
		{name: "add second wrong length", first: ciphertext, second: "00", op: elgamal.Add, wantErr: elgamal.ErrInvalidCiphertext},
		{name: "subtract first empty", first: "", second: ciphertext, op: elgamal.Subtract, wantErr: elgamal.ErrInvalidCiphertext},
		{name: "subtract identical ciphertexts", first: ciphertext, second: ciphertext, op: elgamal.Subtract, wantErr: elgamal.ErrCiphertextArithmetic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.op(tt.first, tt.second)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestAddIsHomomorphicAcrossKeys pins that the arithmetic is per key: a mirror balance held
// under the issuer's key tracks the holder's own balance only because the same amount is
// encrypted separately to each key, never because the ciphertexts are interchangeable.
func TestAddIsHomomorphicAcrossKeys(t *testing.T) {
	holder, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuer, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	holderCt, err := elgamal.Encrypt(12, holder.PubKeyHex, bf)
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(12, issuer.PubKeyHex, bf)
	require.NoError(t, err)

	holderSum, err := elgamal.Add(holderCt, holderCt)
	require.NoError(t, err)
	issuerSum, err := elgamal.Add(issuerCt, issuerCt)
	require.NoError(t, err)

	bounds := elgamal.AmountRange{Low: 0, High: 100}
	holderAmount, err := elgamal.Decrypt(holderSum, holder.PrivKeyHex, bounds)
	require.NoError(t, err)
	issuerAmount, err := elgamal.Decrypt(issuerSum, issuer.PrivKeyHex, bounds)
	require.NoError(t, err)

	require.Equal(t, uint64(24), holderAmount)
	require.Equal(t, uint64(24), issuerAmount)
	require.NotEqual(t, holderSum, issuerSum, "the same amount under two keys is two ciphertexts")
}
