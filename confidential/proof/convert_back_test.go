//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package proof_test

import (
	"encoding/hex"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/commitment"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyConvertBackProof(t *testing.T) {
	const balanceAmount uint64 = 1000
	const withdrawAmount uint64 = 100

	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	otherKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	otherBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(balanceAmount, kp.PubKeyHex, bf)
	require.NoError(t, err)
	otherBalanceCt, err := elgamal.Encrypt(balanceAmount, kp.PubKeyHex, otherBF)
	require.NoError(t, err)
	balanceCommit, err := commitment.Create(balanceAmount, bf)
	require.NoError(t, err)
	otherBalanceCommit, err := commitment.Create(balanceAmount, otherBF)
	require.NoError(t, err)
	ctxHash, err := proof.ConvertBackContextHash(testAccount, testIssuanceID, 1, 0)
	require.NoError(t, err)
	otherCtxHash, err := proof.ConvertBackContextHash(testAccount, testIssuanceID, 1, 1)
	require.NoError(t, err)

	params := proof.Params{
		CommitmentHex:     balanceCommit,
		Amount:            balanceAmount,
		CiphertextHex:     balanceCt,
		BlindingFactorHex: bf,
	}
	proofHex, err := proof.GenerateConvertBackProof(kp.PrivKeyHex, kp.PubKeyHex, ctxHash, withdrawAmount, params)
	require.NoError(t, err)
	require.Len(t, proofHex, mptsizes.ConvertBackProofSize*2)

	proofBytes, err := hex.DecodeString(proofHex)
	require.NoError(t, err)
	proofBytes[0] ^= 1
	tamperedProof := hex.EncodeToString(proofBytes)

	tests := []struct {
		name       string
		proofHex   string
		pubKey     string
		ciphertext string
		commitment string
		amount     uint64
		context    string
		wantErr    error
	}{
		{name: "correct inputs", proofHex: proofHex, pubKey: kp.PubKeyHex, ciphertext: balanceCt, commitment: balanceCommit, amount: withdrawAmount, context: ctxHash},
		{name: "tampered proof", proofHex: tamperedProof, pubKey: kp.PubKeyHex, ciphertext: balanceCt, commitment: balanceCommit, amount: withdrawAmount, context: ctxHash, wantErr: proof.ErrProofVerificationFailed},
		{name: "wrong public key", proofHex: proofHex, pubKey: otherKP.PubKeyHex, ciphertext: balanceCt, commitment: balanceCommit, amount: withdrawAmount, context: ctxHash, wantErr: proof.ErrProofVerificationFailed},
		{name: "wrong ciphertext", proofHex: proofHex, pubKey: kp.PubKeyHex, ciphertext: otherBalanceCt, commitment: balanceCommit, amount: withdrawAmount, context: ctxHash, wantErr: proof.ErrProofVerificationFailed},
		{name: "wrong commitment", proofHex: proofHex, pubKey: kp.PubKeyHex, ciphertext: balanceCt, commitment: otherBalanceCommit, amount: withdrawAmount, context: ctxHash, wantErr: proof.ErrProofVerificationFailed},
		{name: "wrong amount", proofHex: proofHex, pubKey: kp.PubKeyHex, ciphertext: balanceCt, commitment: balanceCommit, amount: withdrawAmount + 1, context: ctxHash, wantErr: proof.ErrProofVerificationFailed},
		{name: "wrong context", proofHex: proofHex, pubKey: kp.PubKeyHex, ciphertext: balanceCt, commitment: balanceCommit, amount: withdrawAmount, context: otherCtxHash, wantErr: proof.ErrProofVerificationFailed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := proof.VerifyConvertBackProof(test.proofHex, test.pubKey, test.ciphertext, test.commitment, test.amount, test.context)
			require.ErrorIs(t, err, test.wantErr)
		})
	}

	_, err = proof.GenerateConvertBackProof(otherKP.PrivKeyHex, kp.PubKeyHex, ctxHash, withdrawAmount, params)
	require.ErrorIs(t, err, proof.ErrProofGenerationFailed)
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
				_, err := proof.GenerateConvertBackProof("zz", "02"+zeroHex(32), zeroHex(32), 100, proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: zeroHex(66), BlindingFactorHex: zeroHex(32)})
				return err
			},
			wantErr: proof.ErrInvalidPrivKey,
		},
		{
			name: "fail - bad pubkey",
			fn: func() error {
				_, err := proof.GenerateConvertBackProof(zeroHex(32), "zz", zeroHex(32), 100, proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: zeroHex(66), BlindingFactorHex: zeroHex(32)})
				return err
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - bad ctx hash",
			fn: func() error {
				_, err := proof.GenerateConvertBackProof(zeroHex(32), "02"+zeroHex(32), "zz", 100, proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: zeroHex(66), BlindingFactorHex: zeroHex(32)})
				return err
			},
			wantErr: proof.ErrInvalidContextHash,
		},
		{
			name: "fail - bad commitment in params",
			fn: func() error {
				_, err := proof.GenerateConvertBackProof(zeroHex(32), "02"+zeroHex(32), zeroHex(32), 100, proof.Params{CommitmentHex: "bad", CiphertextHex: zeroHex(66), BlindingFactorHex: zeroHex(32)})
				return err
			},
			wantErr: proof.ErrInvalidCommitment,
		},
		{
			name: "fail - bad ciphertext in params",
			fn: func() error {
				_, err := proof.GenerateConvertBackProof(zeroHex(32), "02"+zeroHex(32), zeroHex(32), 100, proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: "bad", BlindingFactorHex: zeroHex(32)})
				return err
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
		{
			name: "fail - bad blinding factor in params",
			fn: func() error {
				_, err := proof.GenerateConvertBackProof(zeroHex(32), "02"+zeroHex(32), zeroHex(32), 100, proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: zeroHex(66), BlindingFactorHex: "bad"})
				return err
			},
			wantErr: proof.ErrInvalidBlindingFactor,
		},
		{
			name: "fail - verify bad proof",
			fn: func() error {
				return proof.VerifyConvertBackProof("0102", "02"+zeroHex(32), zeroHex(66), "02"+zeroHex(32), 100, zeroHex(32))
			},
			wantErr: proof.ErrInvalidProof,
		},
		{
			name: "fail - verify bad public key",
			fn: func() error {
				return proof.VerifyConvertBackProof(zeroHex(816), "zz", zeroHex(66), "02"+zeroHex(32), 100, zeroHex(32))
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - verify bad ciphertext",
			fn: func() error {
				return proof.VerifyConvertBackProof(zeroHex(816), "02"+zeroHex(32), "bad", "02"+zeroHex(32), 100, zeroHex(32))
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
		{
			name: "fail - verify bad commitment",
			fn: func() error {
				return proof.VerifyConvertBackProof(zeroHex(816), "02"+zeroHex(32), zeroHex(66), "bad", 100, zeroHex(32))
			},
			wantErr: proof.ErrInvalidCommitment,
		},
		{
			name: "fail - verify bad context",
			fn: func() error {
				return proof.VerifyConvertBackProof(zeroHex(816), "02"+zeroHex(32), zeroHex(66), "02"+zeroHex(32), 100, "bad")
			},
			wantErr: proof.ErrInvalidContextHash,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorIs(t, test.fn(), test.wantErr)
		})
	}
}
