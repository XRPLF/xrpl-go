package amm

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationAMMVote(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 3,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	pool := setupAMMPool(t, runner, client)

	t.Run("pass - vote", func(t *testing.T) {
		preAmm := getAMMInfo(t, client, pool)
		require.NotNil(t, preAmm.AuctionSlot)
		require.NotEmpty(t, preAmm.VoteSlots)

		ammVoteTx := &transaction.AMMVote{
			BaseTx: transaction.BaseTx{
				Account: pool.testWallet.GetAddress(),
			},
			Asset:      pool.asset,
			Asset2:     pool.asset2,
			TradingFee: 150,
		}
		flatAMMVoteTx := ammVoteTx.Flatten()
		_, err = runner.TestTransaction(&flatAMMVoteTx, pool.testWallet, "tesSUCCESS", nil)
		require.NoError(t, err)

		postAmm := getAMMInfo(t, client, pool)
		require.NotNil(t, postAmm.AuctionSlot)
		require.NotEmpty(t, postAmm.VoteSlots)

		diffTradingFee := uint16(76)
		require.Equal(t, preAmm.TradingFee+diffTradingFee, postAmm.TradingFee)

		diffDiscountedFee := uint16(7)
		require.Equal(t, preAmm.AuctionSlot.DiscountedFee+diffDiscountedFee, postAmm.AuctionSlot.DiscountedFee)

		require.Len(t, postAmm.VoteSlots, len(preAmm.VoteSlots)+1)
	})
}

func TestIntegrationAMMVote_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationAMMVote(t, client)
}

func TestIntegrationAMMVote_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationAMMVote(t, client)
}
