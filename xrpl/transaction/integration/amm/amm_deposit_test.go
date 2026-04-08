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

// xrpDropsFromAny extracts the integer XRP drops value from an amm_info Amount (type any).
// The server returns it as a string like "250".
func xrpDropsFromAny(t *testing.T, v any) int {
	t.Helper()
	s, ok := v.(string)
	require.True(t, ok, "expected XRP amount to be a string, got %T", v)
	n, err := strconv.Atoi(s)
	require.NoError(t, err)
	return n
}

// icaValueFromAny extracts the float value from an amm_info Amount2 (type any).
// The server returns it as a map[string]any with "value" key.
func icaValueFromAny(t *testing.T, v any) float64 {
	t.Helper()
	m, ok := v.(map[string]any)
	require.True(t, ok, "expected ICA amount to be a map, got %T", v)
	val, ok := m["value"].(string)
	require.True(t, ok, "expected ICA value to be a string")
	n, err := strconv.ParseFloat(val, 64)
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

		preAmountDrops := xrpDropsFromAny(t, preAmm.Amount)
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

		postAmountDrops := xrpDropsFromAny(t, postAmm.Amount)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops+1000, postAmountDrops)
		require.Equal(t, preAmm.Amount2, postAmm.Amount2)
		require.InDelta(t, preLPTokenValue+191, postLPTokenValue, 2.0)
	})

	t.Run("pass - deposit with Amount and Amount2", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDropsFromAny(t, preAmm.Amount)
		preAmount2Value := icaValueFromAny(t, preAmm.Amount2)
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

		postAmountDrops := xrpDropsFromAny(t, postAmm.Amount)
		postAmount2Value := icaValueFromAny(t, postAmm.Amount2)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops+100, postAmountDrops)
		require.InDelta(t, preAmount2Value+11, postAmount2Value, 2.0)
		require.InDelta(t, preLPTokenValue+34, postLPTokenValue, 2.0)
	})

	t.Run("pass - deposit with Amount and LPTokenOut", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDropsFromAny(t, preAmm.Amount)
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

		postAmountDrops := xrpDropsFromAny(t, postAmm.Amount)
		postLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)

		require.Equal(t, preAmountDrops+30, postAmountDrops)
		require.Equal(t, preAmm.Amount2, postAmm.Amount2)
		require.InDelta(t, preLPTokenValue+5, postLPTokenValue, 1e-6)
	})

	t.Run("pass - deposit with LPTokenOut", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)

		preAmountDrops := xrpDropsFromAny(t, preAmm.Amount)
		preAmount2Value := icaValueFromAny(t, preAmm.Amount2)
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

		postAmountDrops := xrpDropsFromAny(t, postAmm.Amount)
		postAmount2Value := icaValueFromAny(t, postAmm.Amount2)
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
