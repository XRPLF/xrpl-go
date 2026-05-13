package integration

import (
	"maps"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testIntegrationMultisignPayment(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 4})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	master := runner.GetWallet(0)
	destination := runner.GetWallet(1)
	signer1 := runner.GetWallet(2)
	signer2 := runner.GetWallet(3)

	signerListSetTx := installSignerList(t, client, master, signer1, signer2)

	destinationBalanceBefore, err := client.GetXrpDropsBalanceValidated(destination.GetAddress())
	require.NoError(t, err)

	amount := types.XRPCurrencyAmount(1000)
	paymentTx := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: master.GetAddress(),
		},
		Amount:      amount,
		Destination: destination.GetAddress(),
	}
	flatPaymentTx := paymentTx.Flatten()
	err = client.AutofillMultisigned(&flatPaymentTx, uint64(len(signerListSetTx.SignerEntries)))
	require.NoError(t, err)

	signer1Blob, err := multisignTxBlob(signer1, flatPaymentTx)
	require.NoError(t, err)

	signer2Blob, err := multisignTxBlob(signer2, flatPaymentTx)
	require.NoError(t, err)

	blob, err := xrpl.Multisign(signer1Blob, signer2Blob)
	require.NoError(t, err)

	res, err := client.SubmitMultisigned(blob, true)
	require.NoError(t, err)
	require.Equal(t, transaction.TesSUCCESS.String(), res.EngineResult)

	expectedBalance := destinationBalanceBefore + amount
	var lastBalance types.XRPCurrencyAmount
	var lastErr error
	ok := assert.Eventually(t, func() bool {
		lastBalance, lastErr = client.GetXrpDropsBalanceValidated(destination.GetAddress())
		return lastErr == nil && lastBalance >= expectedBalance
	}, 30*time.Second, time.Second)
	if !ok {
		require.FailNowf(t, "destination balance never reached expected value",
			"expected >= %d, last balance=%d, last err=%v", expectedBalance, lastBalance, lastErr)
	}
}

func testIntegrationMultisignPaymentSubQuorum(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 4})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	master := runner.GetWallet(0)
	destination := runner.GetWallet(1)
	signer1 := runner.GetWallet(2)
	signer2 := runner.GetWallet(3)

	installSignerList(t, client, master, signer1, signer2)

	paymentTx := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: master.GetAddress(),
		},
		Amount:      types.XRPCurrencyAmount(1000),
		Destination: destination.GetAddress(),
	}
	flatPaymentTx := paymentTx.Flatten()
	err = client.AutofillMultisigned(&flatPaymentTx, 1)
	require.NoError(t, err)

	signer1Blob, err := multisignTxBlob(signer1, flatPaymentTx)
	require.NoError(t, err)

	blob, err := xrpl.Multisign(signer1Blob)
	require.NoError(t, err)

	res, err := client.SubmitMultisigned(blob, true)
	require.NoError(t, err)
	require.Equal(t, transaction.TefBAD_QUORUM.String(), res.EngineResult)
}

func testIntegrationMultisignPaymentNonListedSigner(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 5})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	master := runner.GetWallet(0)
	destination := runner.GetWallet(1)
	signer1 := runner.GetWallet(2)
	signer2 := runner.GetWallet(3)
	outsider := runner.GetWallet(4)

	installSignerList(t, client, master, signer1, signer2)

	paymentTx := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: master.GetAddress(),
		},
		Amount:      types.XRPCurrencyAmount(1000),
		Destination: destination.GetAddress(),
	}
	flatPaymentTx := paymentTx.Flatten()
	err = client.AutofillMultisigned(&flatPaymentTx, 2)
	require.NoError(t, err)

	signer1Blob, err := multisignTxBlob(signer1, flatPaymentTx)
	require.NoError(t, err)

	outsiderBlob, err := multisignTxBlob(outsider, flatPaymentTx)
	require.NoError(t, err)

	blob, err := xrpl.Multisign(signer1Blob, outsiderBlob)
	require.NoError(t, err)

	res, err := client.SubmitMultisigned(blob, true)
	require.NoError(t, err)
	require.Equal(t, transaction.TefBAD_SIGNATURE.String(), res.EngineResult)
}

func installSignerList(t *testing.T, client integration.Client, master, signer1, signer2 *wallet.Wallet) *transaction.SignerListSet {
	t.Helper()

	tx := &transaction.SignerListSet{
		BaseTx: transaction.BaseTx{
			Account: master.GetAddress(),
		},
		SignerQuorum: uint32(2),
		SignerEntries: []ledger.SignerEntryWrapper{
			{
				SignerEntry: ledger.SignerEntry{
					Account:      signer1.GetAddress(),
					SignerWeight: 1,
				},
			},
			{
				SignerEntry: ledger.SignerEntry{
					Account:      signer2.GetAddress(),
					SignerWeight: 1,
				},
			},
		},
	}

	flat := tx.Flatten()
	require.NoError(t, client.Autofill(&flat))

	blob, _, err := master.Sign(flat)
	require.NoError(t, err)

	_, err = client.SubmitTxBlobAndWait(blob, true)
	require.NoError(t, err)

	return tx
}

func multisignTxBlob(signer *wallet.Wallet, tx transaction.FlatTransaction) (string, error) {
	blob, _, err := signer.Multisign(maps.Clone(tx))
	return blob, err
}

func TestIntegrationMultisignPayment_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationMultisignPayment(t, client)
}

func TestIntegrationMultisignPayment_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationMultisignPayment(t, client)
}

func TestIntegrationMultisignPaymentSubQuorum_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationMultisignPaymentSubQuorum(t, client)
}

func TestIntegrationMultisignPaymentSubQuorum_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationMultisignPaymentSubQuorum(t, client)
}

func TestIntegrationMultisignPaymentNonListedSigner_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationMultisignPaymentNonListedSigner(t, client)
}

func TestIntegrationMultisignPaymentNonListedSigner_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationMultisignPaymentNonListedSigner(t, client)
}
