package payments

import (
	"testing"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationPaymentChannelFund(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	t.Run("pass - base payment channel fund", func(t *testing.T) {
		paymentChannelCreateTx := &transaction.PaymentChannelCreate{
			BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
			Amount:      types.XRPCurrencyAmount(100),
			Destination: receiver.GetAddress(),
			SettleDelay: 86400,
			PublicKey:   sender.PublicKey,
		}

		flatPaymentChannelCreateTx := paymentChannelCreateTx.Flatten()
		res, err := runner.TestTransaction(&flatPaymentChannelCreateTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)

		createSequence := integration.TxFieldUint32(t, res.Tx, "Sequence")
		channelID, err := xrplhash.PaymentChannel(
			sender.GetAddress().String(),
			receiver.GetAddress().String(),
			createSequence,
		)
		require.NoError(t, err)

		paymentChannelFundTx := &transaction.PaymentChannelFund{
			BaseTx:  transaction.BaseTx{Account: sender.GetAddress()},
			Channel: types.Hash256(channelID),
			Amount:  types.XRPCurrencyAmount(100),
		}
		flatPaymentChannelFundTx := paymentChannelFundTx.Flatten()
		_, err = runner.TestTransaction(&flatPaymentChannelFundTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)
	})
}

func TestIntegrationPaymentChannelFund_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationPaymentChannelFund(t, client)
}

func TestIntegrationPaymentChannelFund_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationPaymentChannelFund(t, client)
}
