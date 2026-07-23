package amm

import (
	"testing"

	ammqueries "github.com/Peersyst/xrpl-go/xrpl/queries/amm"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationAMMDelete(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	pool := createAMMPool(t, runner, client, false)

	t.Run("fail - AMMDelete on non-empty pool", func(t *testing.T) {
		ammDeleteTx := &transaction.AMMDelete{
			BaseTx: transaction.BaseTx{
				Account: pool.lpWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
		}
		flatAMMDeleteTx := ammDeleteTx.Flatten()
		_, err := runner.TestTransaction(&flatAMMDeleteTx, pool.lpWallet, "tecAMM_NOT_EMPTY", nil)
		require.NoError(t, err)

		// AMM should still exist after the failed delete
		ammInfo := getAMMInfo(t, client, pool)
		require.NotEmpty(t, ammInfo.Account)
	})

	t.Run("pass - withdraw all LP tokens deletes the AMM", func(t *testing.T) {
		lpToken := getLPToken(t, client, pool.lpWallet.GetAddress())

		// Withdraw all LP tokens. For pools with few trust lines this triggers
		// automatic deletion of the AMM.
		withdrawTx := &transaction.AMMWithdraw{
			BaseTx: transaction.BaseTx{
				Account: pool.lpWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			LPTokenIn: types.IssuedCurrencyAmount{
				Currency: lpToken.Currency,
				Issuer:   lpToken.Account,
				Value:    lpToken.Balance,
			},
		}
		withdrawTx.SetLPTokentFlag()
		flatWithdrawTx := withdrawTx.Flatten()
		_, err := runner.TestTransaction(&flatWithdrawTx, pool.lpWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		// Verify the AMM has been deleted: GetAMMInfo must return an error
		_, err = client.GetAMMInfo(&ammqueries.InfoRequest{
			Asset:  pool.asset,
			Asset2: pool.asset2,
		})
		require.Error(t, err, "expected error querying a deleted AMM")
	})
}

func TestIntegrationAMMDelete_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationAMMDelete(t, client)
}

func TestIntegrationAMMDelete_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationAMMDelete(t, client)
}
