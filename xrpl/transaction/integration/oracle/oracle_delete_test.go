package oracle

import (
	"encoding/hex"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	xrpltime "github.com/Peersyst/xrpl-go/xrpl/time"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type OracleDeleteTest struct {
	Name         string
	OracleSet    *transaction.OracleSet
	OracleDelete *transaction.OracleDelete
}

func testIntegrationOracleDelete(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	closeTime := getLedgerCloseTime(t, client)

	tt := []OracleDeleteTest{
		{
			Name: "pass - base oracle delete",
			OracleSet: &transaction.OracleSet{
				BaseTx:           transaction.BaseTx{Account: owner.GetAddress()},
				OracleDocumentID: 1234,
				LastUpdateTime:   uint32(xrpltime.RippleTimeToUnixSeconds(closeTime)) + 20,
				PriceDataSeries: []ledger.PriceDataWrapper{
					{
						PriceData: ledger.PriceData{
							BaseAsset:  "XRP",
							QuoteAsset: "USD",
							AssetPrice: 740,
							Scale:      3,
						},
					},
				},
				Provider:   hex.EncodeToString([]byte("chainlink")),
				URI:        "6469645F6578616D706C65",
				AssetClass: hex.EncodeToString([]byte("currency")),
			},
			OracleDelete: &transaction.OracleDelete{
				BaseTx:           transaction.BaseTx{Account: owner.GetAddress()},
				OracleDocumentID: 1234,
			},
		},
	}

	for _, tc := range tt {
		t.Run("pass - base oracle delete", func(t *testing.T) {
			flatOracleSetTx := tc.OracleSet.Flatten()
			_, err = runner.TestTransaction(&flatOracleSetTx, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
				Type:    account.OracleObject,
			})
			require.NoError(t, err)
			require.Len(t, objects.AccountObjects, 1, "should be exactly one oracle on the ledger")

			flatOracleDeleteTx := tc.OracleDelete.Flatten()
			_, err = runner.TestTransaction(&flatOracleDeleteTx, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err = client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
				Type:    account.OracleObject,
			})
			require.NoError(t, err)
			require.Empty(t, objects.AccountObjects, "oracle should be deleted from account objects")
		})
	}
}

func TestIntegrationOracleDelete_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationOracleDelete(t, client)
}

func TestIntegrationOracleDelete_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationOracleDelete(t, client)
}
