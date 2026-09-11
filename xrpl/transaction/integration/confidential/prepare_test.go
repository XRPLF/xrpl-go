//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

const (
	prepareFunding       uint64 = 50
	prepareFirstConvert  uint64 = 30
	prepareRepeatConvert uint64 = 6
	prepareSend          uint64 = 12
	prepareConvertBack   uint64 = 5

	prepareClawbackFunding uint64 = 20
	prepareClawbackConvert uint64 = 14
)

// testIntegrationConfidentialMPTPrepareLifecycle runs the same lifecycle the Build*
// scenario runs, but assembled entirely from Prepare* helpers.
//
// A Prepare* helper reads no ledger state: the caller supplies the encryption keys, the
// spending ciphertext, the balance version, and the plaintext balance the proof commits
// to. Those are exactly the inputs a Build* helper fills in for you, so a Build*-only
// suite never proves the network accepts a proof built from caller-supplied inputs. This
// scenario reads them back off the ledger the way an application assembling its own
// transactions would, and submits the result.
func testIntegrationConfidentialMPTPrepareLifecycle(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	sender := runner.GetWallet(1)
	receiver := runner.GetWallet(2)

	auditorKey := generateKey(t)
	config := issuanceConfig{issuerKey: generateKey(t), auditorKey: &auditorKey}
	senderKey := generateKey(t)
	receiverKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, sender, issuanceID)
	authorizeHolder(t, runner, issuer, receiver, issuanceID)
	fundHolder(t, runner, issuer, sender, issuanceID, prepareFunding)

	// The issuance carries the keys every proof encrypts to, in the spelling rippled
	// stores them in rather than the one the issuer registered.
	issuance := getIssuance(t, client, issuer.GetAddress())

	// The first-time form is the only convert that carries a proof, so it is the only one
	// that needs the holder private key and a settled nonce.
	firstConvert, err := builder.PrepareConvert(builder.ConvertParams{
		BuildConvertParams: builder.BuildConvertParams{
			TxOptions:     builder.TxOptions{Sequence: accountSequence(t, client, sender.GetAddress())},
			Account:       sender.GetAddress().String(),
			IssuanceID:    issuanceID,
			Amount:        prepareFirstConvert,
			HolderPrivKey: senderKey.PrivKeyHex,
			HolderPubKey:  senderKey.PubKeyHex,
		},
		IssuerPubKey:  issuance.IssuerEncryptionKey,
		AuditorPubKey: issuance.AuditorEncryptionKey,
		FirstTime:     true,
	})
	require.NoError(t, err)
	require.NotNil(t, firstConvert.HolderEncryptionKey)
	require.NotNil(t, firstConvert.ZKProof)
	require.NotNil(t, firstConvert.AuditorEncryptedAmount)
	submitAndWait(t, runner, firstConvert.Flatten(), sender)
	assertMirrorBalances(t, client, sender.GetAddress(), config, prepareFirstConvert)

	// A repeat convert proves nothing, so no proof binds its nonce and autofill is free to
	// supply one. Submitting it with a zero sequence is what pins that.
	repeatConvert, err := builder.PrepareConvert(builder.ConvertParams{
		BuildConvertParams: builder.BuildConvertParams{
			Account:      sender.GetAddress().String(),
			IssuanceID:   issuanceID,
			Amount:       prepareRepeatConvert,
			HolderPubKey: senderKey.PubKeyHex,
		},
		IssuerPubKey:  issuance.IssuerEncryptionKey,
		AuditorPubKey: issuance.AuditorEncryptionKey,
	})
	require.NoError(t, err)
	require.Zero(t, repeatConvert.Sequence)
	require.Nil(t, repeatConvert.HolderEncryptionKey)
	require.Nil(t, repeatConvert.ZKProof)
	submitAndWait(t, runner, repeatConvert.Flatten(), sender)

	const senderBalanceBeforeSend = prepareFirstConvert + prepareRepeatConvert
	assertMirrorBalances(t, client, sender.GetAddress(), config, senderBalanceBeforeSend)

	// A merge carries no proof either, so it too can leave its nonce to autofill.
	mergeSender, err := builder.PrepareMergeInbox(builder.MergeInboxParams{
		BuildMergeInboxParams: builder.BuildMergeInboxParams{
			Account:    sender.GetAddress().String(),
			IssuanceID: issuanceID,
		},
	})
	require.NoError(t, err)
	require.Zero(t, mergeSender.Sequence)
	submitAndWait(t, runner, mergeSender.Flatten(), sender)

	optIn, err := builder.PrepareConvert(builder.ConvertParams{
		BuildConvertParams: builder.BuildConvertParams{
			TxOptions:     builder.TxOptions{Sequence: accountSequence(t, client, receiver.GetAddress())},
			Account:       receiver.GetAddress().String(),
			IssuanceID:    issuanceID,
			Amount:        0,
			HolderPrivKey: receiverKey.PrivKeyHex,
			HolderPubKey:  receiverKey.PubKeyHex,
		},
		IssuerPubKey:  issuance.IssuerEncryptionKey,
		AuditorPubKey: issuance.AuditorEncryptionKey,
		FirstTime:     true,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, optIn.Flatten(), receiver)

	// A send proof commits to the spending ciphertext, its version, and the plaintext
	// balance behind it, so all three have to come off the ledger together.
	senderToken := getMPToken(t, client, sender.GetAddress())
	receiverKeyOnLedger := getMPToken(t, client, receiver.GetAddress()).HolderEncryptionKey
	require.NotEmpty(t, receiverKeyOnLedger)

	send, err := builder.PrepareSend(builder.SendParams{
		BuildSendParams: builder.BuildSendParams{
			TxOptions:     builder.TxOptions{Sequence: accountSequence(t, client, sender.GetAddress())},
			Account:       sender.GetAddress().String(),
			Destination:   receiver.GetAddress().String(),
			IssuanceID:    issuanceID,
			Amount:        prepareSend,
			SenderPrivKey: senderKey.PrivKeyHex,
			SenderPubKey:  senderKey.PubKeyHex,
		},
		ReceiverPubKey:   receiverKeyOnLedger,
		IssuerPubKey:     issuance.IssuerEncryptionKey,
		AuditorPubKey:    issuance.AuditorEncryptionKey,
		BalanceVersion:   senderToken.ConfidentialBalanceVersion,
		CurrentBalance:   decryptBalance(t, senderToken.ConfidentialBalanceSpending, senderKey.PrivKeyHex),
		CurrentBalanceCt: senderToken.ConfidentialBalanceSpending,
	})
	require.NoError(t, err)
	require.NotNil(t, send.AuditorEncryptedAmount)
	submitAndWait(t, runner, send.Flatten(), sender)

	const senderConfidentialBalance = senderBalanceBeforeSend - prepareSend
	assertMirrorBalances(t, client, sender.GetAddress(), config, senderConfidentialBalance)
	assertMirrorBalances(t, client, receiver.GetAddress(), config, prepareSend)

	mergeReceiver, err := builder.PrepareMergeInbox(builder.MergeInboxParams{
		BuildMergeInboxParams: builder.BuildMergeInboxParams{
			Account:    receiver.GetAddress().String(),
			IssuanceID: issuanceID,
		},
	})
	require.NoError(t, err)
	submitAndWait(t, runner, mergeReceiver.Flatten(), receiver)

	receiverToken := getMPToken(t, client, receiver.GetAddress())
	convertBack, err := builder.PrepareConvertBack(builder.ConvertBackParams{
		BuildConvertBackParams: builder.BuildConvertBackParams{
			TxOptions:     builder.TxOptions{Sequence: accountSequence(t, client, receiver.GetAddress())},
			Account:       receiver.GetAddress().String(),
			IssuanceID:    issuanceID,
			Amount:        prepareConvertBack,
			HolderPrivKey: receiverKey.PrivKeyHex,
			HolderPubKey:  receiverKey.PubKeyHex,
		},
		IssuerPubKey:     issuance.IssuerEncryptionKey,
		AuditorPubKey:    issuance.AuditorEncryptionKey,
		BalanceVersion:   receiverToken.ConfidentialBalanceVersion,
		CurrentBalance:   decryptBalance(t, receiverToken.ConfidentialBalanceSpending, receiverKey.PrivKeyHex),
		CurrentBalanceCt: receiverToken.ConfidentialBalanceSpending,
	})
	require.NoError(t, err)
	require.NotNil(t, convertBack.AuditorEncryptedAmount)
	submitAndWait(t, runner, convertBack.Flatten(), receiver)

	const receiverConfidentialBalance = prepareSend - prepareConvertBack

	finalSender := getMPToken(t, client, sender.GetAddress())
	require.Equal(t, prepareFunding-senderBalanceBeforeSend, parseMPTAmount(t, finalSender.MPTAmount))
	require.Equal(t, senderConfidentialBalance, decryptBalance(t, finalSender.ConfidentialBalanceSpending, senderKey.PrivKeyHex))

	finalReceiver := getMPToken(t, client, receiver.GetAddress())
	require.Equal(t, prepareConvertBack, parseMPTAmount(t, finalReceiver.MPTAmount))
	require.Equal(t, receiverConfidentialBalance, decryptBalance(t, finalReceiver.ConfidentialBalanceSpending, receiverKey.PrivKeyHex))

	finalIssuance := getIssuance(t, client, issuer.GetAddress())
	require.Equal(t, prepareFunding, parseMPTAmount(t, finalIssuance.OutstandingAmount))
	require.Equal(
		t,
		prepareFunding,
		parseMPTAmount(t, finalSender.MPTAmount)+
			parseMPTAmount(t, finalReceiver.MPTAmount)+
			parseMPTAmount(t, finalIssuance.ConfidentialOutstandingAmount),
	)
}

