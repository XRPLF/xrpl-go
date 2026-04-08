package escrow

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationEscrowFinish(t *testing.T, client integration.Client) {
	t.Run("pass - base escrow finish", func(t *testing.T) {
		runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
		err := runner.Setup()
		require.NoError(t, err)
		defer runner.Teardown()

		sender := runner.GetWallet(0)
		receiver := runner.GetWallet(1)

		closeTime := getLedgerCloseTime(t, client)
		const amount = types.XRPCurrencyAmount(10000)

		escrowCreateTx := &transaction.EscrowCreate{
			BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
			Amount:      amount,
			Destination: receiver.GetAddress(),
			FinishAfter: uint32(closeTime + 2),
		}
		flatEscrowCreateTx := escrowCreateTx.Flatten()
		res, err := runner.TestTransaction(&flatEscrowCreateTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)

		objects, err := client.GetAccountObjects(&account.ObjectsRequest{
			Account: sender.GetAddress(),
			Type:    account.EscrowObject,
		})
		require.NoError(t, err)
		require.Len(t, objects.AccountObjects, 1)

		offerSequence := integration.TxFieldUint32(t, res.Tx, "Sequence")
		escrowFinishTx := &transaction.EscrowFinish{
			BaseTx:        transaction.BaseTx{Account: sender.GetAddress()},
			Owner:         sender.GetAddress(),
			OfferSequence: offerSequence,
		}
		waitForLedgerTime(t, client, closeTime+2)
		flatEscrowFinishTx := escrowFinishTx.Flatten()
		_, err = runner.TestTransaction(&flatEscrowFinishTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)

		objects, err = client.GetAccountObjects(&account.ObjectsRequest{
			Account: sender.GetAddress(),
			Type:    account.EscrowObject,
		})
		require.NoError(t, err)
		require.Empty(t, objects.AccountObjects)
	})
}

func TestIntegrationEscrowFinish_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationEscrowFinish(t, client)
}

func TestIntegrationEscrowFinish_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationEscrowFinish(t, client)
}
