//go:build cgo

package proofs_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/confidential/proofs"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyConvertProof(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	ctxHash, err := proofs.ConvertContextHash(testAccount, testIssuanceID, 1)
	require.NoError(t, err)

	proof, err := proofs.GenerateConvertProof(kp.PubKeyHex, kp.PrivKeyHex, ctxHash)
	require.NoError(t, err)
	require.Len(t, proof, mptcrypto.SchnorrProofSize*2)

	err = proofs.VerifyConvertProof(proof, kp.PubKeyHex, ctxHash)
	require.NoError(t, err)
}

func TestVerifyConvertProofWrongKey(t *testing.T) {
	kp1, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	kp2, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	ctxHash, err := proofs.ConvertContextHash(testAccount, testIssuanceID, 1)
	require.NoError(t, err)

	proof, err := proofs.GenerateConvertProof(kp1.PubKeyHex, kp1.PrivKeyHex, ctxHash)
	require.NoError(t, err)

	err = proofs.VerifyConvertProof(proof, kp2.PubKeyHex, ctxHash)
	require.ErrorIs(t, err, proofs.ErrProofVerificationFailed)
}

func TestConvertProofInvalidInputs(t *testing.T) {
	kp, _ := elgamal.GenerateKeypair()
	ctxHash, _ := proofs.ConvertContextHash(testAccount, testIssuanceID, 1)

	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - generate bad pubkey",
			fn: func() error {
				_, err := proofs.GenerateConvertProof("zz", kp.PrivKeyHex, ctxHash)
				return err
			},
			wantErr: proofs.ErrInvalidPubKeyLength,
		},
		{
			name: "fail - generate bad privkey",
			fn: func() error {
				_, err := proofs.GenerateConvertProof(kp.PubKeyHex, "short", ctxHash)
				return err
			},
			wantErr: proofs.ErrInvalidPrivKeyLength,
		},
		{
			name: "fail - generate bad ctx hash",
			fn: func() error {
				_, err := proofs.GenerateConvertProof(kp.PubKeyHex, kp.PrivKeyHex, "zz")
				return err
			},
			wantErr: proofs.ErrInvalidContextHash,
		},
		{
			name: "fail - verify bad proof",
			fn: func() error {
				return proofs.VerifyConvertProof("0102", kp.PubKeyHex, ctxHash)
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