// testIntegrationConfidentialMPTPrepareClawback covers the one Prepare* helper the
// lifecycle cannot reach, because a clawback needs an issuance created with
// LsfMPTCanClawback. The issuer supplies both the amount and the ciphertext the proof
// commits to, which is the whole of what BuildClawback would otherwise derive.
func testIntegrationConfidentialMPTPrepareClawback(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	holder := runner.GetWallet(1)

	auditorKey := generateKey(t)
	config := issuanceConfig{issuerKey: generateKey(t), auditorKey: &auditorKey, canClawback: true}
	holderKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, holder, issuanceID)
	fundHolder(t, runner, issuer, holder, issuanceID, prepareClawbackFunding)

	issuance := getIssuance(t, client, issuer.GetAddress())
	convert, err := builder.PrepareConvert(builder.ConvertParams{
		BuildConvertParams: builder.BuildConvertParams{
			TxOptions:     builder.TxOptions{Sequence: accountSequence(t, client, holder.GetAddress())},
			Account:       holder.GetAddress().String(),
			IssuanceID:    issuanceID,
			Amount:        prepareClawbackConvert,
			HolderPrivKey: holderKey.PrivKeyHex,
			HolderPubKey:  holderKey.PubKeyHex,
		},
		IssuerPubKey:  issuance.IssuerEncryptionKey,
		AuditorPubKey: issuance.AuditorEncryptionKey,
		FirstTime:     true,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, convert.Flatten(), holder)

	merge, err := builder.PrepareMergeInbox(builder.MergeInboxParams{
		BuildMergeInboxParams: builder.BuildMergeInboxParams{
			Account:    holder.GetAddress().String(),
			IssuanceID: issuanceID,
		},
	})
	require.NoError(t, err)
	submitAndWait(t, runner, merge.Flatten(), holder)

	// The clawback proof commits to the issuer mirror, so the amount and the ciphertext
	// have to be the pair the ledger currently holds.
	before := getMPToken(t, client, holder.GetAddress())
	clawback, err := builder.PrepareClawback(builder.ClawbackParams{
		BuildClawbackParams: builder.BuildClawbackParams{
			TxOptions:     builder.TxOptions{Sequence: accountSequence(t, client, issuer.GetAddress())},
			Account:       issuer.GetAddress().String(),
			Holder:        holder.GetAddress().String(),
			IssuanceID:    issuanceID,
			IssuerPrivKey: config.issuerKey.PrivKeyHex,
		},
		Amount:           decryptBalance(t, before.IssuerEncryptedBalance, config.issuerKey.PrivKeyHex),
		IssuerPubKey:     issuance.IssuerEncryptionKey,
		IssuerCiphertext: before.IssuerEncryptedBalance,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, clawback.Flatten(), issuer)

	after := getMPToken(t, client, holder.GetAddress())
	require.Equal(t, before.ConfidentialBalanceVersion+1, after.ConfidentialBalanceVersion)
	require.Equal(t, prepareClawbackFunding-prepareClawbackConvert, parseMPTAmount(t, after.MPTAmount))
	require.Equal(t, uint64(0), decryptBalance(t, after.ConfidentialBalanceSpending, holderKey.PrivKeyHex))
	assertMirrorBalances(t, client, holder.GetAddress(), config, 0)
	require.Empty(t, getIssuance(t, client, issuer.GetAddress()).ConfidentialOutstandingAmount)
}

func TestIntegrationConfidentialMPTPrepareLifecycle_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTPrepareLifecycle(t, client)
}

func TestIntegrationConfidentialMPTPrepareLifecycle_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTPrepareLifecycle(t, client)
}

func TestIntegrationConfidentialMPTPrepareClawback_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTPrepareClawback(t, client)
}

func TestIntegrationConfidentialMPTPrepareClawback_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTPrepareClawback(t, client)
}
