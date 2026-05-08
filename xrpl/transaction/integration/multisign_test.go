package integration

import (
	"maps"
	"strconv"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
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

	signerListSetTx := &transaction.SignerListSet{
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

	flatSignerListSetTx := signerListSetTx.Flatten()
	err = client.Autofill(&flatSignerListSetTx)
	require.NoError(t, err)

	signerListSetBlob, _, err := master.Sign(flatSignerListSetTx)
	require.NoError(t, err)

	_, err = client.SubmitTxBlobAndWait(signerListSetBlob, true)
	require.NoError(t, err)

	destinationBalanceBefore, err := xrpBalanceDrops(client, destination.GetAddress())
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
	err = client.AutofillMultisigned(&flatPaymentTx, 2)
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

	expectedBalance := destinationBalanceBefore + amount.Uint64()
	require.Eventually(t, func() bool {
		destinationBalanceAfter, err := xrpBalanceDrops(client, destination.GetAddress())
		return err == nil && destinationBalanceAfter == expectedBalance
	}, 30*time.Second, time.Second)
}

func multisignTxBlob(signer *wallet.Wallet, tx transaction.FlatTransaction) (string, error) {
	blob, _, err := signer.Multisign(maps.Clone(tx))
	return blob, err
}

func xrpBalanceDrops(client integration.Client, address types.Address) (uint64, error) {
	balance, err := client.GetXrpBalanceValidated(address)
	if err != nil {
		return 0, err
	}

	drops, err := currency.XrpToDrops(balance)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(drops, 10, 64)
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
