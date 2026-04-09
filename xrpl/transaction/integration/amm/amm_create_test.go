package amm

import (
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationAMMCreate(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	pool := createAMMPool(t, runner, client, false)

	t.Run("pass - base", func(t *testing.T) {
		result := getAMMInfo(t, client, pool)

		require.True(t, addresscodec.IsValidAddress(result.Account.String()), "AMM account should be a valid classic address")

		require.Equal(t, types.XRPCurrencyAmount(250), result.Amount)
		require.Equal(t, types.IssuedCurrencyAmount{
			Currency: pool.asset2.Currency,
			Issuer:   pool.asset2.Issuer,
			Value:    "250",
		}, result.Amount2)

		require.Equal(t, uint16(12), result.TradingFee)
	})
}

func TestIntegrationAMMCreate_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationAMMCreate(t, client)
}

func TestIntegrationAMMCreate_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationAMMCreate(t, client)
}
