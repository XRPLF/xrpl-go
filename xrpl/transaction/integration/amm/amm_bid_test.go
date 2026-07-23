package amm

import (
	"strconv"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationAMMBid(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 3,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	pool := setupAMMPool(t, runner, client)

	t.Run("pass - bid", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)
		require.NotNil(t, preAmm.AuctionSlot)

		ammBidTx := &transaction.AMMBid{
			BaseTx: transaction.BaseTx{
				Account: pool.testWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
		}
		flatAMMBidTx := ammBidTx.Flatten()
		_, err = runner.TestTransaction(&flatAMMBidTx, pool.testWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)
		require.NotNil(t, postAmm.AuctionSlot)

		afterPriceValue, err := strconv.ParseFloat(postAmm.AuctionSlot.Price.Value, 64)
		require.NoError(t, err)
		beforePriceValue, err := strconv.ParseFloat(preAmm.AuctionSlot.Price.Value, 64)
		require.NoError(t, err)
		// Minimum bid (1/5% of LP tokens); no previous holder so no rebate is issued.
		diffPriceValue := 0.002683192572241211
		require.InDelta(t, beforePriceValue+diffPriceValue, afterPriceValue, 1e-9)

		afterLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)
		beforeLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)
		// No rebate ⇒ burned == paid, so diffLPTokenValue == -diffPriceValue.
		diffLPTokenValue := -diffPriceValue
		require.InDelta(t, beforeLPTokenValue+diffLPTokenValue, afterLPTokenValue, 1e-9)
	})

	t.Run("pass - bid with AuthAccounts, BidMin, BidMax", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)
		require.NotNil(t, preAmm.AuctionSlot)

		lpToken := getLPToken(t, client, pool.testWallet.GetAddress())

		ammBidTx := &transaction.AMMBid{
			BaseTx: transaction.BaseTx{
				Account: pool.testWallet.GetAddress(),
			},
			Asset:  pool.asset,
			Asset2: pool.asset2,
			AuthAccounts: []ledger.AuthAccounts{
				{
					AuthAccount: ledger.AuthAccount{
						Account: pool.issuerWallet.GetAddress(),
					},
				},
			},
			BidMin: types.IssuedCurrencyAmount{
				Currency: lpToken.Currency,
				Issuer:   lpToken.Account,
				Value:    "5",
			},
			BidMax: types.IssuedCurrencyAmount{
				Currency: lpToken.Currency,
				Issuer:   lpToken.Account,
				Value:    "10",
			},
		}
		flatAMMBidTx := ammBidTx.Flatten()
		_, err = runner.TestTransaction(&flatAMMBidTx, pool.testWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)
		require.NotNil(t, postAmm.AuctionSlot)

		afterPriceValue, err := strconv.ParseFloat(postAmm.AuctionSlot.Price.Value, 64)
		require.NoError(t, err)
		beforePriceValue, err := strconv.ParseFloat(preAmm.AuctionSlot.Price.Value, 64)
		require.NoError(t, err)
		// New slot price = BidMin=5; diff is 5 minus the previous slot price (≈ 0.002683).
		diffPriceValue := 4.997316807427759
		require.InDelta(t, beforePriceValue+diffPriceValue, afterPriceValue, 1e-9)

		afterLPTokenValue, err := strconv.ParseFloat(postAmm.LPToken.Value, 64)
		require.NoError(t, err)
		beforeLPTokenValue, err := strconv.ParseFloat(preAmm.LPToken.Value, 64)
		require.NoError(t, err)
		// Previous holder receives a rebate of 19/20 of their price (19 intervals remaining).
		// Burned = 5 − rebate ≈ 4.99745, slightly more than diffPriceValue because the
		// unreturned 1/20th of the old price is also burned.
		diffLPTokenValue := -4.9974509670563
		require.InDelta(t, beforeLPTokenValue+diffLPTokenValue, afterLPTokenValue, 1e-9)

		require.Len(t, postAmm.AuctionSlot.AuthAccounts, 1)
		require.Equal(t, pool.issuerWallet.GetAddress(), postAmm.AuctionSlot.AuthAccounts[0].Account)
	})
}

func TestIntegrationAMMBid_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationAMMBid(t, client)
}

func TestIntegrationAMMBid_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationAMMBid(t, client)
}
