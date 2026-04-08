package builder

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

// TestSendBaseValidation verifies all validateSendBase branches through both entry points.
func TestSendBaseValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	cases := []struct {
		name    string
		base    BuildSendParams
		wantErr error
	}{
		{name: "fail - missing account", base: BuildSendParams{Destination: testDestination, IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex}, wantErr: ErrMissingAccount},
		{name: "fail - invalid account", base: BuildSendParams{Account: "notanaddress", Destination: testDestination, IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidAccount},
		{name: "fail - missing destination", base: BuildSendParams{Account: testAccount, IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex}, wantErr: ErrMissingDestination},
		{name: "fail - invalid destination", base: BuildSendParams{Account: testAccount, Destination: "notanaddress", IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidDestination},
		{name: "fail - self send", base: BuildSendParams{Account: testAccount, Destination: testAccount, IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex}, wantErr: ErrSelfSend},
		{name: "fail - missing issuance ID", base: BuildSendParams{Account: testAccount, Destination: testDestination, Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex}, wantErr: ErrMissingIssuanceID},
		{name: "fail - invalid issuance ID (not hex)", base: BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: strings.Repeat("GG", 24), Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - invalid issuance ID (wrong length)", base: BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: "aabb", Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidIssuanceID},
		{name: "fail - zero amount", base: BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: testIssuanceID, Amount: 0, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex}, wantErr: ErrZeroAmount},
		{name: "fail - missing sender priv key", base: BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: testIssuanceID, Amount: 1, SenderPubKey: kp.PubKeyHex}, wantErr: ErrMissingSenderKey},
		{name: "fail - invalid sender priv key (not hex)", base: BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: strings.Repeat("ZZ", 32), SenderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidPrivKey},
		{name: "fail - invalid sender priv key (wrong length)", base: BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: "aabb", SenderPubKey: kp.PubKeyHex}, wantErr: ErrInvalidPrivKey},
		{name: "fail - missing sender pub key", base: BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: kp.PrivKeyHex}, wantErr: ErrMissingSenderKey},
		{name: "fail - invalid sender pub key (not hex)", base: BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: strings.Repeat("ZZ", 33)}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid sender pub key (wrong length)", base: BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: testIssuanceID, Amount: 1, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: "aabb"}, wantErr: ErrInvalidPubKey},
	}

	t.Run("fail - validation PrepareSend", func(t *testing.T) {
		rkp, err := elgamal.GenerateKeypair()
		require.NoError(t, err)
		ikp, err := elgamal.GenerateKeypair()
		require.NoError(t, err)

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := PrepareSend(SendParams{
					BuildSendParams:  tc.base,
					ReceiverPubKey:   rkp.PubKeyHex,
					IssuerPubKey:     ikp.PubKeyHex,
					CurrentBalance:   100,
					CurrentBalanceCt: "aa",
				})
				require.ErrorIs(t, err, tc.wantErr)
			})
		}
	})

	t.Run("fail - validation BuildSend", func(t *testing.T) {
		q := &mockQuerier{}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := BuildSend(q, tc.base)
				require.ErrorIs(t, err, tc.wantErr)
			})
		}
	})
}

