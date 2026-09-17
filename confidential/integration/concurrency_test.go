//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

const (
	concurrencyFunding uint64 = 30
	concurrencySend    uint64 = 9
)

// testIntegrationConfidentialMPTStaleBalance covers what happens when a confidential balance
// moves between building a proof and submitting it. XLS-96 binds a send proof to the
// spending ciphertext and its version (section A.7), so any transaction that rewrites the
// spending balance invalidates every proof already built against it. That is the failure a
// wallet with two transactions in flight actually hits, and the builder's
// requireCurrentBalanceVersion exists to catch it before a fee is spent.
//
// The stale send here spends a ticket rather than the account sequence. The merge that
// invalidates it spends the account sequence, so the ticket is still unspent at resubmission
// and the proof's nonce binding still holds: the balance version is then the only thing that
// changed, and tecBAD_PROOF can only be attributed to it.
func testIntegrationConfidentialMPTStaleBalance(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	holder := runner.GetWallet(1)
	receiver := runner.GetWallet(2)

	config := issuanceConfig{issuerKey: generateKey(t)}
	holderKey := generateKey(t)
	receiverKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, holder, issuanceID)
	authorizeHolder(t, runner, issuer, receiver, issuanceID)
	fundHolder(t, runner, issuer, holder, issuanceID, concurrencyFunding)

	convert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        concurrencyFunding,
		HolderPrivKey: holderKey.PrivKeyHex,
		HolderPubKey:  holderKey.PubKeyHex,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, convert.Flatten(), holder)

	merge, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    holder.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, merge.Flatten(), holder)

	optIn, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        0,
		HolderPrivKey: receiverKey.PrivKeyHex,
		HolderPubKey:  receiverKey.PubKeyHex,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, optIn.Flatten(), receiver)

	ticketSequence := createTicket(t, runner, client, holder)
	staleSend, err := builder.BuildSend(client, builder.BuildSendParams{
		TxOptions:     builder.TxOptions{TicketSequence: ticketSequence},
		Account:       holder.GetAddress().String(),
		Destination:   receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        concurrencySend,
		SenderPrivKey: holderKey.PrivKeyHex,
		SenderPubKey:  holderKey.PubKeyHex,
		BalanceRange:  exactRange(concurrencyFunding),
	})
	require.NoError(t, err)

	// A merge with nothing in the inbox still applies and still advances the version
	// (section 9.3), which is what a wallet merging defensively before every send relies on.
	// Here it is also the cheapest way to invalidate the proof built above.
	versionBefore := getMPToken(t, client, holder.GetAddress()).ConfidentialBalanceVersion
	noOpMerge, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    holder.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, noOpMerge.Flatten(), holder)

	afterMerge := getMPToken(t, client, holder.GetAddress())
	require.Equal(t, versionBefore+1, afterMerge.ConfidentialBalanceVersion)
	// The merge moved nothing, so the balance the stale proof commits to is still the
	// balance the holder has. Only the version moved.
	require.Equal(t, concurrencyFunding, decryptBalance(t, afterMerge.ConfidentialBalanceSpending, holderKey.PrivKeyHex))

	submitExpectingResult(t, client, staleSend.Flatten(), holder, "tecBAD_PROOF")

	// The builder reads the open ledger so it can reject a stale build without paying for
	// it. Reaching that branch needs a transaction the open ledger has and the validated
	// ledger does not, so this submits one without waiting. A localnet closes ledgers fast
	// enough that the merge may already be validated by the time the build reads, in which
	// case the builder correctly pins the newer version and succeeds. Either outcome is
	// right, so only a wrong error fails the test.
	racingMerge, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    holder.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.NoError(t, err)
	flat := racingMerge.Flatten()
	require.NoError(t, client.Autofill(&flat))
	blob, _, err := holder.Sign(flat)
	require.NoError(t, err)
	_, err = client.SubmitTxBlob(blob, false)
	require.NoError(t, err)

	_, err = builder.BuildSend(client, builder.BuildSendParams{
		Account:       holder.GetAddress().String(),
		Destination:   receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        concurrencySend,
		SenderPrivKey: holderKey.PrivKeyHex,
		SenderPubKey:  holderKey.PubKeyHex,
		BalanceRange:  exactRange(concurrencyFunding),
	})
	if err != nil {
		require.ErrorIs(t, err, builder.ErrStaleBalanceVersion)
		return
	}
	t.Log("merge validated before the build read the open ledger, so the staleness branch was skipped")
}

// testIntegrationConfidentialMPTDeletionBlocked covers the MPToken deletion rule confidential
// balances add. rippled blocks the delete with tecHAS_OBLIGATIONS when the issuance still
// reports a ConfidentialOutstandingAmount and the holder carries confidential fields at all,
// deliberately without checking whether that holder's own balance is zero: the balances are
// ciphertexts, so it cannot tell. A holder that only ever opted in is therefore blocked by
// somebody else's balance, which is the case worth pinning.
func testIntegrationConfidentialMPTDeletionBlocked(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	holder := runner.GetWallet(1)
	optedIn := runner.GetWallet(2)

	config := issuanceConfig{issuerKey: generateKey(t)}
	holderKey := generateKey(t)
	optedInKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, holder, issuanceID)
	authorizeHolder(t, runner, issuer, optedIn, issuanceID)
	fundHolder(t, runner, issuer, holder, issuanceID, concurrencyFunding)

	// One holder puts the issuance's confidential outstanding amount above zero.
	convert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        concurrencyFunding,
		HolderPrivKey: holderKey.PrivKeyHex,
		HolderPubKey:  holderKey.PubKeyHex,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, convert.Flatten(), holder)

	// The other only registers a key: no public balance, no confidential balance, but the
	// confidential fields now exist on its MPToken.
	optIn, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       optedIn.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        0,
		HolderPrivKey: optedInKey.PrivKeyHex,
		HolderPubKey:  optedInKey.PubKeyHex,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, optIn.Flatten(), optedIn)

	token := getMPToken(t, client, optedIn.GetAddress())
	require.Empty(t, token.MPTAmount)
	require.Equal(t, uint64(0), decryptBalance(t, token.ConfidentialBalanceSpending, optedInKey.PrivKeyHex))
	require.NotEmpty(t, getIssuance(t, client, issuer.GetAddress()).ConfidentialOutstandingAmount)

	unauthorize := transaction.MPTokenAuthorize{
		BaseTx:            transaction.BaseTx{Account: optedIn.GetAddress()},
		MPTokenIssuanceID: issuanceID,
	}
	unauthorize.SetMPTUnauthorizeFlag()
	submitExpectingResult(t, client, unauthorize.Flatten(), optedIn, "tecHAS_OBLIGATIONS")
}

func TestIntegrationConfidentialMPTDeletionBlocked_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTDeletionBlocked(t, client)
}

func TestIntegrationConfidentialMPTDeletionBlocked_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTDeletionBlocked(t, client)
}

func TestIntegrationConfidentialMPTStaleBalance_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTStaleBalance(t, client)
}

func TestIntegrationConfidentialMPTStaleBalance_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTStaleBalance(t, client)
}
