//go:build cgo

package proofs_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proofs"
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

	ctxHash, err := proofs.ClawbackContextHash(testAccount, testIssuanceID, 1, testHolder)
	require.NoError(t, err)

	proof, err := proofs.GenerateClawbackProof(issuerKP.PrivKeyHex, issuerKP.PubKeyHex, ctxHash, clawbackAmount, issuerBalanceCt)
	require.NoError(t, err)
	require.NotEmpty(t, proof)

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
			wantErr:      proofs.ErrProofVerificationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := proofs.VerifyClawbackProof(proof, tt.verifyAmount, issuerKP.PubKeyHex, issuerBalanceCt, ctxHash)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestClawbackProofInvalidInputs(t *testing.T) {
	kp, _ := elgamal.GenerateKeypair()
	ctxHash, _ := proofs.ClawbackContextHash(testAccount, testIssuanceID, 1, testHolder)

	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - generate bad privkey",
			fn: func() error {
				_, err := proofs.GenerateClawbackProof("zz", kp.PubKeyHex, ctxHash, 100, zeroHex(66))
				return err
			},
			wantErr: proofs.ErrInvalidPrivKeyLength,
		},
		{
			name: "fail - generate bad ciphertext",
			fn: func() error {
				_, err := proofs.GenerateClawbackProof(kp.PrivKeyHex, kp.PubKeyHex, ctxHash, 100, "bad")
				return err
			},
			wantErr: proofs.ErrInvalidCiphertextLength,
		},
		{
			name: "fail - verify bad proof",
			fn: func() error {
				return proofs.VerifyClawbackProof("0102", 100, kp.PubKeyHex, zeroHex(66), ctxHash)
			},
			wantErr: proofs.ErrInvalidProofLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