func TestPrepareSend_Pass(t *testing.T) {
	const currentBalance uint64 = 1000
	const sendAmount uint64 = 500

	senderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	receiverKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	// Simulate existing balance state.
	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(currentBalance, senderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	result, err := PrepareSend(SendParams{
		BuildSendParams: BuildSendParams{
			Account:       testAccount,
			Destination:   testDestination,
			IssuanceID:    testIssuanceID,
			Amount:        sendAmount,
			SenderPrivKey: senderKP.PrivKeyHex,
			SenderPubKey:  senderKP.PubKeyHex,
		},
		ReceiverPubKey:   receiverKP.PubKeyHex,
		IssuerPubKey:     issuerKP.PubKeyHex,
		Sequence:         1,
		BalanceVersion:   0,
		CurrentBalance:   currentBalance,
		CurrentBalanceCt: balanceCt,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, transaction.ConfidentialMPTSendTx, result.TxType())

	// Transaction fields.
	require.Len(t, result.SenderEncryptedAmount, 132)
	require.Len(t, result.DestinationEncryptedAmount, 132)
	require.Len(t, result.IssuerEncryptedAmount, 132)
	require.Nil(t, result.AuditorEncryptedAmount)
	require.NotEmpty(t, result.ZKProof)
	require.Len(t, result.AmountCommitment, 66)
	require.Len(t, result.BalanceCommitment, 66)

	// Verify the composite proof cryptographically.
	ctxHash, err := proof.SendContextHash(testAccount, testIssuanceID, uint32(1), testDestination, uint32(0))
	require.NoError(t, err)

	participants := []proof.Participant{
		{PubKeyHex: senderKP.PubKeyHex, CiphertextHex: result.SenderEncryptedAmount},
		{PubKeyHex: receiverKP.PubKeyHex, CiphertextHex: result.DestinationEncryptedAmount},
		{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: result.IssuerEncryptedAmount},
	}
	err = proof.VerifySendProof(result.ZKProof, participants, balanceCt, result.AmountCommitment, result.BalanceCommitment, ctxHash)
	require.NoError(t, err)

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareSend_PassWithAuditor(t *testing.T) {
	senderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	receiverKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	auditorKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(1000, senderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	result, err := PrepareSend(SendParams{
		BuildSendParams: BuildSendParams{
			Account:       testAccount,
			Destination:   testDestination,
			IssuanceID:    testIssuanceID,
			Amount:        500,
			SenderPrivKey: senderKP.PrivKeyHex,
			SenderPubKey:  senderKP.PubKeyHex,
		},
		ReceiverPubKey:   receiverKP.PubKeyHex,
		IssuerPubKey:     issuerKP.PubKeyHex,
		AuditorPubKey:    auditorKP.PubKeyHex,
		Sequence:         1,
		CurrentBalance:   1000,
		CurrentBalanceCt: balanceCt,
	})
	require.NoError(t, err)
	require.NotNil(t, result.AuditorEncryptedAmount)
	require.Len(t, *result.AuditorEncryptedAmount, 132)

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareSend_PassWithCredentialIDs(t *testing.T) {
	senderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	receiverKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(1000, senderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	credIDs := []string{"A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2", "B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3"}
	result, err := PrepareSend(SendParams{
		BuildSendParams: BuildSendParams{
			Account:       testAccount,
			Destination:   testDestination,
			IssuanceID:    testIssuanceID,
			Amount:        100,
			SenderPrivKey: senderKP.PrivKeyHex,
			SenderPubKey:  senderKP.PubKeyHex,
			CredentialIDs: credIDs,
		},
		ReceiverPubKey:   receiverKP.PubKeyHex,
		IssuerPubKey:     issuerKP.PubKeyHex,
		Sequence:         1,
		CurrentBalance:   1000,
		CurrentBalanceCt: balanceCt,
	})
	require.NoError(t, err)
	require.Equal(t, credIDs, []string(result.CredentialIDs))

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

func TestPrepareSend_FailValidation(t *testing.T) {
	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	ikp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	rkp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	validCt, err := elgamal.Encrypt(100, kp.PubKeyHex, bf)
	require.NoError(t, err)

	validBase := BuildSendParams{
		Account:       testAccount,
		Destination:   testDestination,
		IssuanceID:    testIssuanceID,
		Amount:        1,
		SenderPrivKey: kp.PrivKeyHex,
		SenderPubKey:  kp.PubKeyHex,
	}

	tests := []struct {
		name    string
		params  SendParams
		wantErr error
	}{
		{name: "fail - missing receiver key", params: SendParams{BuildSendParams: validBase, IssuerPubKey: ikp.PubKeyHex, CurrentBalanceCt: validCt, CurrentBalance: 100}, wantErr: ErrMissingReceiverKey},
		{name: "fail - invalid receiver pub key (not hex)", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: strings.Repeat("ZZ", 33), IssuerPubKey: ikp.PubKeyHex, CurrentBalanceCt: validCt, CurrentBalance: 100}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid receiver pub key (wrong length)", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: "aabb", IssuerPubKey: ikp.PubKeyHex, CurrentBalanceCt: validCt, CurrentBalance: 100}, wantErr: ErrInvalidPubKey},
		{name: "fail - missing issuer key", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: rkp.PubKeyHex, CurrentBalanceCt: validCt, CurrentBalance: 100}, wantErr: ErrMissingIssuerKey},
		{name: "fail - invalid issuer pub key (not hex)", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: rkp.PubKeyHex, IssuerPubKey: strings.Repeat("ZZ", 33), CurrentBalanceCt: validCt, CurrentBalance: 100}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid issuer pub key (wrong length)", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: rkp.PubKeyHex, IssuerPubKey: "aabb", CurrentBalanceCt: validCt, CurrentBalance: 100}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid auditor pub key (not hex)", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: rkp.PubKeyHex, IssuerPubKey: ikp.PubKeyHex, AuditorPubKey: strings.Repeat("ZZ", 33), CurrentBalanceCt: validCt, CurrentBalance: 100}, wantErr: ErrInvalidPubKey},
		{name: "fail - invalid auditor pub key (wrong length)", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: rkp.PubKeyHex, IssuerPubKey: ikp.PubKeyHex, AuditorPubKey: "aabb", CurrentBalanceCt: validCt, CurrentBalance: 100}, wantErr: ErrInvalidPubKey},
		{name: "fail - missing sender state", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: rkp.PubKeyHex, IssuerPubKey: ikp.PubKeyHex}, wantErr: ErrMissingSenderState},
		{name: "fail - invalid ciphertext (not hex)", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: rkp.PubKeyHex, IssuerPubKey: ikp.PubKeyHex, CurrentBalanceCt: strings.Repeat("ZZ", 66), CurrentBalance: 100}, wantErr: ErrInvalidCiphertext},
		{name: "fail - invalid ciphertext (wrong length)", params: SendParams{BuildSendParams: validBase, ReceiverPubKey: rkp.PubKeyHex, IssuerPubKey: ikp.PubKeyHex, CurrentBalanceCt: "aabb", CurrentBalance: 100}, wantErr: ErrInvalidCiphertext},
		{
			name: "fail - insufficient balance",
			params: SendParams{
				BuildSendParams:  BuildSendParams{Account: testAccount, Destination: testDestination, IssuanceID: testIssuanceID, Amount: 200, SenderPrivKey: kp.PrivKeyHex, SenderPubKey: kp.PubKeyHex},
				ReceiverPubKey:   rkp.PubKeyHex,
				IssuerPubKey:     ikp.PubKeyHex,
				CurrentBalance:   100,
				CurrentBalanceCt: validCt,
			},
			wantErr: ErrInsufficientBalance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PrepareSend(tt.params)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestBuildSend_Pass(t *testing.T) {
	senderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	receiverKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	const currentBalance uint64 = 1000
	const sendAmount uint64 = 300

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	senderBalanceCt, err := elgamal.Encrypt(currentBalance, senderKP.PubKeyHex, bf)
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	senderMPTIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)
	receiverMPTIndex, err := xrplhash.MPToken(testIssuanceID, testDestination)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 8,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex:    buildIssuanceEntry(issuerKP.PubKeyHex, ""),
			senderMPTIndex:   buildMPTokenEntry(senderKP.PubKeyHex, senderBalanceCt, 2, ""),
			receiverMPTIndex: buildMPTokenEntry(receiverKP.PubKeyHex, "", 0, ""),
		},
	}

	result, err := BuildSend(q, BuildSendParams{
		Account:       testAccount,
		Destination:   testDestination,
		IssuanceID:    testIssuanceID,
		Amount:        sendAmount,
		SenderPrivKey: senderKP.PrivKeyHex,
		SenderPubKey:  senderKP.PubKeyHex,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, uint32(8), result.Sequence)
	require.NotEmpty(t, result.ZKProof)
}

func TestBuildSend_FailReceiverNotOptedIn(t *testing.T) {
	senderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	const currentBalance uint64 = 1000

	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	senderBalanceCt, err := elgamal.Encrypt(currentBalance, senderKP.PubKeyHex, bf)
	require.NoError(t, err)

	issuanceIndex, err := xrplhash.MPTokenIssuance(testIssuanceID)
	require.NoError(t, err)
	senderMPTIndex, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	q := &mockQuerier{
		accountSeq: 1,
		entries: map[string]ledgerentries.FlatLedgerObject{
			issuanceIndex:  buildIssuanceEntry(issuerKP.PubKeyHex, ""),
			senderMPTIndex: buildMPTokenEntry(senderKP.PubKeyHex, senderBalanceCt, 0, ""),
		},
	}

	_, err = BuildSend(q, BuildSendParams{
		Account:       testAccount,
		Destination:   testDestination,
		IssuanceID:    testIssuanceID,
		Amount:        100,
		SenderPrivKey: senderKP.PrivKeyHex,
		SenderPubKey:  senderKP.PubKeyHex,
	})
	require.ErrorIs(t, err, ErrReceiverNotOptedIn)
}
