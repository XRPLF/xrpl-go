//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

const (
	auditorlessFunding uint64 = 10
	auditorlessSend    uint64 = 2
)

// testIntegrationConfidentialMPTWithoutAuditor covers an issuance that registers no
// auditor key. Every confidential transaction then omits its auditor ciphertext, and the
// ledger keeps the auditor mirror empty rather than encrypting to a key nobody holds. The
// scenario ends by replaying a spent proof, which is the one rejection that has to come
// from the network rather than from the builder.
func testIntegrationConfidentialMPTWithoutAuditor(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	sender := runner.GetWallet(1)
	receiver := runner.GetWallet(2)

	// This scenario also carries the two configurations no other one does: the post-creation
	// route to confidentiality, where one MPTokenIssuanceSet both enables it and registers
	// the issuer key, and a permissionless issuance, where holders need no authorization.
	config := issuanceConfig{issuerKey: generateKey(t), postCreationEnable: true, permissionless: true}
	senderKey := generateKey(t)
	receiverKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	// A permissionless issuance still needs each holder to create its own MPToken, it just
	// does not need the issuer to approve it afterwards.
	optInHolder(t, runner, sender, issuanceID)
	optInHolder(t, runner, receiver, issuanceID)
	fundHolder(t, runner, issuer, sender, issuanceID, auditorlessFunding)

	convert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       sender.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        auditorlessFunding,
		HolderPrivKey: senderKey.PrivKeyHex,
		HolderPubKey:  senderKey.PubKeyHex,
	})
	require.NoError(t, err)
	require.Nil(t, convert.AuditorEncryptedAmount)
	submitAndWait(t, runner, convert.Flatten(), sender)

	merge, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		Account:    sender.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, merge.Flatten(), sender)

	optIn, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        0,
		HolderPrivKey: receiverKey.PrivKeyHex,
		HolderPubKey:  receiverKey.PubKeyHex,
	})
	require.NoError(t, err)
	require.Nil(t, optIn.AuditorEncryptedAmount)
	submitAndWait(t, runner, optIn.Flatten(), receiver)

	send, err := builder.BuildSend(client, builder.BuildSendParams{
		Account:       sender.GetAddress().String(),
		Destination:   receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        auditorlessSend,
		SenderPrivKey: senderKey.PrivKeyHex,
		SenderPubKey:  senderKey.PubKeyHex,
		BalanceRange:  exactRange(auditorlessFunding),
	})
	require.NoError(t, err)
	require.Nil(t, send.AuditorEncryptedAmount)
	submitAndWait(t, runner, send.Flatten(), sender)

	assertMirrorBalances(t, client, sender.GetAddress(), config, auditorlessFunding-auditorlessSend)
	assertMirrorBalances(t, client, receiver.GetAddress(), config, auditorlessSend)
	issuance := getIssuance(t, client, issuer.GetAddress())
	require.Empty(t, issuance.AuditorEncryptionKey)
	require.NotEmpty(t, issuance.IssuerEncryptionKey)
	require.Equal(
		t,
		ledger.LsfMPTCanTransfer|ledger.LsfMPTCanHoldConfidentialBalance,
		issuance.Flags,
	)

	// Resubmitting the same proof under a new sequence breaks the binding between the
	// proof and the transaction that carries it, and only the network can catch that.
	stale := send.Flatten()
	delete(stale, "Sequence")
	submitExpectingResult(t, client, stale, sender, "tecBAD_PROOF")
}

func TestIntegrationConfidentialMPTWithoutAuditor_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTWithoutAuditor(t, client)
}

func TestIntegrationConfidentialMPTWithoutAuditor_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTWithoutAuditor(t, client)
}
