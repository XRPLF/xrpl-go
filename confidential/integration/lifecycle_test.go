//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

// Amounts for the lifecycle scenario. They are distinct so that a misattributed balance
// shows up as a wrong number rather than as a coincidence.
const (
	lifecycleFunding        uint64 = 100
	lifecycleFirstConvert   uint64 = 60
	lifecycleSecondConvert  uint64 = 5
	lifecycleSend           uint64 = 20
	lifecycleConvertBack    uint64 = 3
	lifecycleDestinationTag uint32 = 7
)

// testIntegrationConfidentialMPTLifecycle walks one issuance through every confidential
// MPT transaction type except clawback, which needs an issuance created with
// LsfMPTCanClawback and so has its own scenario.
func testIntegrationConfidentialMPTLifecycle(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 4})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	sender := runner.GetWallet(1)
	receiver := runner.GetWallet(2)
	notOptedIn := runner.GetWallet(3)

	auditorKey := generateKey(t)
	config := issuanceConfig{issuerKey: generateKey(t), auditorKey: &auditorKey}
	senderKey := generateKey(t)
	receiverKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, sender, issuanceID)
	authorizeHolder(t, runner, issuer, receiver, issuanceID)
	fundHolder(t, runner, issuer, sender, issuanceID, lifecycleFunding)

	// A holder with no registered encryption key converts in the first-time form, which
	// carries the key and the Schnorr proof that the holder owns it.
	firstConvert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       sender.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        lifecycleFirstConvert,
		HolderPrivKey: senderKey.PrivKeyHex,
		HolderPubKey:  senderKey.PubKeyHex,
	})
	require.NoError(t, err)
	require.NotNil(t, firstConvert.HolderEncryptionKey)
	require.NotNil(t, firstConvert.ZKProof)
	require.NotNil(t, firstConvert.AuditorEncryptedAmount)
	submitAndWait(t, runner, firstConvert.Flatten(), sender)
	assertMirrorBalances(t, client, sender.GetAddress(), config, lifecycleFirstConvert)
	// A convert credits the inbox, not the spending balance. The mirrors above sum the two,
	// so only this distinguishes a correct credit from one posted straight to spending.
	assertSplitBalances(t, client, sender.GetAddress(), senderKey.PrivKeyHex, lifecycleFirstConvert, 0, 0)

	// Every later conversion reuses the registered key, so it sends neither the key nor
	// the registration proof again.
	secondConvert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       sender.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        lifecycleSecondConvert,
		HolderPrivKey: senderKey.PrivKeyHex,
		HolderPubKey:  senderKey.PubKeyHex,
	})
	require.NoError(t, err)
	require.Nil(t, secondConvert.HolderEncryptionKey)
	require.Nil(t, secondConvert.ZKProof)
	require.NotNil(t, secondConvert.AuditorEncryptedAmount)
	submitAndWait(t, runner, secondConvert.Flatten(), sender)

	const senderBalanceBeforeSend = lifecycleFirstConvert + lifecycleSecondConvert
	assertMirrorBalances(t, client, sender.GetAddress(), config, senderBalanceBeforeSend)

	// Converted amounts land in the inbox, so the sender merges before spending them.
	mergeSender, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    sender.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, mergeSender.Flatten(), sender)
	assertMirrorBalances(t, client, sender.GetAddress(), config, senderBalanceBeforeSend)

	// A zero-value conversion registers the receiver's key without moving any balance,
	// which is how a holder opts in to receive before holding anything.
	optIn, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        0,
		HolderPrivKey: receiverKey.PrivKeyHex,
		HolderPubKey:  receiverKey.PubKeyHex,
	})
	require.NoError(t, err)
	require.NotNil(t, optIn.HolderEncryptionKey)
	require.NotNil(t, optIn.ZKProof)
	submitAndWait(t, runner, optIn.Flatten(), receiver)

	// A holder who never registered a key cannot receive, and the builder reports that
	// before spending a fee and a sequence on it.
	_, err = builder.BuildSend(client, builder.BuildSendParams{
		Account:       sender.GetAddress().String(),
		Destination:   notOptedIn.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        1,
		SenderPrivKey: senderKey.PrivKeyHex,
		SenderPubKey:  senderKey.PubKeyHex,
		BalanceRange:  exactRange(senderBalanceBeforeSend),
	})
	require.ErrorIs(t, err, builder.ErrReceiverNotOptedIn)

	destinationTag := lifecycleDestinationTag
	send, err := builder.BuildSend(client, builder.BuildSendParams{
		Account:        sender.GetAddress().String(),
		Destination:    receiver.GetAddress().String(),
		DestinationTag: &destinationTag,
		IssuanceID:     issuanceID,
		Amount:         lifecycleSend,
		SenderPrivKey:  senderKey.PrivKeyHex,
		SenderPubKey:   senderKey.PubKeyHex,
		BalanceRange:   exactRange(senderBalanceBeforeSend),
	})
	require.NoError(t, err)
	require.Equal(t, destinationTag, *send.DestinationTag)
	require.NotNil(t, send.AuditorEncryptedAmount)
	sendResponse := submitAndWait(t, runner, send.Flatten(), sender)
	require.Equal(t, destinationTag, integration.TxFieldUint32(t, sendResponse.TxJSON, "DestinationTag"))

	const senderConfidentialBalance = senderBalanceBeforeSend - lifecycleSend
	assertMirrorBalances(t, client, sender.GetAddress(), config, senderConfidentialBalance)
	assertMirrorBalances(t, client, receiver.GetAddress(), config, lifecycleSend)
	// A send credits the destination's inbox and leaves the destination's version alone:
	// only the sender's version advances, so a receiver never invalidates its own proofs.
	assertSplitBalances(t, client, receiver.GetAddress(), receiverKey.PrivKeyHex, lifecycleSend, 0, 0)

	mergeReceiver, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    receiver.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, mergeReceiver.Flatten(), receiver)
	assertMirrorBalances(t, client, receiver.GetAddress(), config, lifecycleSend)

	convertBack, err := builder.BuildConvertBack(client, builder.BuildConvertBackParams{
		Account:       receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        lifecycleConvertBack,
		HolderPrivKey: receiverKey.PrivKeyHex,
		HolderPubKey:  receiverKey.PubKeyHex,
		BalanceRange:  exactRange(lifecycleSend),
	})
	require.NoError(t, err)
	require.NotNil(t, convertBack.AuditorEncryptedAmount)
	submitAndWait(t, runner, convertBack.Flatten(), receiver)

	// The builder decrypts the spending balance before proving, so it rejects an
	// overspend without asking the network to.
	_, err = builder.BuildSend(client, builder.BuildSendParams{
		Account:       sender.GetAddress().String(),
		Destination:   receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        senderConfidentialBalance + 1,
		SenderPrivKey: senderKey.PrivKeyHex,
		SenderPubKey:  senderKey.PubKeyHex,
		BalanceRange:  exactRange(senderConfidentialBalance),
	})
	require.ErrorIs(t, err, builder.ErrInsufficientBalance)

	const receiverConfidentialBalance = lifecycleSend - lifecycleConvertBack
	const senderPublicBalance = lifecycleFunding - senderBalanceBeforeSend

	senderToken := getMPToken(t, client, sender.GetAddress())
	require.Equal(t, senderPublicBalance, parseMPTAmount(t, senderToken.MPTAmount))
	// The merge and the send each rewrote the spending balance, and nothing else did.
	require.Equal(t, uint32(2), senderToken.ConfidentialBalanceVersion)
	require.Equal(t, senderConfidentialBalance, decryptBalance(t, senderToken.ConfidentialBalanceSpending, senderKey.PrivKeyHex))
	require.Equal(t, uint64(0), decryptBalance(t, senderToken.ConfidentialBalanceInbox, senderKey.PrivKeyHex))

	receiverToken := getMPToken(t, client, receiver.GetAddress())
	require.Equal(t, lifecycleConvertBack, parseMPTAmount(t, receiverToken.MPTAmount))
	// The merge and the convert back each rewrote the spending balance.
	require.Equal(t, uint32(2), receiverToken.ConfidentialBalanceVersion)
	require.Equal(t, receiverConfidentialBalance, decryptBalance(t, receiverToken.ConfidentialBalanceSpending, receiverKey.PrivKeyHex))
	require.Equal(t, uint64(0), decryptBalance(t, receiverToken.ConfidentialBalanceInbox, receiverKey.PrivKeyHex))

	issuance := getIssuance(t, client, issuer.GetAddress())
	require.True(t, strings.EqualFold(config.issuerKey.PubKeyHex, issuance.IssuerEncryptionKey))
	require.True(t, strings.EqualFold(config.auditorKey.PubKeyHex, issuance.AuditorEncryptionKey))
	require.Equal(t, lifecycleFunding, parseMPTAmount(t, issuance.OutstandingAmount))
	require.Equal(t, senderConfidentialBalance+receiverConfidentialBalance, parseMPTAmount(t, issuance.ConfidentialOutstandingAmount))
	require.Equal(t, ledger.LsfMPTRequireAuth|ledger.LsfMPTCanTransfer|ledger.LsfMPTCanHoldConfidentialBalance, issuance.Flags)

	// Nothing was created or destroyed: what the issuer paid out is still split between
	// the two holders, across their public and confidential balances.
	require.Equal(
		t,
		lifecycleFunding,
		parseMPTAmount(t, senderToken.MPTAmount)+
			parseMPTAmount(t, receiverToken.MPTAmount)+
			parseMPTAmount(t, issuance.ConfidentialOutstandingAmount),
	)
}

func TestIntegrationConfidentialMPTLifecycle_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTLifecycle(t, client)
}

func TestIntegrationConfidentialMPTLifecycle_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTLifecycle(t, client)
}
