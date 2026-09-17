//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"encoding/hex"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/builder"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

const (
	credentialsFunding uint64 = 25
	credentialsSend    uint64 = 8
)

// credentialType is the XLS-70 credential type the receiver preauthorizes, hex-encoded as
// the protocol requires.
var credentialType = types.CredentialType(hex.EncodeToString([]byte("ConfidentialSender")))

// issueAcceptedCredential has the issuer issue a credential to the subject and the subject
// accept it, then returns the Credential ledger index. That index is what a sender passes
// in CredentialIDs, so it has to come off the ledger rather than be derived here.
func issueAcceptedCredential(
	t *testing.T,
	runner *integration.Runner,
	client confidentialClient,
	issuer, subject *wallet.Wallet,
) string {
	t.Helper()

	create := transaction.CredentialCreate{
		BaseTx:         transaction.BaseTx{Account: issuer.GetAddress()},
		Subject:        subject.GetAddress(),
		CredentialType: credentialType,
	}
	submitAndWait(t, runner, create.Flatten(), issuer)

	accept := transaction.CredentialAccept{
		BaseTx:         transaction.BaseTx{Account: subject.GetAddress()},
		Issuer:         issuer.GetAddress(),
		CredentialType: credentialType,
	}
	submitAndWait(t, runner, accept.Flatten(), subject)

	objects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: subject.GetAddress(),
		Type:    account.CredentialObject,
	})
	require.NoError(t, err)
	require.Len(t, objects.AccountObjects, 1)

	index, ok := objects.AccountObjects[0]["index"].(string)
	require.True(t, ok)
	return index
}

// testIntegrationConfidentialMPTCredentials covers the Deposit Authorization path on
// ConfidentialMPTSend. XLS-96 §8.2 puts CredentialIDs on the transaction and §8.3.2.1 makes
// an unauthorized send fail, and rippled checks both in ConfidentialMPTSend::preclaim ahead
// of proof verification. CredentialIDs is the only optional field on the confidential
// transaction types whose encoding no other scenario submits, and the builder documents
// credential authorization as something it cannot pre-check, so only the network can
// confirm it.
func testIntegrationConfidentialMPTCredentials(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	sender := runner.GetWallet(1)
	receiver := runner.GetWallet(2)

	config := issuanceConfig{issuerKey: generateKey(t)}
	senderKey := generateKey(t)
	receiverKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, sender, issuanceID)
	authorizeHolder(t, runner, issuer, receiver, issuanceID)
	fundHolder(t, runner, issuer, sender, issuanceID, credentialsFunding)

	convert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       sender.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        credentialsFunding,
		HolderPrivKey: senderKey.PrivKeyHex,
		HolderPubKey:  senderKey.PubKeyHex,
	})
	require.NoError(t, err)
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
	submitAndWait(t, runner, optIn.Flatten(), receiver)

	// From here the receiver accepts deposits only from preauthorized senders.
	depositAuth := transaction.AccountSet{BaseTx: transaction.BaseTx{Account: receiver.GetAddress()}}
	depositAuth.SetAsfDepositAuth()
	submitAndWait(t, runner, depositAuth.Flatten(), receiver)

	// An unauthorized send is rejected by the network, not by the builder: the builder has
	// no way to know the receiver's preauth list. rippled checks this before verifying the
	// proof, so the rejection costs a fee but no proof work.
	unauthorized, err := builder.BuildSend(client, builder.BuildSendParams{
		Account:       sender.GetAddress().String(),
		Destination:   receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        credentialsSend,
		SenderPrivKey: senderKey.PrivKeyHex,
		SenderPubKey:  senderKey.PubKeyHex,
		BalanceRange:  exactRange(credentialsFunding),
	})
	require.NoError(t, err)
	require.Empty(t, unauthorized.CredentialIDs)
	submitExpectingResult(t, client, unauthorized.Flatten(), sender, "tecNO_PERMISSION")

	// The rejection claimed a fee and a sequence but changed no balance, so the sender still
	// holds everything it converted.
	assertMirrorBalances(t, client, sender.GetAddress(), config, credentialsFunding)

	credentialID := issueAcceptedCredential(t, runner, client, issuer, sender)

	preauth := transaction.DepositPreauth{
		BaseTx: transaction.BaseTx{Account: receiver.GetAddress()},
		AuthorizeCredentials: []types.AuthorizeCredentialsWrapper{
			{Credential: types.AuthorizeCredentials{
				Issuer:         issuer.GetAddress(),
				CredentialType: credentialType,
			}},
		},
	}
	submitAndWait(t, runner, preauth.Flatten(), receiver)

	// The same send now carries the credential that satisfies the receiver's preauth.
	authorized, err := builder.BuildSend(client, builder.BuildSendParams{
		Account:       sender.GetAddress().String(),
		Destination:   receiver.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        credentialsSend,
		SenderPrivKey: senderKey.PrivKeyHex,
		SenderPubKey:  senderKey.PubKeyHex,
		BalanceRange:  exactRange(credentialsFunding),
		CredentialIDs: types.CredentialIDs{credentialID},
	})
	require.NoError(t, err)
	require.Equal(t, types.CredentialIDs{credentialID}, authorized.CredentialIDs)

	sendResponse := submitAndWait(t, runner, authorized.Flatten(), sender)
	require.Equal(t, []any{credentialID}, sendResponse.TxJSON["CredentialIDs"])

	assertMirrorBalances(t, client, sender.GetAddress(), config, credentialsFunding-credentialsSend)
	assertMirrorBalances(t, client, receiver.GetAddress(), config, credentialsSend)
}

func TestIntegrationConfidentialMPTCredentials_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTCredentials(t, client)
}

func TestIntegrationConfidentialMPTCredentials_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTCredentials(t, client)
}
