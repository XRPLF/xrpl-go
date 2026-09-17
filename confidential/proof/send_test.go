//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package proof_test

import (
	"fmt"
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

	// Balance proof parameters.
	balanceParams = proof.Params{
		CommitmentHex:     balanceCommitHex,
		Amount:            senderBalance,
		CiphertextHex:     senderBalanceCt,
		BlindingFactorHex: balanceBF,
	}

	return
}

func testSendParticipants(pubKey, ciphertext string) []proof.Participant {
	participant := proof.Participant{PubKeyHex: pubKey, CiphertextHex: ciphertext}
	return []proof.Participant{participant, participant, participant}
}

func TestGenerateAndVerifySendProof(t *testing.T) {
	tests := []struct {
		name          string
		sendAmount    uint64
		senderBalance uint64
		withAuditor   bool
	}{
		{"pass - zero amount", 0, 1000, false},
		{"pass - 3 participants", 500, 1000, false},
		{"pass - 4 participants with auditor", 500, 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			senderKP, participants, txBF, ctxHash, balanceParams, senderBalanceCt, amountCommitHex, balanceCommitHex := setupSendProofScenario(t, tt.sendAmount, tt.senderBalance, tt.withAuditor)

			proofHex, err := proof.GenerateSendProof(
				senderKP.PrivKeyHex, senderKP.PubKeyHex, tt.sendAmount, participants, txBF, ctxHash,
				amountCommitHex, balanceParams,
			)
			require.NoError(t, err)
			require.NotEmpty(t, proofHex)

			err = proof.VerifySendProof(proofHex, participants, senderBalanceCt, amountCommitHex, balanceCommitHex, ctxHash)
			require.NoError(t, err)
		})
	}
}

// TestGenerateSendProofRejectsMismatchedPrivateKey pins the self-verification step. The
// native generator returns no error for a private key that does not match the public key,
// so only the verification GenerateSendProof runs on its own output catches it.
func TestGenerateSendProofRejectsMismatchedPrivateKey(t *testing.T) {
	senderKP, participants, txBF, ctxHash, balanceParams, _, amountCommitHex, _ := setupSendProofScenario(t, 500, 1000, false)

	wrongKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	_, err = proof.GenerateSendProof(
		wrongKP.PrivKeyHex, senderKP.PubKeyHex, 500, participants, txBF, ctxHash,
		amountCommitHex, balanceParams,
	)
	require.ErrorIs(t, err, proof.ErrProofGenerationFailed)
}

