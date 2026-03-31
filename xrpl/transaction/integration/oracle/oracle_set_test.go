package oracle

import (
	"encoding/hex"
	"strings"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/time"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func integrationTestOracleSet(t *testing.T, client integration.Client) {
	t.Run("pass - base oracle set", func(t *testing.T) {
		runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
		err := runner.Setup()
		require.NoError(t, err)
		defer runner.Teardown()

		owner := runner.GetWallet(0)
		closeTime := getLedgerCloseTime(t, client)

		oracleSetTx := &transaction.OracleSet{
			BaseTx:           transaction.BaseTx{Account: owner.GetAddress()},
			OracleDocumentID: 1234,
			LastUpdateTime:   uint32(time.RippleTimeToUnixTime(int64(closeTime))/1000) + 20,
			PriceDataSeries: []ledger.PriceDataWrapper{
				{
					PriceData: ledger.PriceData{
						BaseAsset:  "XRP",
						QuoteAsset: "USD",
						AssetPrice: 740,
						Scale:      3,
					},
				},
				{
					PriceData: ledger.PriceData{
						BaseAsset:  "XRP",
						QuoteAsset: "INR",
						AssetPrice: 0xffffffffffffffff,
						Scale:      3,
					},
				},
			},
			Provider:   hex.EncodeToString([]byte("chainlink")),
			URI:        "6469645F6578616D706C65",
			AssetClass: hex.EncodeToString([]byte("currency")),
		}

		flatOracleSetTx := oracleSetTx.Flatten()
		_, err = runner.TestTransaction(&flatOracleSetTx, owner, "tesSUCCESS", nil)
		require.NoError(t, err)

		objects, err := client.GetAccountObjects(&account.ObjectsRequest{
			Account: owner.GetAddress(),
			Type:    account.OracleObject,
		})
		require.NoError(t, err)
		require.Len(t, objects.AccountObjects, 1, "there should be exactly one oracle on the ledger")

		oracle := objects.AccountObjects[0]
		require.Equal(t, oracleSetTx.LastUpdateTime, txFieldUint32(t, oracle, "LastUpdateTime"))
		require.Equal(t, string(owner.GetAddress()), oracle["Owner"].(string))
		require.Equal(t, strings.ToLower(oracleSetTx.AssetClass), strings.ToLower(oracle["AssetClass"].(string)))
		require.Equal(t, strings.ToLower(oracleSetTx.Provider), strings.ToLower(oracle["Provider"].(string)))

		priceDataSeries := oracle["PriceDataSeries"].([]any)
		require.Len(t, priceDataSeries, 2)

		firstPriceData := priceDataSeries[0].(map[string]any)["PriceData"].(map[string]any)
		require.Equal(t, "XRP", firstPriceData["BaseAsset"].(string))
		require.Equal(t, "USD", firstPriceData["QuoteAsset"].(string))
		require.Equal(t, "2e4", firstPriceData["AssetPrice"].(string))
		require.Equal(t, float64(3), txFieldFloat64(t, firstPriceData, "Scale"))
		require.Equal(t, float64(0), txFieldFloat64(t, oracle, "Flags"))

		secondPriceData := priceDataSeries[1].(map[string]any)["PriceData"].(map[string]any)
		require.Equal(t, "ffffffffffffffff", secondPriceData["AssetPrice"].(string))
	})
}

func TestIntegrationOracleSet_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	integrationTestOracleSet(t, client)
}

func TestIntegrationOracleSet_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	integrationTestOracleSet(t, client)
}
