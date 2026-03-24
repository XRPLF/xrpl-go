package integration

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type MPTokenIssuanceCreateTest struct {
	Name                  string
	MPTokenIssuanceCreate *transaction.MPTokenIssuanceCreate
	ExpectedError         string
}

func mtpIntegrationTtestMetadata() (string, error) {
	assetSubclass := "treasury"
	desc := "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments."
	metadata := types.ParsedMPTokenMetadata{
		Ticker:        "TBILL",
		Name:          "T-Bill Yield Token",
		Desc:          &desc,
		Icon:          "example.org/tbill-icon.png",
		AssetClass:    "rwa",
		AssetSubclass: &assetSubclass,
		IssuerName:    "Example Yield Co.",
		URIs: []types.ParsedMPTokenMetadataURI{
			{
				URI:      "exampleyield.co/tbill",
				Category: "website",
				Title:    "Product Page",
			},
			{
				URI:      "exampleyield.co/docs",
				Category: "docs",
				Title:    "Yield Token Docs",
			},
		},
		AdditionalInfo: map[string]any{
			"interest_rate": "5.00%",
			"interest_type": "variable",
			"yield_source":  "U.S. Treasury Bills",
			"maturity_date": "2045-06-30",
			"cusip":         "912796RX0",
		},
	}

	return types.EncodeMPTokenMetadata(metadata)
}

func mtpIntegrationTestCreate(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 1,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	encodedMetadata, err := mtpIntegrationTtestMetadata()
	require.NoError(t, err)
	assetScale := uint8(2)
	maxAmount := types.XRPCurrencyAmount(1)

	tt := []MPTokenIssuanceCreateTest{
		{
			Name: "pass - base",
			MPTokenIssuanceCreate: &transaction.MPTokenIssuanceCreate{
				BaseTx: transaction.BaseTx{
					Account: sender.GetAddress(),
				},
				MaximumAmount:   &maxAmount,
				AssetScale:      &assetScale,
				MPTokenMetadata: &encodedMetadata,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatTx := tc.MPTokenIssuanceCreate.Flatten()
			_, err := runner.TestTransaction(&flatTx, sender, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
				return
			}
			require.NoError(t, err)

			accountObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: sender.GetAddress(),
				Type:    account.MPTIssuance,
			})
			require.NoError(t, err)
			require.Len(t, accountObjects.AccountObjects, 1)
		})
	}
}

func TestIntegrationMPTokenIssuanceCreate_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	mtpIntegrationTestCreate(t, client)
}

func TestIntegrationMPTokenIssuanceCreate_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	mtpIntegrationTestCreate(t, client)
}
