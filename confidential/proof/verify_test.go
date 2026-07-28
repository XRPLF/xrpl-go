//go:build cgo

package proof_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/stretchr/testify/require"
)

func TestVerifyRevealedAmountWithoutAuditor(t *testing.T) {
	const amount uint64 = 42

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	holderCt, err := elgamal.Encrypt(amount, holderKP.PubKeyHex, bf)
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(amount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)

	holder := proof.Participant{PubKeyHex: holderKP.PubKeyHex, CiphertextHex: holderCt}
	issuer := proof.Participant{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: issuerCt}

	err = proof.VerifyRevealedAmount(amount, bf, holder, issuer, nil)
	require.NoError(t, err)
}

func TestVerifyRevealedAmountWithAuditor(t *testing.T) {
	const amount uint64 = 42

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	auditorKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	holderCt, err := elgamal.Encrypt(amount, holderKP.PubKeyHex, bf)
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(amount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)
	auditorCt, err := elgamal.Encrypt(amount, auditorKP.PubKeyHex, bf)
	require.NoError(t, err)

	holder := proof.Participant{PubKeyHex: holderKP.PubKeyHex, CiphertextHex: holderCt}
	issuer := proof.Participant{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: issuerCt}
	auditor := proof.Participant{PubKeyHex: auditorKP.PubKeyHex, CiphertextHex: auditorCt}

	err = proof.VerifyRevealedAmount(amount, bf, holder, issuer, &auditor)
	require.NoError(t, err)
}

func TestVerifyRevealedAmountWrongAmount(t *testing.T) {
	const amount uint64 = 42

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	holderCt, err := elgamal.Encrypt(amount, holderKP.PubKeyHex, bf)
	require.NoError(t, err)
	issuerCt, err := elgamal.Encrypt(amount, issuerKP.PubKeyHex, bf)
	require.NoError(t, err)

	holder := proof.Participant{PubKeyHex: holderKP.PubKeyHex, CiphertextHex: holderCt}
	issuer := proof.Participant{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: issuerCt}

	err = proof.VerifyRevealedAmount(999, bf, holder, issuer, nil)
	require.ErrorIs(t, err, proof.ErrProofVerificationFailed)
}

func TestVerifyRevealedAmountInvalidInputs(t *testing.T) {
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
					proof.Participant{PubKeyHex: "zz", CiphertextHex: zeroHex(66)},
					proof.Participant{PubKeyHex: zeroHex(33), CiphertextHex: zeroHex(66)},
					nil)
			},
			wantErr: proof.ErrInvalidPubKey,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestVerifySendRangeProofRoundtrip(t *testing.T) {
	senderKP, participants, txBF, ctxHash, amountParams, balanceParams, _, amountCommitHex, balanceCommitHex := setupSendProofScenario(t, 500, 1000, false)

	proofHex, err := proof.GenerateSendProof(
		senderKP.PrivKeyHex, senderKP.PubKeyHex, 500, participants, txBF, ctxHash,
		amountParams, balanceParams,
	)
	require.NoError(t, err)

	rangeProofHex := proofHex[mptcrypto.CompactSendProofSize*2:]

	err = proof.VerifySendRangeProof(rangeProofHex, amountCommitHex, balanceCommitHex, ctxHash)
	require.NoError(t, err)
}
