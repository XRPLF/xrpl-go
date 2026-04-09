package amm

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationAMMClawback(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	pool := createAMMPool(t, runner, client, true)
	holderWallet := pool.lpWallet

	t.Run("pass - base", func(t *testing.T) {
		depositTx := &transaction.AMMDeposit{
			BaseTx: transaction.BaseTx{
				Account: holderWallet.GetAddress(),
			},
			Asset:  ledger.Asset{Currency: "USD", Issuer: pool.issuerWallet.GetAddress()},
			Asset2: ledger.Asset{Currency: "XRP"},
			Amount: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   pool.issuerWallet.GetAddress(),
				Value:    "10",
			},
		}
		depositTx.SetSingleAssetFlag()
		flatDepositTx := depositTx.Flatten()
		_, err := runner.TestTransaction(&flatDepositTx, holderWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		clawbackTx := &transaction.AMMClawback{
			BaseTx: transaction.BaseTx{
				Account: pool.issuerWallet.GetAddress(),
			},
			Holder: holderWallet.GetAddress().String(),
			Asset: types.IssuedCurrency{
				Currency: "USD",
				Issuer:   pool.issuerWallet.GetAddress(),
			},
			// XRP has no issuer/value; IssuedCurrencyAmount with only Currency set
			// serializes to {"currency":"XRP"}, which is the correct asset specifier.
			Asset2: types.IssuedCurrencyAmount{Currency: "XRP"},
		}
		flatAMMClawbackTx := clawbackTx.Flatten()
		_, err = runner.TestTransaction(&flatAMMClawbackTx, pool.issuerWallet, "tesSUCCESS", nil)
		require.NoError(t, err)
	})
}

func TestIntegrationAMMClawback_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationAMMClawback(t, client)
}

func TestIntegrationAMMClawback_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationAMMClawback(t, client)
}
