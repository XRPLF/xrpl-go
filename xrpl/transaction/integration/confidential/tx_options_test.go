//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package confidential

import (
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
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
	optionsFunding        uint64 = 40
	optionsSend           uint64 = 12
	optionsDestinationTag uint32 = 42
	optionsDelegatedSend  uint64 = 5
)

// createTicket creates one Ticket and returns the sequence it reserved.
func createTicket(t *testing.T, runner *integration.Runner, client confidentialClient, owner *wallet.Wallet) uint32 {
	t.Helper()

	ticketCreate := transaction.TicketCreate{
		BaseTx:      transaction.BaseTx{Account: owner.GetAddress()},
		TicketCount: 1,
	}
	submitAndWait(t, runner, ticketCreate.Flatten(), owner)

	objects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: owner.GetAddress(),
		Type:    account.TicketObject,
	})
	require.NoError(t, err)
	require.Len(t, objects.AccountObjects, 1)

	ticketSequence := integration.TxFieldUint32(t, objects.AccountObjects[0], "TicketSequence")
	require.NotZero(t, ticketSequence)
	return ticketSequence
}

// authorizeDelegate grants one transaction type to the delegate, per XLS-75.
func authorizeDelegate(
	t *testing.T,
	runner *integration.Runner,
	delegator, delegate *wallet.Wallet,
	txType transaction.TxType,
) {
	t.Helper()

	delegateSet := transaction.DelegateSet{
		BaseTx:    transaction.BaseTx{Account: delegator.GetAddress()},
		Authorize: delegate.GetAddress(),
		Permissions: []types.Permission{
			{Permission: types.PermissionValue{PermissionValue: txType.String()}},
		},
	}
	submitAndWait(t, runner, delegateSet.Flatten(), delegator)
}

func xAddress(t *testing.T, classic types.Address, tag uint32, hasTag bool) string {
	t.Helper()

	encoded, err := addresscodec.ClassicAddressToXAddress(classic.String(), tag, hasTag, true)
	require.NoError(t, err)
	return encoded
}

