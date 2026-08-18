//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package proof_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyConvertProof(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	wrongKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	ctxHash, err := proof.ConvertContextHash(testAccount, testIssuanceID, 1)
	require.NoError(t, err)
	wrongCtxHash, err := proof.ConvertContextHash(testAccount, testIssuanceID, 2)
	require.NoError(t, err)

	proofHex, err := proof.GenerateConvertProof(kp.PubKeyHex, kp.PrivKeyHex, ctxHash)
	require.NoError(t, err)
	require.Len(t, proofHex, mptsizes.SchnorrProofSize*2)

	tests := []struct {
		name        string
		proofHex    string
		pubKey      string
		contextHash string
		wantErr     error
	}{
		{name: "pass - correct statement", proofHex: proofHex, pubKey: kp.PubKeyHex, contextHash: ctxHash},
		{name: "fail - changed proof", proofHex: changedHex(proofHex), pubKey: kp.PubKeyHex, contextHash: ctxHash, wantErr: proof.ErrProofVerificationFailed},
		{name: "fail - wrong key", proofHex: proofHex, pubKey: wrongKP.PubKeyHex, contextHash: ctxHash, wantErr: proof.ErrProofVerificationFailed},
		{name: "fail - wrong context", proofHex: proofHex, pubKey: kp.PubKeyHex, contextHash: wrongCtxHash, wantErr: proof.ErrProofVerificationFailed},
		{name: "fail - malformed key", proofHex: proofHex, pubKey: "bad", contextHash: ctxHash, wantErr: proof.ErrInvalidPubKey},
		{name: "fail - malformed context", proofHex: proofHex, pubKey: kp.PubKeyHex, contextHash: "bad", wantErr: proof.ErrInvalidContextHash},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := proof.VerifyConvertProof(tt.proofHex, tt.pubKey, tt.contextHash)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestConvertProofInvalidInputs(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	otherKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	ctxHash, err := proof.ConvertContextHash(testAccount, testIssuanceID, 1)
	require.NoError(t, err)

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
			name: "fail - generate mismatched keypair",
			fn: func() error {
				_, err := proof.GenerateConvertProof(kp.PubKeyHex, otherKP.PrivKeyHex, ctxHash)
				return err
			},
			wantErr: proof.ErrProofGenerationFailed,
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