func TestVerifySendProofRejectsChangedStatement(t *testing.T) {
	senderKP, participants, txBF, ctxHash, balanceParams, senderBalanceCt, amountCommitHex, balanceCommitHex := setupSendProofScenario(t, 500, 1000, true)
	proofHex, err := proof.GenerateSendProof(
		senderKP.PrivKeyHex, senderKP.PubKeyHex, 500, participants, txBF, ctxHash,
		amountCommitHex, balanceParams,
	)
	require.NoError(t, err)

	wrongKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	wrongCtxHash, err := proof.SendContextHash(testAccount, testIssuanceID, 2, testDest, 0)
	require.NoError(t, err)

	senderDestSwapped := append([]proof.Participant(nil), participants...)
	senderDestSwapped[0], senderDestSwapped[1] = senderDestSwapped[1], senderDestSwapped[0]

	destIssuerSwapped := append([]proof.Participant(nil), participants...)
	destIssuerSwapped[1], destIssuerSwapped[2] = destIssuerSwapped[2], destIssuerSwapped[1]

	wrongFirstKey := append([]proof.Participant(nil), participants...)
	wrongFirstKey[0].PubKeyHex = wrongKP.PubKeyHex

	// Each case is the verified statement with exactly one component replaced. Naming every
	// component keeps the two commitments distinguishable, since both are hex strings.
	type changedStatement struct {
		name              string
		proofHex          string
		participants      []proof.Participant
		balanceCiphertext string
		amountCommitment  string
		balanceCommitment string
		contextHash       string
	}

	tests := []changedStatement{
		{
			name: "sender and destination order", proofHex: proofHex, participants: senderDestSwapped,
			balanceCiphertext: senderBalanceCt, amountCommitment: amountCommitHex,
			balanceCommitment: balanceCommitHex, contextHash: ctxHash,
		},
		{
			name: "destination and issuer order", proofHex: proofHex, participants: destIssuerSwapped,
			balanceCiphertext: senderBalanceCt, amountCommitment: amountCommitHex,
			balanceCommitment: balanceCommitHex, contextHash: ctxHash,
		},
		{
			name: "first public key", proofHex: proofHex, participants: wrongFirstKey,
			balanceCiphertext: senderBalanceCt, amountCommitment: amountCommitHex,
			balanceCommitment: balanceCommitHex, contextHash: ctxHash,
		},
		{
			name: "sender balance ciphertext", proofHex: proofHex, participants: participants,
			balanceCiphertext: participants[0].CiphertextHex, amountCommitment: amountCommitHex,
			balanceCommitment: balanceCommitHex, contextHash: ctxHash,
		},
		{
			name: "amount commitment", proofHex: proofHex, participants: participants,
			balanceCiphertext: senderBalanceCt, amountCommitment: balanceCommitHex,
			balanceCommitment: balanceCommitHex, contextHash: ctxHash,
		},
		{
			name: "balance commitment", proofHex: proofHex, participants: participants,
			balanceCiphertext: senderBalanceCt, amountCommitment: amountCommitHex,
			balanceCommitment: amountCommitHex, contextHash: ctxHash,
		},
		{
			name: "context hash", proofHex: proofHex, participants: participants,
			balanceCiphertext: senderBalanceCt, amountCommitment: amountCommitHex,
			balanceCommitment: balanceCommitHex, contextHash: wrongCtxHash,
		},
		{
			name: "proof bytes", proofHex: changedHex(proofHex), participants: participants,
			balanceCiphertext: senderBalanceCt, amountCommitment: amountCommitHex,
			balanceCommitment: balanceCommitHex, contextHash: ctxHash,
		},
	}

	for i := range participants {
		changedParticipants := append([]proof.Participant(nil), participants...)
		changedParticipants[i].CiphertextHex = senderBalanceCt
		tests = append(tests, changedStatement{
			name: fmt.Sprintf("participant %d ciphertext", i), proofHex: proofHex,
			participants: changedParticipants, balanceCiphertext: senderBalanceCt,
			amountCommitment: amountCommitHex, balanceCommitment: balanceCommitHex, contextHash: ctxHash,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := proof.VerifySendProof(tt.proofHex, tt.participants, tt.balanceCiphertext,
				tt.amountCommitment, tt.balanceCommitment, tt.contextHash)
			require.ErrorIs(t, err, proof.ErrProofVerificationFailed)
		})
	}
}

