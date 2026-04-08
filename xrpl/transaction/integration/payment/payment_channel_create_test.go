package payment

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

func testIntegrationPaymentChannelCreate(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	t.Run("pass - payment channel create base", func(t *testing.T) {
		paymentChannelCreateTx := &transaction.PaymentChannelCreate{
			BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
			Amount:      types.XRPCurrencyAmount(100),
			Destination: receiver.GetAddress(),
			SettleDelay: 86400,
			PublicKey:   sender.PublicKey,
		}

		flatPaymentChannelCreateTx := paymentChannelCreateTx.Flatten()
		_, err = runner.TestTransaction(&flatPaymentChannelCreateTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)

		objects, err := client.GetAccountObjects(&account.ObjectsRequest{
			Account: sender.GetAddress(),
			Type:    account.PaymentChannelObject,
		})
		require.NoError(t, err)
		require.Len(t, objects.AccountObjects, 1)
	})
}

func TestIntegrationPaymentChannelCreate_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationPaymentChannelCreate(t, client)
}

func TestIntegrationPaymentChannelCreate_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationPaymentChannelCreate(t, client)
}
