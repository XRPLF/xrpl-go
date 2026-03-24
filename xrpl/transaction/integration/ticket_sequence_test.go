package integration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	requests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	rpctypes "github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	testintegration "github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	wstypes "github.com/Peersyst/xrpl-go/xrpl/websocket/types"
	"github.com/stretchr/testify/require"
)

const ticketCredentialType = types.CredentialType("7469636b65745f63726564656e7469616c")

type accountObjectsClient interface {
	GetAccountObjects(*account.ObjectsRequest) (*account.ObjectsResponse, error)
}

type submitAndWaitFunc func(transaction.FlatTransaction) (*requests.TxResponse, error)

func uint32FromValue(v any) (uint32, error) {
	switch value := v.(type) {
	case json.Number:
		n, err := value.Int64()
		if err != nil {
			return 0, err
		}
		return uint32(n), nil
	case float64:
		return uint32(value), nil
	case uint32:
		return value, nil
	case int:
		return uint32(value), nil
	default:
		return 0, fmt.Errorf("unexpected numeric type: %T", v)
	}
}

func firstTicketSequence(t *testing.T, client accountObjectsClient, walletAddress types.Address) uint32 {
	t.Helper()

	objects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: walletAddress,
		Type:    account.TicketObject,
	})
	require.NoError(t, err)
	require.NotEmpty(t, objects.AccountObjects)

	for _, object := range objects.AccountObjects {
		if object["TicketSequence"] == nil {
			continue
		}

		sequence, err := uint32FromValue(object["TicketSequence"])
		require.NoError(t, err)
		return sequence
	}

	t.Fatal("ticket sequence not found in account objects response")
	return 0
}

// submitTicketCreate creates a ticket and waits for validation.
// It uses the Runner for autofill+sign (which retries on tefPAST_SEQ), then
// SubmitTxBlobAndWait to ensure the ticket is validated before querying it.
func submitTicketCreate(t *testing.T, runner *testintegration.Runner, client accountObjectsClient, sender *wallet.Wallet) uint32 {
	t.Helper()

	ticketCreate := (&transaction.TicketCreate{
		BaseTx: transaction.BaseTx{
			Account: sender.GetAddress(),
		},
		TicketCount: 1,
	}).Flatten()

	err := runner.GetClient().Autofill(&ticketCreate)
	require.NoError(t, err)

	blob, _, err := sender.Sign(ticketCreate)
	require.NoError(t, err)

	resp, err := runner.GetClient().SubmitTxBlobAndWait(blob, false)
	require.NoError(t, err)
	require.True(t, resp.Validated)

	return firstTicketSequence(t, client, sender.GetAddress())
}

func assertTicketedTxPreserved(t *testing.T, resp *requests.TxResponse, ticketSequence uint32) {
	t.Helper()

	require.True(t, resp.Validated)
	require.Equal(t, transaction.TesSUCCESS.String(), resp.Meta.TransactionResult)

	sequence, err := uint32FromValue(resp.TxJSON["Sequence"])
	require.NoError(t, err)
	require.Zero(t, sequence)

	validatedTicketSequence, err := uint32FromValue(resp.TxJSON["TicketSequence"])
	require.NoError(t, err)
	require.Equal(t, ticketSequence, validatedTicketSequence)
}

func runTicketedPaymentScenario(t *testing.T, submit submitAndWaitFunc, runner *testintegration.Runner, client accountObjectsClient, sender, receiver *wallet.Wallet) {
	t.Helper()

	ticketSequence := submitTicketCreate(t, runner, client, sender)

	payment := (&transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account:        sender.GetAddress(),
			Sequence:       0,
			TicketSequence: ticketSequence,
		},
		Destination: receiver.GetAddress(),
		Amount:      types.XRPCurrencyAmount(1),
	}).Flatten()

	resp, err := submit(payment)
	require.NoError(t, err)
	assertTicketedTxPreserved(t, resp, ticketSequence)
}

func runTicketedCredentialCreateScenario(t *testing.T, submit submitAndWaitFunc, runner *testintegration.Runner, client accountObjectsClient, sender, subject *wallet.Wallet) {
	t.Helper()

	ticketSequence := submitTicketCreate(t, runner, client, sender)

	credentialCreate := (&transaction.CredentialCreate{
		BaseTx: transaction.BaseTx{
			Account:        sender.GetAddress(),
			Sequence:       0,
			TicketSequence: ticketSequence,
		},
		CredentialType: ticketCredentialType,
		Subject:        subject.GetAddress(),
	}).Flatten()

	resp, err := submit(credentialCreate)
	require.NoError(t, err)
	assertTicketedTxPreserved(t, resp, ticketSequence)

	credentialType, ok := resp.TxJSON["CredentialType"].(string)
	require.True(t, ok)
	require.True(t, strings.EqualFold(string(ticketCredentialType), credentialType))

	txSubject, ok := resp.TxJSON["Subject"].(string)
	require.True(t, ok)
	require.Equal(t, subject.GetAddress().String(), txSubject)
}

func TestIntegrationTicketSequenceAutofillPreservesZeroSequence_Websocket(t *testing.T) {
	env := testintegration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := testintegration.NewRunner(t, client, &testintegration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	submit := func(tx transaction.FlatTransaction) (*requests.TxResponse, error) {
		return client.SubmitTxAndWait(tx, &wstypes.SubmitOptions{
			Autofill: true,
			Wallet:   sender,
		})
	}

	runTicketedPaymentScenario(t, submit, runner, client, sender, receiver)
}

func TestIntegrationTicketSequenceAutofillPreservesZeroSequence_RPCClient(t *testing.T) {
	env := testintegration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)

	runner := testintegration.NewRunner(t, client, testintegration.NewRunnerConfig(
		testintegration.WithWallets(2),
	))

	err = runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	submit := func(tx transaction.FlatTransaction) (*requests.TxResponse, error) {
		return client.SubmitTxAndWait(tx, &rpctypes.SubmitOptions{
			Autofill: true,
			Wallet:   sender,
		})
	}

	runTicketedPaymentScenario(t, submit, runner, client, sender, receiver)
}

func TestIntegrationTicketSequenceAutofillPreservesZeroSequence_CredentialCreate_Websocket(t *testing.T) {
	env := testintegration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	runner := testintegration.NewRunner(t, client, &testintegration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	subject := runner.GetWallet(1)

	submit := func(tx transaction.FlatTransaction) (*requests.TxResponse, error) {
		return client.SubmitTxAndWait(tx, &wstypes.SubmitOptions{
			Autofill: true,
			Wallet:   issuer,
		})
	}

	runTicketedCredentialCreateScenario(t, submit, runner, client, issuer, subject)
}

func TestIntegrationTicketSequenceAutofillPreservesZeroSequence_CredentialCreate_RPCClient(t *testing.T) {
	env := testintegration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)

	runner := testintegration.NewRunner(t, client, testintegration.NewRunnerConfig(
		testintegration.WithWallets(2),
	))

	err = runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	subject := runner.GetWallet(1)

	submit := func(tx transaction.FlatTransaction) (*requests.TxResponse, error) {
		return client.SubmitTxAndWait(tx, &rpctypes.SubmitOptions{
			Autofill: true,
			Wallet:   issuer,
		})
	}

	runTicketedCredentialCreateScenario(t, submit, runner, client, issuer, subject)
}
