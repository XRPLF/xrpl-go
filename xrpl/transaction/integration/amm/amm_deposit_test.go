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

// xrpDrops returns the XRP drops value from an XRPCurrencyAmount.
func xrpDrops(t *testing.T, v types.CurrencyAmount) int {
	t.Helper()
	xrp, ok := v.(types.XRPCurrencyAmount)
	require.True(t, ok, "expected XRP amount, got %T", v)
	return int(xrp.Uint64())
}

// icaValue returns the numeric value from an IssuedCurrencyAmount.
func icaValue(t *testing.T, v types.CurrencyAmount) float64 {
	t.Helper()
	ica, ok := v.(types.IssuedCurrencyAmount)
	require.True(t, ok, "expected IssuedCurrencyAmount, got %T", v)
	n, err := strconv.ParseFloat(ica.Value, 64)
	require.NoError(t, err)
	return n
}

func testIntegrationAMMDeposit(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 3,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	pool := setupAMMPool(t, runner, client)

	t.Run("pass - deposit with Amount", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDrops(t, preAmm.Amount)
		preLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)

		depositTx := &transaction.AMMDeposit{
			BaseTx: transaction.BaseTx{
				Account: pool.testWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			Amount: types.XRPCurrencyAmount(1000),
		}
		depositTx.SetSingleAssetFlag()
		flatDepositTx := depositTx.Flatten()
		_, err = runner.TestTransaction(&flatDepositTx, pool.testWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)

		postAmountDrops := xrpDrops(t, postAmm.Amount)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops+1000, postAmountDrops)
		require.Equal(t, preAmm.Amount2, postAmm.Amount2)
		require.InDelta(t, preLPTokenValue+191, postLPTokenValue, 2.0)
	})

	t.Run("pass - deposit with Amount and Amount2", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDrops(t, preAmm.Amount)
		preAmount2Value := icaValue(t, preAmm.Amount2)
		preLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)

		depositTx := &transaction.AMMDeposit{
			BaseTx: transaction.BaseTx{
				Account: pool.issuerWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			Amount: types.XRPCurrencyAmount(100),
			Amount2: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   pool.issuerWallet.GetAddress(),
				Value:    "100",
			},
		}
		depositTx.SetTwoAssetFlag()
		flatDepositTx := depositTx.Flatten()
		_, err = runner.TestTransaction(&flatDepositTx, pool.issuerWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)

		postAmountDrops := xrpDrops(t, postAmm.Amount)
		postAmount2Value := icaValue(t, postAmm.Amount2)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops+100, postAmountDrops)
		require.InDelta(t, preAmount2Value+11, postAmount2Value, 2.0)
		require.InDelta(t, preLPTokenValue+34, postLPTokenValue, 2.0)
	})

	t.Run("pass - deposit with Amount and LPTokenOut", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDrops(t, preAmm.Amount)
		preLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)

		lpToken := getLPToken(t, client, pool.testWallet.GetAddress())

		depositTx := &transaction.AMMDeposit{
			BaseTx: transaction.BaseTx{
				Account: pool.testWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			Amount: types.XRPCurrencyAmount(100),
			LPTokenOut: types.IssuedCurrencyAmount{
				Currency: lpToken.Currency,
				Issuer:   lpToken.Account,
				Value:    "5",
			},
		}
		depositTx.SetOneAssetLPTokenFlag()
		flatDepositTx := depositTx.Flatten()
		_, err = runner.TestTransaction(&flatDepositTx, pool.testWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)

		postAmountDrops := xrpDrops(t, postAmm.Amount)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops+30, postAmountDrops)
		require.Equal(t, preAmm.Amount2, postAmm.Amount2)
		require.InDelta(t, preLPTokenValue+5, postLPTokenValue, 1e-6)
	})

	t.Run("pass - deposit with LPTokenOut", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDrops(t, preAmm.Amount)
		preAmount2Value := icaValue(t, preAmm.Amount2)
		preLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)

		lpToken := getLPToken(t, client, pool.lpWallet.GetAddress())

		depositTx := &transaction.AMMDeposit{
			BaseTx: transaction.BaseTx{
				Account: pool.lpWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			LPTokenOut: types.IssuedCurrencyAmount{
				Currency: lpToken.Currency,
				Issuer:   lpToken.Account,
				Value:    "5",
			},
		}
		depositTx.SetLPTokentFlag()
		flatDepositTx := depositTx.Flatten()
		_, err = runner.TestTransaction(&flatDepositTx, pool.lpWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)

		postAmountDrops := xrpDrops(t, postAmm.Amount)
		postAmount2Value := icaValue(t, postAmm.Amount2)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops+15, postAmountDrops)
		require.InDelta(t, preAmount2Value+1, postAmount2Value, 2.0)
		require.InDelta(t, preLPTokenValue+5, postLPTokenValue, 1e-6)
	})
}

func TestIntegrationAMMDeposit_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationAMMDeposit(t, client)
}

func TestIntegrationAMMDeposit_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationAMMDeposit(t, client)
}
