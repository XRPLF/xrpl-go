//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package proof_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
	"github.com/stretchr/testify/require"
)

func TestVerifyRevealedAmount(t *testing.T) {
	const amount uint64 = 42

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	auditorKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	wrongBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	holderCt, err := elgamal.Encrypt(amount, holderKP.PubKeyHex, bf)
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(amount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)
	auditorCt, err := elgamal.Encrypt(amount, auditorKP.PubKeyHex, bf)
	require.NoError(t, err)

	holder := proof.Participant{PubKeyHex: holderKP.PubKeyHex, CiphertextHex: holderCt}
	issuer := proof.Participant{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: issuerCt}
	auditor := &proof.Participant{PubKeyHex: auditorKP.PubKeyHex, CiphertextHex: auditorCt}
	wrongHolder := proof.Participant{PubKeyHex: holderKP.PubKeyHex, CiphertextHex: issuerCt}
	wrongIssuer := proof.Participant{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: holderCt}
	wrongAuditor := &proof.Participant{PubKeyHex: auditorKP.PubKeyHex, CiphertextHex: holderCt}

	tests := []struct {
		name           string
		verifyAmount   uint64
		blindingFactor string
		holder         proof.Participant
		issuer         proof.Participant
		auditor        *proof.Participant
		wantErr        error
	}{
		{name: "pass - without auditor", verifyAmount: amount, blindingFactor: bf, holder: holder, issuer: issuer},
		{name: "pass - with auditor", verifyAmount: amount, blindingFactor: bf, holder: holder, issuer: issuer, auditor: auditor},
		{name: "fail - wrong amount", verifyAmount: 999, blindingFactor: bf, holder: holder, issuer: issuer, wantErr: proof.ErrProofVerificationFailed},
		{name: "fail - wrong blinding factor", verifyAmount: amount, blindingFactor: wrongBF, holder: holder, issuer: issuer, wantErr: proof.ErrProofVerificationFailed},
		{name: "fail - wrong holder ciphertext", verifyAmount: amount, blindingFactor: bf, holder: wrongHolder, issuer: issuer, wantErr: proof.ErrProofVerificationFailed},
		{name: "fail - wrong issuer ciphertext", verifyAmount: amount, blindingFactor: bf, holder: holder, issuer: wrongIssuer, wantErr: proof.ErrProofVerificationFailed},
		{name: "fail - wrong auditor ciphertext", verifyAmount: amount, blindingFactor: bf, holder: holder, issuer: issuer, auditor: wrongAuditor, wantErr: proof.ErrProofVerificationFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := proof.VerifyRevealedAmount(tt.verifyAmount, tt.blindingFactor, tt.holder, tt.issuer, tt.auditor)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestVerifyRevealedAmountInvalidInputs(t *testing.T) {
	valid := proof.Participant{PubKeyHex: zeroHex(33), CiphertextHex: zeroHex(66)}
	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - bad blinding factor",
			fn: func() error {
				return proof.VerifyRevealedAmount(42, "bad",
					proof.Participant{PubKeyHex: zeroHex(33), CiphertextHex: zeroHex(66)},
					proof.Participant{PubKeyHex: zeroHex(33), CiphertextHex: zeroHex(66)},
					nil)
			},
			wantErr: proof.ErrInvalidBlindingFactor,
		},
		{
			name: "fail - bad holder pubkey",
			fn: func() error {
				return proof.VerifyRevealedAmount(42, zeroHex(32),
					proof.Participant{PubKeyHex: "zz", CiphertextHex: zeroHex(66)}, valid, nil)
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - bad holder ciphertext",
			fn: func() error {
				return proof.VerifyRevealedAmount(42, zeroHex(32),
					proof.Participant{PubKeyHex: zeroHex(33), CiphertextHex: "bad"}, valid, nil)
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
		{
			name: "fail - bad issuer pubkey",
			fn: func() error {
				return proof.VerifyRevealedAmount(42, zeroHex(32), valid,
					proof.Participant{PubKeyHex: "bad", CiphertextHex: zeroHex(66)}, nil)
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - bad issuer ciphertext",
			fn: func() error {
				return proof.VerifyRevealedAmount(42, zeroHex(32), valid,
					proof.Participant{PubKeyHex: zeroHex(33), CiphertextHex: "bad"}, nil)
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
		{
			name: "fail - bad auditor pubkey",
			fn: func() error {
				return proof.VerifyRevealedAmount(42, zeroHex(32), valid, valid,
					&proof.Participant{PubKeyHex: "bad", CiphertextHex: zeroHex(66)})
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - bad auditor ciphertext",
			fn: func() error {
				return proof.VerifyRevealedAmount(42, zeroHex(32), valid, valid,
					&proof.Participant{PubKeyHex: zeroHex(33), CiphertextHex: "bad"})
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestVerifySendRangeProofInvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - bad proof",
			fn: func() error {
				return proof.VerifySendRangeProof("zz", "02"+zeroHex(32), "02"+zeroHex(32), zeroHex(32))
			},
			wantErr: proof.ErrInvalidProof,
		},
		{
			name: "fail - bad amount commitment",
			fn: func() error {
				return proof.VerifySendRangeProof(zeroHex(754), "zz", "02"+zeroHex(32), zeroHex(32))
			},
			wantErr: proof.ErrInvalidCommitment,
		},
		{
			name: "fail - bad balance commitment",
			fn: func() error {
				return proof.VerifySendRangeProof(zeroHex(754), "02"+zeroHex(32), "zz", zeroHex(32))
			},
			wantErr: proof.ErrInvalidCommitment,
		},
		{
			name: "fail - bad context hash",
			fn: func() error {
				return proof.VerifySendRangeProof(zeroHex(754), "02"+zeroHex(32), "02"+zeroHex(32), "bad")
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

func TestVerifySendRangeProofRoundtrip(t *testing.T) {
	senderKP, participants, txBF, ctxHash, balanceParams, _, amountCommitHex, balanceCommitHex := setupSendProofScenario(t, 500, 1000, false)

	proofHex, err := proof.GenerateSendProof(
		senderKP.PrivKeyHex, senderKP.PubKeyHex, 500, participants, txBF, ctxHash,
		amountCommitHex, balanceParams,
	)
	require.NoError(t, err)

	rangeProofHex := proofHex[mptsizes.CompactSendProofSize*2:]

	err = proof.VerifySendRangeProof(rangeProofHex, amountCommitHex, balanceCommitHex, ctxHash)
	require.NoError(t, err)

	tests := []struct {
		name              string
		amountCommitment  string
		balanceCommitment string
		contextHash       string
	}{
		{name: "amount commitment", amountCommitment: balanceCommitHex, balanceCommitment: balanceCommitHex, contextHash: ctxHash},
		{name: "balance commitment", amountCommitment: amountCommitHex, balanceCommitment: amountCommitHex, contextHash: ctxHash},
		{name: "context hash", amountCommitment: amountCommitHex, balanceCommitment: balanceCommitHex, contextHash: changedHex(ctxHash)},
	}
	for _, tt := range tests {
		t.Run("fail - wrong "+tt.name, func(t *testing.T) {
			err := proof.VerifySendRangeProof(rangeProofHex, tt.amountCommitment, tt.balanceCommitment, tt.contextHash)
			require.ErrorIs(t, err, proof.ErrProofVerificationFailed)
		})
	}
}