func TestVerifySendProofRejectsWrongLength(t *testing.T) {
	const amount = 500
	senderKP, participants, txBF, ctxHash, balanceParams, senderBalanceCt, amountCommitHex, balanceCommitHex := setupSendProofScenario(t, amount, 1000, false)

	proofHex, err := proof.GenerateSendProof(
		senderKP.PrivKeyHex, senderKP.PubKeyHex, amount, participants, txBF, ctxHash,
		amountCommitHex, balanceParams,
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
				_, err := proof.GenerateSendProof("zz", "zz", 100, nil, zeroHex(32), zeroHex(32), "", proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidPrivKey,
		},
		{
			name: "fail - bad pubkey",
			fn: func() error {
				_, err := proof.GenerateSendProof(zeroHex(32), "zz", 100, nil, zeroHex(32), zeroHex(32), "", proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - bad tx blinding factor",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					testSendParticipants(pubKey, zeroHex(66)),
					"bad", zeroHex(32), "", proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidBlindingFactor,
		},
		{
			name: "fail - bad ctx hash",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					testSendParticipants(pubKey, zeroHex(66)),
					zeroHex(32), "bad", "", proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidContextHash,
		},
		{
			name: "fail - no participants",
			fn: func() error {
				_, err := proof.GenerateSendProof(zeroHex(32), "02"+zeroHex(32), 100, nil,
					zeroHex(32), zeroHex(32), "", proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidParticipantCount,
		},
		{
			name: "fail - two participants",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				participant := proof.Participant{PubKeyHex: pubKey, CiphertextHex: zeroHex(66)}
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					[]proof.Participant{participant, participant}, zeroHex(32), zeroHex(32), "02"+zeroHex(32), proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidParticipantCount,
		},
		{
			name: "fail - five participants",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				participant := proof.Participant{PubKeyHex: pubKey, CiphertextHex: zeroHex(66)}
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					[]proof.Participant{participant, participant, participant, participant, participant},
					zeroHex(32), zeroHex(32), "02"+zeroHex(32), proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidParticipantCount,
		},
		{
			name: "fail - bad participant pubkey",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					testSendParticipants("zz", zeroHex(66)),
					zeroHex(32), zeroHex(32), "02"+zeroHex(32),
					proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: zeroHex(66), BlindingFactorHex: zeroHex(32)})
				return err
			},
			wantErr: proof.ErrInvalidPubKey,
		},
		{
			name: "fail - bad participant ciphertext",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					testSendParticipants(pubKey, "bad"),
					zeroHex(32), zeroHex(32), "02"+zeroHex(32), proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
		{
			name: "fail - bad amount commitment",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					testSendParticipants(pubKey, zeroHex(66)),
					zeroHex(32), zeroHex(32), "bad", proof.Params{})
				return err
			},
			wantErr: proof.ErrInvalidCommitment,
		},
		{
			name: "fail - bad balance commitment",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					testSendParticipants(pubKey, zeroHex(66)),
					zeroHex(32), zeroHex(32), "02"+zeroHex(32),
					proof.Params{CommitmentHex: "bad", CiphertextHex: zeroHex(66), BlindingFactorHex: zeroHex(32)})
				return err
			},
			wantErr: proof.ErrInvalidCommitment,
		},
		{
			name: "fail - bad balance ciphertext",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					testSendParticipants(pubKey, zeroHex(66)),
					zeroHex(32), zeroHex(32), "02"+zeroHex(32),
					proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: "bad", BlindingFactorHex: zeroHex(32)})
				return err
			},
			wantErr: proof.ErrInvalidCiphertext,
		},
		{
			name: "fail - bad balance blinding factor",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				_, err := proof.GenerateSendProof(zeroHex(32), pubKey, 100,
					testSendParticipants(pubKey, zeroHex(66)),
					zeroHex(32), zeroHex(32), "02"+zeroHex(32),
					proof.Params{CommitmentHex: "02" + zeroHex(32), CiphertextHex: zeroHex(66), BlindingFactorHex: "bad"})
				return err
			},
			wantErr: proof.ErrInvalidBlindingFactor,
		},
		{
			name: "fail - verify bad proof hex",
			fn: func() error {
				return proof.VerifySendProof("zzzz", nil, zeroHex(66), "02"+zeroHex(32), "02"+zeroHex(32), zeroHex(32))
			},
			wantErr: proof.ErrInvalidProof,
		},
		{
			name: "fail - verify invalid participant count",
			fn: func() error {
				pubKey := "02" + zeroHex(32)
				participant := proof.Participant{PubKeyHex: pubKey, CiphertextHex: zeroHex(66)}
				return proof.VerifySendProof(zeroHex(946), []proof.Participant{participant, participant}, zeroHex(66), "02"+zeroHex(32), "02"+zeroHex(32), zeroHex(32))
			},
			wantErr: proof.ErrInvalidParticipantCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