// testIntegrationConfidentialMPTTransactionOptions covers what TxOptions adds to every
// confidential builder: a ticket in place of the account sequence, a delegate submitting
// on the holder's behalf, and either address form on the account and destination fields.
// Each of those changes what a proof commits to or which account the proof binds, so a
// unit test cannot tell whether the network agrees.
func testIntegrationConfidentialMPTTransactionOptions(t *testing.T, client confidentialClient) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	holder := runner.GetWallet(1)
	// The delegate submits for the holder and also receives the confidential send, so one
	// funded wallet covers both roles.
	delegate := runner.GetWallet(2)

	auditorKey := generateKey(t)
	config := issuanceConfig{issuerKey: generateKey(t), auditorKey: &auditorKey}
	holderKey := generateKey(t)
	delegateKey := generateKey(t)

	issuanceID := createIssuance(t, runner, client, issuer, config)
	authorizeHolder(t, runner, issuer, holder, issuanceID)
	authorizeHolder(t, runner, issuer, delegate, issuanceID)
	fundHolder(t, runner, issuer, holder, issuanceID, optionsFunding)

	// A ticket replaces the account sequence, and rippled hashes the sequence proxy, so
	// the proof has to commit to the ticket rather than to the account sequence.
	ticketSequence := createTicket(t, runner, client, holder)
	convert, err := builder.BuildConvert(client, builder.BuildConvertParams{
		TxOptions:     builder.TxOptions{TicketSequence: ticketSequence},
		Account:       holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        optionsFunding,
		HolderPrivKey: holderKey.PrivKeyHex,
		HolderPubKey:  holderKey.PubKeyHex,
	})
	require.NoError(t, err)
	require.Zero(t, convert.Sequence)
	require.Equal(t, ticketSequence, convert.TicketSequence)
	convertResponse := submitAndWait(t, runner, convert.Flatten(), holder)
	require.Zero(t, integration.TxFieldUint32(t, convertResponse.TxJSON, "Sequence"))
	require.Equal(t, ticketSequence, integration.TxFieldUint32(t, convertResponse.TxJSON, "TicketSequence"))
	assertMirrorBalances(t, client, holder.GetAddress(), config, optionsFunding)

	optIn, err := builder.BuildConvert(client, builder.BuildConvertParams{
		Account:       delegate.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        0,
		HolderPrivKey: delegateKey.PrivKeyHex,
		HolderPubKey:  delegateKey.PubKeyHex,
	})
	require.NoError(t, err)
	submitAndWait(t, runner, optIn.Flatten(), delegate)

	// A delegated transaction spends the holder's sequence but is signed and paid for by
	// the delegate, so the builder resolves the nonce from the holder either way.
	authorizeDelegate(t, runner, holder, delegate, transaction.ConfidentialMPTMergeInboxTx)
	merge, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		TxOptions:  builder.TxOptions{Delegate: delegate.GetAddress().String()},
		Account:    holder.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.NoError(t, err)
	require.Equal(t, delegate.GetAddress(), merge.Delegate)
	mergeResponse := submitAndWait(t, runner, merge.Flatten(), delegate)
	require.Equal(t, delegate.GetAddress().String(), mergeResponse.TxJSON["Delegate"])
	require.Equal(t, holder.GetAddress().String(), mergeResponse.TxJSON["Account"])

	// A tagged X-address destination carries its tag in the address, so the builder
	// leaves DestinationTag unset and autofill splits the address into both fields.
	senderXAddress := xAddress(t, holder.GetAddress(), 0, false)
	destinationXAddress := xAddress(t, delegate.GetAddress(), optionsDestinationTag, true)
	send, err := builder.BuildSend(client, builder.BuildSendParams{
		Account:       senderXAddress,
		Destination:   destinationXAddress,
		IssuanceID:    issuanceID,
		Amount:        optionsSend,
		SenderPrivKey: holderKey.PrivKeyHex,
		SenderPubKey:  holderKey.PubKeyHex,
		BalanceRange:  exactRange(optionsFunding),
	})
	require.NoError(t, err)
	require.Nil(t, send.DestinationTag)
	require.Equal(t, types.Address(destinationXAddress), send.Destination)
	sendResponse := submitAndWait(t, runner, send.Flatten(), holder)
	require.Equal(t, holder.GetAddress().String(), sendResponse.TxJSON["Account"])
	require.Equal(t, delegate.GetAddress().String(), sendResponse.TxJSON["Destination"])
	require.Equal(t, optionsDestinationTag, integration.TxFieldUint32(t, sendResponse.TxJSON, "DestinationTag"))

	assertMirrorBalances(t, client, holder.GetAddress(), config, optionsFunding-optionsSend)
	assertMirrorBalances(t, client, delegate.GetAddress(), config, optionsSend)

	// The merge above delegates a transaction that carries no proof. A send does carry one,
	// and its context has to bind the holder in Account rather than the delegate that signs
	// and pays. Only the network can tell those apart, and it reports the difference as
	// tecBAD_PROOF, so the assertion that matters is that this applies at all. Granting the
	// permission again replaces the merge grant, which is spent by now.
	authorizeDelegate(t, runner, holder, delegate, transaction.ConfidentialMPTSendTx)
	delegatedSend, err := builder.BuildSend(client, builder.BuildSendParams{
		TxOptions:     builder.TxOptions{Delegate: delegate.GetAddress().String()},
		Account:       holder.GetAddress().String(),
		Destination:   delegate.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        optionsDelegatedSend,
		SenderPrivKey: holderKey.PrivKeyHex,
		SenderPubKey:  holderKey.PubKeyHex,
		BalanceRange:  exactRange(optionsFunding - optionsSend),
	})
	require.NoError(t, err)
	require.Equal(t, delegate.GetAddress(), delegatedSend.Delegate)
	delegatedResponse := submitAndWait(t, runner, delegatedSend.Flatten(), delegate)
	require.Equal(t, delegate.GetAddress().String(), delegatedResponse.TxJSON["Delegate"])
	require.Equal(t, holder.GetAddress().String(), delegatedResponse.TxJSON["Account"])

	assertMirrorBalances(t, client, holder.GetAddress(), config, optionsFunding-optionsSend-optionsDelegatedSend)
	assertMirrorBalances(t, client, delegate.GetAddress(), config, optionsSend+optionsDelegatedSend)

	// The two nonces are mutually exclusive, and rippled marks ConfidentialMPTConvert
	// non-delegable. Both are rejected before any ledger query or proof work.
	_, err = builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
		TxOptions:  builder.TxOptions{Sequence: 1, TicketSequence: ticketSequence},
		Account:    holder.GetAddress().String(),
		IssuanceID: issuanceID,
	})
	require.ErrorIs(t, err, builder.ErrConflictingNonce)

	_, err = builder.BuildConvert(client, builder.BuildConvertParams{
		TxOptions:     builder.TxOptions{Delegate: delegate.GetAddress().String()},
		Account:       holder.GetAddress().String(),
		IssuanceID:    issuanceID,
		Amount:        1,
		HolderPrivKey: holderKey.PrivKeyHex,
		HolderPubKey:  holderKey.PubKeyHex,
	})
	require.ErrorIs(t, err, builder.ErrDelegateNotAllowed)
}

func TestIntegrationConfidentialMPTTransactionOptions_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationConfidentialMPTTransactionOptions(t, client)
}

func TestIntegrationConfidentialMPTTransactionOptions_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationConfidentialMPTTransactionOptions(t, client)
}
