//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package proof_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyClawbackProof(t *testing.T) {
	const clawbackAmount uint64 = 500

	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	issuerBalanceCt, err := elgamal.Encrypt(clawbackAmount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)

	ctxHash, err := proof.ClawbackContextHash(testAccount, testIssuanceID, 1, testHolder)
	require.NoError(t, err)

	proofHex, err := proof.GenerateClawbackProof(issuerKP.PrivKeyHex, issuerKP.PubKeyHex, ctxHash, clawbackAmount, issuerBalanceCt)
	require.NoError(t, err)
	require.NotEmpty(t, proofHex)

	tests := []struct {
		name         string
		verifyAmount uint64
		wantErr      error
	}{
		{
			name:         "correct amount",
			verifyAmount: clawbackAmount,
		},
		{
			name:         "wrong amount",
			verifyAmount: 999,
			wantErr:      proof.ErrProofVerificationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := proof.VerifyClawbackProof(proofHex, tt.verifyAmount, issuerKP.PubKeyHex, issuerBalanceCt, ctxHash)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestClawbackProofInvalidInputs(t *testing.T) {
	kp, _ := elgamal.GenerateKeypair()
	ctxHash, _ := proof.ClawbackContextHash(testAccount, testIssuanceID, 1, testHolder)

	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - generate bad privkey",
			fn: func() error {
				_, err := proof.GenerateClawbackProof("zz", kp.PubKeyHex, ctxHash, 100, zeroHex(66))
				return err
			},
			wantErr: proof.ErrInvalidPrivKey,
		},
		{
			name: "fail - generate bad ciphertext",
			fn: func() error {
				_, err := proof.GenerateClawbackProof(kp.PrivKeyHex, kp.PubKeyHex, ctxHash, 100, "bad")
				return err
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
		{
			name: "fail - verify bad proof",
			fn: func() error {
				return proof.VerifyClawbackProof("0102", 100, kp.PubKeyHex, zeroHex(66), ctxHash)
			},
			wantErr: proof.ErrInvalidProof,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
