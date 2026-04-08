package amm

import (
	"strconv"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationAMMWithdraw(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 3,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	pool := setupAMMPool(t, runner, client)

	t.Run("pass - withdraw with Amount", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDropsFromAny(t, preAmm.Amount)
		preLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)

		withdrawTx := &transaction.AMMWithdraw{
			BaseTx: transaction.BaseTx{
				Account: pool.testWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			Amount: types.XRPCurrencyAmount(500),
		}
		withdrawTx.SetSingleAssetFlag()
		flatWithdrawTx := withdrawTx.Flatten()
		_, err = runner.TestTransaction(&flatWithdrawTx, pool.testWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)

		postAmountDrops := xrpDropsFromAny(t, postAmm.Amount)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops-500, postAmountDrops)
		require.Equal(t, preAmm.Amount2, postAmm.Amount2)
		require.InDelta(t, preLPTokenValue-126, postLPTokenValue, 2.0)
	})

	t.Run("pass - withdraw with Amount and Amount2", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDropsFromAny(t, preAmm.Amount)
		preAmount2Value := icaValueFromAny(t, preAmm.Amount2)
		preLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)

		withdrawTx := &transaction.AMMWithdraw{
			BaseTx: transaction.BaseTx{
				Account: pool.testWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			Amount: types.XRPCurrencyAmount(50),
			Amount2: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   pool.issuerWallet.GetAddress(),
				Value:    "50",
			},
		}
		withdrawTx.SetTwoAssetFlag()
		flatWithdrawTx := withdrawTx.Flatten()
		_, err = runner.TestTransaction(&flatWithdrawTx, pool.testWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)

		postAmountDrops := xrpDropsFromAny(t, postAmm.Amount)
		postAmount2Value := icaValueFromAny(t, postAmm.Amount2)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops-50, postAmountDrops)
		require.InDelta(t, preAmount2Value-17, postAmount2Value, 2.0)
		require.InDelta(t, preLPTokenValue-28, postLPTokenValue, 2.0)
	})

	t.Run("pass - withdraw with Amount and LPTokenIn", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDropsFromAny(t, preAmm.Amount)
		preLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)

		lpToken := getLPToken(t, client, pool.testWallet.GetAddress())

		withdrawTx := &transaction.AMMWithdraw{
			BaseTx: transaction.BaseTx{
				Account: pool.testWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			Amount: types.XRPCurrencyAmount(5),
			LPTokenIn: types.IssuedCurrencyAmount{
				Currency: lpToken.Currency,
				Issuer:   lpToken.Account,
				Value:    "5",
			},
		}
		withdrawTx.SetOneAssetLPTokenFlag()
		flatWithdrawTx := withdrawTx.Flatten()
		_, err = runner.TestTransaction(&flatWithdrawTx, pool.testWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)

		postAmountDrops := xrpDropsFromAny(t, postAmm.Amount)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops-17, postAmountDrops)
		require.Equal(t, preAmm.Amount2, postAmm.Amount2)
		require.InDelta(t, preLPTokenValue-5, postLPTokenValue, 1e-6)
	})

	t.Run("pass - withdraw with LPTokenIn", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDropsFromAny(t, preAmm.Amount)
		preAmount2Value := icaValueFromAny(t, preAmm.Amount2)
		preLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)

		lpToken := getLPToken(t, client, pool.testWallet.GetAddress())

		withdrawTx := &transaction.AMMWithdraw{
			BaseTx: transaction.BaseTx{
				Account: pool.testWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			LPTokenIn: types.IssuedCurrencyAmount{
				Currency: lpToken.Currency,
				Issuer:   lpToken.Account,
				Value:    "5",
			},
		}
		withdrawTx.SetLPTokentFlag()
		flatWithdrawTx := withdrawTx.Flatten()
		_, err = runner.TestTransaction(&flatWithdrawTx, pool.testWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)

		postAmountDrops := xrpDropsFromAny(t, postAmm.Amount)
		postAmount2Value := icaValueFromAny(t, postAmm.Amount2)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops-9, postAmountDrops)
		require.InDelta(t, preAmount2Value-3, postAmount2Value, 2.0)
		require.InDelta(t, preLPTokenValue-5, postLPTokenValue, 1e-6)
	})
}

func TestIntegrationAMMWithdraw_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationAMMWithdraw(t, client)
}

func TestIntegrationAMMWithdraw_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationAMMWithdraw(t, client)
}
