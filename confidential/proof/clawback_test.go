//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package proof_test

import (
	"encoding/hex"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyClawbackProof(t *testing.T) {
	const clawbackAmount uint64 = 500

	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	otherKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	issuerBalanceCt, err := elgamal.Encrypt(clawbackAmount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)
	otherBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	otherBalanceCt, err := elgamal.Encrypt(clawbackAmount, issuerKP.PubKeyHex, otherBF)
	require.NoError(t, err)

	ctxHash, err := proof.ClawbackContextHash(testAccount, testIssuanceID, 1, testHolder)
	require.NoError(t, err)
	otherCtxHash, err := proof.ClawbackContextHash(testAccount, testIssuanceID, 2, testHolder)
	require.NoError(t, err)

	proofHex, err := proof.GenerateClawbackProof(issuerKP.PrivKeyHex, issuerKP.PubKeyHex, ctxHash, clawbackAmount, issuerBalanceCt)
	require.NoError(t, err)
	require.Len(t, proofHex, mptsizes.CompactClawbackProofSize*2)
	proofBytes, err := hex.DecodeString(proofHex)
	require.NoError(t, err)
	proofBytes[0] ^= 1
	tamperedProof := hex.EncodeToString(proofBytes)

	_, err = proof.GenerateClawbackProof(issuerKP.PrivKeyHex, issuerKP.PubKeyHex, ctxHash, clawbackAmount+1, issuerBalanceCt)
	require.ErrorIs(t, err, proof.ErrProofGenerationFailed)

	tests := []struct {
		name       string
		amount     uint64
		pubKey     string
		ciphertext string
		proofHex   string
		context    string
		wantErr    error
	}{
		{
			name:       "correct inputs",
			amount:     clawbackAmount,
			proofHex:   proofHex,
			pubKey:     issuerKP.PubKeyHex,
			ciphertext: issuerBalanceCt,
			context:    ctxHash,
		},
		{
			name:       "tampered proof",
			amount:     clawbackAmount,
			pubKey:     issuerKP.PubKeyHex,
			ciphertext: issuerBalanceCt,
			proofHex:   tamperedProof,
			context:    ctxHash,
			wantErr:    proof.ErrProofVerificationFailed,
		},
		{
			name:       "wrong amount",
			amount:     999,
			proofHex:   proofHex,
			pubKey:     issuerKP.PubKeyHex,
			ciphertext: issuerBalanceCt,
			context:    ctxHash,
			wantErr:    proof.ErrProofVerificationFailed,
		},
		{
			name:       "wrong public key",
			amount:     clawbackAmount,
			proofHex:   proofHex,
			pubKey:     otherKP.PubKeyHex,
			ciphertext: issuerBalanceCt,
			context:    ctxHash,
			wantErr:    proof.ErrProofVerificationFailed,
		},
		{
			name:       "wrong ciphertext",
			amount:     clawbackAmount,
			proofHex:   proofHex,
			pubKey:     issuerKP.PubKeyHex,
			ciphertext: otherBalanceCt,
			context:    ctxHash,
			wantErr:    proof.ErrProofVerificationFailed,
		},
		{
			name:       "wrong context",
			amount:     clawbackAmount,
			proofHex:   proofHex,
			pubKey:     issuerKP.PubKeyHex,
			ciphertext: issuerBalanceCt,
			context:    otherCtxHash,
			wantErr:    proof.ErrProofVerificationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := proof.VerifyClawbackProof(tt.proofHex, tt.amount, tt.pubKey, tt.ciphertext, tt.context)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestClawbackProofInvalidInputs(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	otherKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	ciphertext, err := elgamal.Encrypt(100, kp.PubKeyHex, bf)
	require.NoError(t, err)
	ctxHash, err := proof.ClawbackContextHash(testAccount, testIssuanceID, 1, testHolder)
	require.NoError(t, err)
	proofHex, err := proof.GenerateClawbackProof(kp.PrivKeyHex, kp.PubKeyHex, ctxHash, 100, ciphertext)
	require.NoError(t, err)

	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - generate bad privkey",
			fn: func() error {
				_, err := proof.GenerateClawbackProof("zz", kp.PubKeyHex, ctxHash, 100, ciphertext)
				return err
			},
			wantErr: proof.ErrInvalidPrivKey,
		},
		{
			name: "fail - generate bad public key",
			fn: func() error {
				_, err := proof.GenerateClawbackProof(kp.PrivKeyHex, "zz", ctxHash, 100, ciphertext)
				return err
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - generate bad context",
			fn: func() error {
				_, err := proof.GenerateClawbackProof(kp.PrivKeyHex, kp.PubKeyHex, "zz", 100, ciphertext)
				return err
			},
			wantErr: proof.ErrInvalidContextHash,
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
			name: "fail - generate mismatched keypair",
			fn: func() error {
				_, err := proof.GenerateClawbackProof(otherKP.PrivKeyHex, kp.PubKeyHex, ctxHash, 100, ciphertext)
				return err
			},
			wantErr: proof.ErrProofGenerationFailed,
		},
		{
			name: "fail - verify bad proof",
			fn: func() error {
				return proof.VerifyClawbackProof("0102", 100, kp.PubKeyHex, ciphertext, ctxHash)
			},
			wantErr: proof.ErrInvalidProof,
		},
		{
			name: "fail - verify bad public key",
			fn: func() error {
				return proof.VerifyClawbackProof(proofHex, 100, "zz", ciphertext, ctxHash)
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - verify bad ciphertext",
			fn: func() error {
				return proof.VerifyClawbackProof(proofHex, 100, kp.PubKeyHex, "bad", ctxHash)
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
		{
			name: "fail - verify bad context",
			fn: func() error {
				return proof.VerifyClawbackProof(proofHex, 100, kp.PubKeyHex, ciphertext, "zz")
			},
			wantErr: proof.ErrInvalidContextHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// setupClawbackProofScenario builds a valid issuer key pair, context hash, ciphertext, and
// proof for tests that only need a well-formed starting point.
func setupClawbackProofScenario(t *testing.T) (kp elgamal.Keypair, ctxHash, ciphertext, proofHex string) {
	t.Helper()
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	ciphertext, err = elgamal.Encrypt(100, kp.PubKeyHex, bf)
	require.NoError(t, err)
	ctxHash, err = proof.ClawbackContextHash(testAccount, testIssuanceID, 1, testHolder)
	require.NoError(t, err)
	proofHex, err = proof.GenerateClawbackProof(kp.PrivKeyHex, kp.PubKeyHex, ctxHash, 100, ciphertext)
	require.NoError(t, err)
	return kp, ctxHash, ciphertext, proofHex
}

// TestClawbackProofPreservesHexCause pins that the sentinel wraps the decoder's own error,
// so a caller can recover which byte was malformed and not just that something was.
func TestClawbackProofPreservesHexCause(t *testing.T) {
	kp, ctxHash, ciphertext, proofHex := setupClawbackProofScenario(t)

	_, err := proof.GenerateClawbackProof("zz", kp.PubKeyHex, ctxHash, 100, ciphertext)
	var generationHexCause hex.InvalidByteError
	require.ErrorAs(t, err, &generationHexCause)

	err = proof.VerifyClawbackProof(proofHex, 100, "zz", ciphertext, ctxHash)
	var verificationHexCause hex.InvalidByteError
	require.ErrorAs(t, err, &verificationHexCause)
}
