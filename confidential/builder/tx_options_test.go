package builder

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/proof"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

// requireTicketOptions pins that a builder placed a ticket sequence and a delegate on the
// transaction, and left the account sequence at zero, which XRPL requires alongside a ticket.
func requireTicketOptions(t *testing.T, tx transaction.BaseTx, ticketSequence uint32, delegate string) {
	t.Helper()
	require.Zero(t, tx.Sequence)
	require.Equal(t, ticketSequence, tx.TicketSequence)
	require.Equal(t, delegate, tx.Delegate.String())
}

func requireSequenceOptions(t *testing.T, tx transaction.BaseTx, sequence uint32, delegate string) {
	t.Helper()
	require.Equal(t, sequence, tx.Sequence)
	require.Zero(t, tx.TicketSequence)
	require.Equal(t, delegate, tx.Delegate.String())
}

// TestTxOptionsValidation covers the option combinations a confidential transaction cannot
// carry. The delegate address conditions repeat what BaseTx.Validate reports, because a
// builder must reach them before spending a ledger query or a proof on a doomed
// transaction. The nonce conflict and the delegability of the type have no preflight
// counterpart, so these cases are the only coverage they get. A Ticket passes for every
// type: xrpld verifies a proof committed to the sequence proxy, so no type is refused one.
func TestTxOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		options TxOptions
		txType  transaction.TxType
		wantErr error
	}{
		{
			name:    "fail - sequence and ticket sequence",
			options: TxOptions{Sequence: 1, TicketSequence: 2},
			txType:  transaction.ConfidentialMPTSendTx,
			wantErr: ErrConflictingNonce,
		},
		{
			name:    "fail - invalid delegate",
			options: TxOptions{Sequence: 1, Delegate: "not-an-address"},
			txType:  transaction.ConfidentialMPTSendTx,
			wantErr: transaction.ErrInvalidDelegate,
		},
		{
			name:    "fail - delegate is ACCOUNT_ZERO",
			options: TxOptions{Sequence: 1, Delegate: zeroClassicAccount},
			txType:  transaction.ConfidentialMPTSendTx,
			wantErr: transaction.ErrDelegateZero,
		},
		{
			name:    "fail - tagged delegate X-address",
			options: TxOptions{Sequence: 1, Delegate: taggedXAddressOf(t, testDelegate, 42)},
			txType:  transaction.ConfidentialMPTSendTx,
			wantErr: transaction.ErrDelegateTagNotAllowed,
		},
		{
			name:    "fail - delegate equals account",
			options: TxOptions{Sequence: 1, Delegate: testAccount},
			txType:  transaction.ConfidentialMPTSendTx,
			wantErr: transaction.ErrDelegateAccountConflict,
		},
		{
			name:    "fail - delegate equals account in X-address form",
			options: TxOptions{Sequence: 1, Delegate: xAddressOf(t, testAccount)},
			txType:  transaction.ConfidentialMPTSendTx,
			wantErr: transaction.ErrDelegateAccountConflict,
		},
		{
			name:    "fail - delegate on a non-delegable type",
			options: TxOptions{Sequence: 1, Delegate: testDelegate},
			txType:  transaction.ConfidentialMPTConvertTx,
			wantErr: ErrDelegateNotAllowed,
		},
		{
			name:    "pass - ticket on send",
			options: TxOptions{TicketSequence: 1},
			txType:  transaction.ConfidentialMPTSendTx,
		},
		{
			name:    "pass - ticket on convert back",
			options: TxOptions{TicketSequence: 1},
			txType:  transaction.ConfidentialMPTConvertBackTx,
		},
		{
			name:    "pass - ticket on clawback",
			options: TxOptions{TicketSequence: 1},
			txType:  transaction.ConfidentialMPTClawbackTx,
		},
		{
			name:    "pass - ticket on convert",
			options: TxOptions{TicketSequence: 1},
			txType:  transaction.ConfidentialMPTConvertTx,
		},
		{
			name:    "pass - ticket on merge inbox",
			options: TxOptions{TicketSequence: 1},
			txType:  transaction.ConfidentialMPTMergeInboxTx,
		},
		{
			name:    "pass - delegate",
			options: TxOptions{Sequence: 1, Delegate: testDelegate},
			txType:  transaction.ConfidentialMPTSendTx,
		},
		{
			name:    "pass - untagged delegate X-address",
			options: TxOptions{Sequence: 1, Delegate: xAddressOf(t, testDelegate)},
			txType:  transaction.ConfidentialMPTSendTx,
		},
		{
			name:   "pass - no nonce and no delegate",
			txType: transaction.ConfidentialMPTConvertTx,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.options.validate(testAccount, test.txType)
			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

// TestPrepareSendTicketBindsProofToTicketSequence pins that a ticketed send commits its proof
// to the ticket sequence. xrpld hashes the sequence proxy, which is the ticket whenever the
// transaction spends one, so a proof bound to anything else would be rejected with
// tecBAD_PROOF. Nothing refuses a Ticket here: the stale-proof hazard the package documents is
// a per-MPToken collision between transactions in flight, which no single build can observe.
func TestPrepareSendTicketBindsProofToTicketSequence(t *testing.T) {
	const currentBalance uint64 = 1000
	const sendAmount uint64 = 500
	const ticketSequence uint32 = 11

	senderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	receiverKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(currentBalance, senderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	result, err := PrepareSend(SendParams{
		BuildSendParams: BuildSendParams{
			TxOptions:     TxOptions{TicketSequence: ticketSequence, Delegate: testDelegate},
			Account:       testAccount,
			Destination:   testDestination,
			IssuanceID:    testIssuanceID,
			Amount:        sendAmount,
			SenderPrivKey: senderKP.PrivKeyHex,
			SenderPubKey:  senderKP.PubKeyHex,
		},
		ReceiverPubKey:   receiverKP.PubKeyHex,
		IssuerPubKey:     issuerKP.PubKeyHex,
		BalanceVersion:   0,
		CurrentBalance:   currentBalance,
		CurrentBalanceCt: balanceCt,
	})
	require.NoError(t, err)
	requireTicketOptions(t, result.BaseTx, ticketSequence, testDelegate)

	ctxHash, err := proof.SendContextHash(testAccount, testIssuanceID, ticketSequence, testDestination, uint32(0))
	require.NoError(t, err)
	participants := []proof.Participant{
		{PubKeyHex: senderKP.PubKeyHex, CiphertextHex: result.SenderEncryptedAmount},
		{PubKeyHex: receiverKP.PubKeyHex, CiphertextHex: result.DestinationEncryptedAmount},
		{PubKeyHex: issuerKP.PubKeyHex, CiphertextHex: result.IssuerEncryptedAmount},
	}
	require.NoError(t, proof.VerifySendProof(result.ZKProof, participants, balanceCt, result.AmountCommitment, result.BalanceCommitment, ctxHash))

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

// TestPrepareConvertBackTicketBindsProofToTicketSequence pins the same binding for the other
// helper whose proof commits to the submitter's own ConfidentialBalanceVersion.
func TestPrepareConvertBackTicketBindsProofToTicketSequence(t *testing.T) {
	const currentBalance uint64 = 1000
	const withdrawAmount uint64 = 400
	const ticketSequence uint32 = 7

	holderKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	issuerKP, err := elgamal.GenerateKeypair()
	require.NoError(t, err)

	balanceBF, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(currentBalance, holderKP.PubKeyHex, balanceBF)
	require.NoError(t, err)

	result, err := PrepareConvertBack(ConvertBackParams{
		BuildConvertBackParams: BuildConvertBackParams{
			TxOptions:     TxOptions{TicketSequence: ticketSequence, Delegate: testDelegate},
			Account:       testAccount,
			IssuanceID:    testIssuanceID,
			Amount:        withdrawAmount,
			HolderPrivKey: holderKP.PrivKeyHex,
			HolderPubKey:  holderKP.PubKeyHex,
		},
		IssuerPubKey:     issuerKP.PubKeyHex,
		BalanceVersion:   0,
		CurrentBalance:   currentBalance,
		CurrentBalanceCt: balanceCt,
	})
	require.NoError(t, err)
	requireTicketOptions(t, result.BaseTx, ticketSequence, testDelegate)

	ctxHash, err := proof.ConvertBackContextHash(testAccount, testIssuanceID, ticketSequence, uint32(0))
	require.NoError(t, err)
	require.NoError(t, proof.VerifyConvertBackProof(result.ZKProof, holderKP.PubKeyHex, balanceCt, result.BalanceCommitment, withdrawAmount, ctxHash))

	ok, err := result.Validate()
	require.NoError(t, err)
	require.True(t, ok)
}

// TestTxOptionsProofSequence pins the nonce a proof context commits to. xrpld hashes the
// sequence proxy, so a transaction spending a Ticket must prove against that ticket.
func TestTxOptionsProofSequence(t *testing.T) {
	tests := []struct {
		name    string
		options TxOptions
		want    uint32
		wantErr error
	}{
		{name: "pass - sequence", options: TxOptions{Sequence: 4}, want: 4},
		{name: "pass - ticket sequence", options: TxOptions{TicketSequence: 8}, want: 8},
		{name: "fail - unresolved", wantErr: ErrMissingSequence},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sequence, err := test.options.proofSequence()
			require.Equal(t, test.want, sequence)
			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

// TestPrepareParamsUseBuildParamsTxOptions pins that each Prepare*Params type holds exactly
// one TxOptions value, the one inside its embedded Build*Params, so a promoted selector and
// a composite literal cannot disagree about which options a builder reads.
func TestPrepareParamsUseBuildParamsTxOptions(t *testing.T) {
	send := SendParams{BuildSendParams: BuildSendParams{TxOptions: TxOptions{Sequence: 1}}}
	convert := ConvertParams{BuildConvertParams: BuildConvertParams{TxOptions: TxOptions{Sequence: 2}}}
	convertBack := ConvertBackParams{BuildConvertBackParams: BuildConvertBackParams{TxOptions: TxOptions{Sequence: 3}}}
	clawback := ClawbackParams{BuildClawbackParams: BuildClawbackParams{TxOptions: TxOptions{Sequence: 4}}}
	mergeInbox := MergeInboxParams{BuildMergeInboxParams: BuildMergeInboxParams{TxOptions: TxOptions{Sequence: 5}}}

	send.Sequence = 11
	convert.Sequence = 12
	convertBack.Sequence = 13
	clawback.Sequence = 14
	mergeInbox.Sequence = 15

	require.Equal(t, uint32(11), send.BuildSendParams.TxOptions.Sequence)
	require.Equal(t, uint32(12), convert.BuildConvertParams.TxOptions.Sequence)
	require.Equal(t, uint32(13), convertBack.BuildConvertBackParams.TxOptions.Sequence)
	require.Equal(t, uint32(14), clawback.BuildClawbackParams.TxOptions.Sequence)
	require.Equal(t, uint32(15), mergeInbox.BuildMergeInboxParams.TxOptions.Sequence)
}

// TestBuildSkipsSequenceLookupWhenNonceIsGiven pins that a caller-supplied nonce replaces the
// account query entirely, and that the state reads still bind to one validated ledger: the
// first selects it, and the second requests the hash the first reported.
func TestBuildSkipsSequenceLookupWhenNonceIsGiven(t *testing.T) {
	tests := []struct {
		name               string
		options            TxOptions
		wantSequence       uint32
		wantTicketSequence uint32
	}{
		{
			name:               "ticket sequence",
			options:            TxOptions{TicketSequence: 9, Delegate: testDelegate},
			wantTicketSequence: 9,
		},
		{
			name:         "account sequence",
			options:      TxOptions{Sequence: 13, Delegate: xAddressOf(t, testDelegate)},
			wantSequence: 13,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := newMergeInboxLedger(t, 42)
			q.accountErr = errors.New("the account sequence must not be read")

			tx, err := BuildMergeInbox(q, BuildMergeInboxParams{
				TxOptions:  test.options,
				Account:    testAccount,
				IssuanceID: testIssuanceID,
			})

			require.NoError(t, err)
			require.Equal(t, test.wantSequence, tx.Sequence)
			require.Equal(t, test.wantTicketSequence, tx.TicketSequence)
			require.Equal(t, test.options.Delegate, tx.Delegate.String())
			require.Empty(t, q.accountRequests)
			require.Len(t, q.entryRequests, 2)
			require.Equal(t, common.Validated, q.entryRequests[0].LedgerIndex)
			require.Equal(t, mockLedgerHash, q.entryRequests[1].LedgerHash)
		})
	}
}

// TestPrepareMergeInboxAllowsZeroNonce pins that the one proof-free builder still accepts a
// transaction with no nonce, which a later autofill resolves.
func TestPrepareMergeInboxAllowsZeroNonce(t *testing.T) {
	tx, err := PrepareMergeInbox(MergeInboxParams{
		BuildMergeInboxParams: BuildMergeInboxParams{
			Account:    testAccount,
			IssuanceID: testIssuanceID,
		},
	})

	require.NoError(t, err)
	require.Zero(t, tx.Sequence)
	require.Zero(t, tx.TicketSequence)
}

// TestPrepareRejectsConflictingNonce pins that the guard runs on the offline path too, where
// no ledger read would surface it. Both routes into it are covered: the proof-free helper
// calls validate directly, and a proof-bearing one reaches it through validateForProof, which
// must refuse before resolving a nonce it would otherwise commit to.
func TestPrepareRejectsConflictingNonce(t *testing.T) {
	const currentBalance uint64 = 1000

	kp, err := elgamal.GenerateKeypair()
	require.NoError(t, err)
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	balanceCt, err := elgamal.Encrypt(currentBalance, kp.PubKeyHex, bf)
	require.NoError(t, err)
	conflicting := TxOptions{Sequence: 1, TicketSequence: 2}

	tests := []struct {
		name    string
		prepare func() error
	}{
		{
			name: "merge inbox",
			prepare: func() error {
				_, err := PrepareMergeInbox(MergeInboxParams{
					BuildMergeInboxParams: BuildMergeInboxParams{
						TxOptions:  conflicting,
						Account:    testAccount,
						IssuanceID: testIssuanceID,
					},
				})
				return err
			},
		},
		{
			name: "convert back",
			prepare: func() error {
				_, err := PrepareConvertBack(ConvertBackParams{
					BuildConvertBackParams: BuildConvertBackParams{
						TxOptions:     conflicting,
						Account:       testAccount,
						IssuanceID:    testIssuanceID,
						Amount:        500,
						HolderPrivKey: kp.PrivKeyHex,
						HolderPubKey:  kp.PubKeyHex,
					},
					IssuerPubKey:     kp.PubKeyHex,
					CurrentBalance:   currentBalance,
					CurrentBalanceCt: balanceCt,
				})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorIs(t, test.prepare(), ErrConflictingNonce)
		})
	}
}
