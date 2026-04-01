//go:build cgo

package proof_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/commitment"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyConvertBackProof(t *testing.T) {
	const balanceAmount uint64 = 1000
	const withdrawAmount uint64 = 100

	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	balanceCt, err := elgamal.Encrypt(balanceAmount, kp.PubKeyHex, bf)
	require.NoError(t, err)

	balanceCommit, err := commitment.Create(balanceAmount, bf)
	require.NoError(t, err)

	ctxHash, err := proof.ConvertBackContextHash(testAccount, testIssuanceID, 1, 0)
	require.NoError(t, err)

	params := proof.HexProofParams{
		CommitmentHex:     balanceCommit,
		Amount:            balanceAmount,
		CiphertextHex:     balanceCt,
		BlindingFactorHex: bf,
	}

	proofHex, err := proof.GenerateConvertBackProof(kp.PrivKeyHex, kp.PubKeyHex, ctxHash, withdrawAmount, params)
	require.NoError(t, err)
	require.NotEmpty(t, proofHex)

	err = proof.VerifyConvertBackProof(proofHex, kp.PubKeyHex, balanceCt, balanceCommit, withdrawAmount, ctxHash)
	require.NoError(t, err)
}

func TestConvertBackProofInvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - bad privkey",
			fn: func() error {
				_, err := proof.GenerateConvertBackProof("zz", "02"+zeroHex(32), zeroHex(32), 100, proof.HexProofParams{
					CommitmentHex:     "02" + zeroHex(32),
					CiphertextHex:     zeroHex(66),
					BlindingFactorHex: zeroHex(32),
				})
				return err
			},
			wantErr: proof.ErrInvalidPrivKey,
		},
		{
			name: "fail - bad pubkey",
			fn: func() error {
				_, err := proof.GenerateConvertBackProof(zeroHex(32), "zz", zeroHex(32), 100, proof.HexProofParams{
					CommitmentHex:     "02" + zeroHex(32),
					CiphertextHex:     zeroHex(66),
					BlindingFactorHex: zeroHex(32),
				})
				return err
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - bad ctx hash",
			fn: func() error {
				_, err := proof.GenerateConvertBackProof(zeroHex(32), "02"+zeroHex(32), "zz", 100, proof.HexProofParams{
					CommitmentHex:     "02" + zeroHex(32),
					CiphertextHex:     zeroHex(66),
					BlindingFactorHex: zeroHex(32),
				})
				return err
			},
			wantErr: proof.ErrInvalidContextHash,
		},
		{
			name: "fail - bad commitment in params",
			fn: func() error {
				_, err := proof.GenerateConvertBackProof(zeroHex(32), "02"+zeroHex(32), zeroHex(32), 100, proof.HexProofParams{
					CommitmentHex:     "bad",
					CiphertextHex:     zeroHex(66),
					BlindingFactorHex: zeroHex(32),
				})
				return err
			},
			wantErr: proof.ErrInvalidCommitment,
		},
		{
			name: "fail - verify bad proof",
			fn: func() error {
				return proof.VerifyConvertBackProof("0102", "02"+zeroHex(32), zeroHex(66), "02"+zeroHex(32), 100, zeroHex(32))
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
