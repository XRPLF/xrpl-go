//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package proof_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyConvertProof(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	ctxHash, err := proof.ConvertContextHash(testAccount, testIssuanceID, 1)
	require.NoError(t, err)

	proofHex, err := proof.GenerateConvertProof(kp.PubKeyHex, kp.PrivKeyHex, ctxHash)
	require.NoError(t, err)
	require.Len(t, proofHex, mptcrypto.SchnorrProofSize*2)

	err = proof.VerifyConvertProof(proofHex, kp.PubKeyHex, ctxHash)
	require.NoError(t, err)
}

func TestVerifyConvertProofWrongKey(t *testing.T) {
	kp1, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	kp2, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	ctxHash, err := proof.ConvertContextHash(testAccount, testIssuanceID, 1)
	require.NoError(t, err)

	proofHex, err := proof.GenerateConvertProof(kp1.PubKeyHex, kp1.PrivKeyHex, ctxHash)
	require.NoError(t, err)

	err = proof.VerifyConvertProof(proofHex, kp2.PubKeyHex, ctxHash)
	require.ErrorIs(t, err, proof.ErrProofVerificationFailed)
}

func TestConvertProofInvalidInputs(t *testing.T) {
	kp, _ := elgamal.GenerateKeypair()
	ctxHash, _ := proof.ConvertContextHash(testAccount, testIssuanceID, 1)

	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - generate bad pubkey",
			fn: func() error {
				_, err := proof.GenerateConvertProof("zz", kp.PrivKeyHex, ctxHash)
				return err
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - generate bad privkey",
			fn: func() error {
				_, err := proof.GenerateConvertProof(kp.PubKeyHex, "short", ctxHash)
				return err
			},
			wantErr: proof.ErrInvalidPrivKey,
		},
		{
			name: "fail - generate bad ctx hash",
			fn: func() error {
				_, err := proof.GenerateConvertProof(kp.PubKeyHex, kp.PrivKeyHex, "zz")
				return err
			},
			wantErr: proof.ErrInvalidContextHash,
		},
		{
			name: "fail - verify bad proof",
			fn: func() error {
				return proof.VerifyConvertProof("0102", kp.PubKeyHex, ctxHash)
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
