//go:build cgo

package proof_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/commitment"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/stretchr/testify/require"
)

// setupSendProofScenario creates a full scenario for testing ConfidentialMPTSend proof.
// Returns all the hex-encoded data needed to generate and verify a send proof.
func setupSendProofScenario(t *testing.T, sendAmount, senderBalance uint64, withAuditor bool) (
	senderKP elgamal.Keypair,
	participants []proof.Participant,
	txBF string,
	ctxHash string,
	amountParams proof.Params,
	balanceParams proof.Params,
	senderBalanceCt string,
	amountCommitHex string,
	balanceCommitHex string,
) {
	t.Helper()

	// Generate keypairs for all parties.
	senderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	destKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	// Transaction blinding factor (used for send amount encryption).
	txBF, err = elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	// Balance blinding factor.
	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	// Encrypt send amount under each participant's key with the same bf.
	senderAmountCt, err := elgamal.Encrypt(sendAmount, senderKP.PubKeyHex, txBF)
	require.NoError(t, err)
	destAmountCt, err := elgamal.Encrypt(sendAmount, destKP.PubKeyHex, txBF)
	require.NoError(t, err)
	issuerAmountCt, err := elgamal.Encrypt(sendAmount, issuerKP.PubKeyHex, txBF)
	require.NoError(t, err)

	// Sender's balance ciphertext.
	senderBalanceCt, err = elgamal.Encrypt(senderBalance, senderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	// Commitments.
	amountCommitHex, err = commitment.Create(sendAmount, txBF)
	require.NoError(t, err)
	balanceCommitHex, err = commitment.Create(senderBalance, balanceBF)
	require.NoError(t, err)

	// Context hash.
	ctxHash, err = proof.SendContextHash(testAccount, testIssuanceID, 1, testDest, 0)
	require.NoError(t, err)

	// Participants array.
	participants = []proof.Participant{
		{PubKeyHex: senderKP.PubKeyHex, CiphertextHex: senderAmountCt},
		{PubKeyHex: destKP.PubKeyHex, CiphertextHex: destAmountCt},
		{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: issuerAmountCt},
	}

	if withAuditor {
		auditorKP, err := elgamal.GenerateKeypair()
		require.NoError(t, err)
		auditorAmountCt, err := elgamal.Encrypt(sendAmount, auditorKP.PubKeyHex, txBF)
		require.NoError(t, err)
		participants = append(participants, proof.Participant{
			PubKeyHex:     auditorKP.PubKeyHex,
			CiphertextHex: auditorAmountCt,
		})
	}

	// Proof params.
	amountParams = proof.Params{
		CommitmentHex:     amountCommitHex,
		Amount:            sendAmount,
		CiphertextHex:     senderAmountCt,
		BlindingFactorHex: txBF,
	}
	balanceParams = proof.Params{
		CommitmentHex:     balanceCommitHex,
		Amount:            senderBalance,
		CiphertextHex:     senderBalanceCt,
		BlindingFactorHex: balanceBF,
	}

	return
}

func TestGenerateAndVerifySendProof(t *testing.T) {
	tests := []struct {
		name          string
		sendAmount    uint64
		senderBalance uint64
		withAuditor   bool
	}{
		{"pass - 3 participants", 500, 1000, false},
		{"pass - 4 participants with auditor", 500, 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			senderKP, participants, txBF, ctxHash, amountParams, balanceParams, senderBalanceCt, amountCommitHex, balanceCommitHex := setupSendProofScenario(t, tt.sendAmount, tt.senderBalance, tt.withAuditor)

			proofHex, err := proof.GenerateSendProof(
				senderKP.PrivKeyHex, senderKP.PubKeyHex, tt.sendAmount, participants, txBF, ctxHash,
				amountParams, balanceParams,
			)
			require.NoError(t, err)
			require.NotEmpty(t, proofHex)

			err = proof.VerifySendProof(proofHex, participants, senderBalanceCt, amountCommitHex, balanceCommitHex, ctxHash)
			require.NoError(t, err)
		})
	}
}

func TestVerifySendProofRejectsWrongLength(t *testing.T) {
	const amount = 500
	senderKP, participants, txBF, ctxHash, amountParams, balanceParams, senderBalanceCt, amountCommitHex, balanceCommitHex := setupSendProofScenario(t, amount, 1000, false)

	proofHex, err := proof.GenerateSendProof(
		senderKP.PrivKeyHex, senderKP.PubKeyHex, amount, participants, txBF, ctxHash,
		amountParams, balanceParams,
	)
	require.NoError(t, err)

	err = proof.VerifySendProof(proofHex[:len(proofHex)-2], participants, senderBalanceCt, amountCommitHex, balanceCommitHex, ctxHash)
	require.ErrorIs(t, err, proof.ErrInvalidProof)
}

func TestSendProofInvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "fail - bad privkey",
			fn: func() error {
				_, err := proof.GenerateSendProof("zz", "zz", 100, nil, zeroHex(32), zeroHex(32),
					proof.Params{}, proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidPrivKey,
		},
		{
			name: "fail - bad pubkey",
			fn: func() error {
				_, err := proof.GenerateSendProof(zeroHex(32), "zz", 100, nil, zeroHex(32), zeroHex(32),
					proof.Params{}, proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - bad tx blinding factor",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					[]proof.Participant{{PubKeyHex: pubKey, CiphertextHex: zeroHex(66)}},
					"bad", zeroHex(32),
					proof.Params{}, proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidBlindingFactor,
		},
		{
			name: "fail - bad ctx hash",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					[]proof.Participant{{PubKeyHex: pubKey, CiphertextHex: zeroHex(66)}},
					zeroHex(32), "bad",
					proof.Params{}, proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidContextHash,
		},
		{
			name: "fail - bad participant pubkey",
			fn: func() error {
				_, err := proof.GenerateSendProof(zeroHex(32), zeroHex(32), 100,
					[]proof.Participant{{PubKeyHex: "zz", CiphertextHex: zeroHex(66)}},
					zeroHex(32), zeroHex(32),
					proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: zeroHex(66), BlindingFactorHex: zeroHex(32)},
					proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: zeroHex(66), BlindingFactorHex: zeroHex(32)})
				return err
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - verify bad proof hex",
			fn: func() error {
				return proof.VerifySendProof("zzzz", nil, zeroHex(66), "02"+zeroHex(32), "02"+zeroHex(32), zeroHex(32))
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
