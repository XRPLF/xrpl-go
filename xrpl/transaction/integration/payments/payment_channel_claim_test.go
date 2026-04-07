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

func testIntegrationPaymentChannelClaim(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	t.Run("pass - base payment channel claim", func(t *testing.T) {
		paymentChannelCreateTx := &transaction.PaymentChannelCreate{
			BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
			Amount:      types.XRPCurrencyAmount(100),
			Destination: receiver.GetAddress(),
			SettleDelay: 86400,
			PublicKey:   sender.PublicKey,
		}
		flatPaymentChannelCreateTX := paymentChannelCreateTx.Flatten()
		res, err := runner.TestTransaction(&flatPaymentChannelCreateTX, sender, "tesSUCCESS", nil)
		require.NoError(t, err)

		createSequence := integration.TxFieldUint32(t, res.Tx, "Sequence")
		channelID, err := xrplhash.PaymentChannel(
			sender.GetAddress().String(),
			receiver.GetAddress().String(),
			createSequence,
		)
		require.NoError(t, err)

		paymentChannelClaimTx := &transaction.PaymentChannelClaim{
			BaseTx:  transaction.BaseTx{Account: sender.GetAddress()},
			Channel: types.Hash256(channelID),
			Amount:  types.XRPCurrencyAmount(100),
		}
		flatPaymentChannelClaimTx := paymentChannelClaimTx.Flatten()
		_, err = runner.TestTransaction(&flatPaymentChannelClaimTx, sender, "tesSUCCESS", nil)
		require.NoError(t, err)
	})
}

func TestIntegrationPaymentChannelClaim_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationPaymentChannelClaim(t, client)
}

func TestIntegrationPaymentChannelClaim_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationPaymentChannelClaim(t, client)
}
