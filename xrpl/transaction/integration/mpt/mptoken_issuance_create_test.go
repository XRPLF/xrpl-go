package mpt

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
}

func testIntegrationMptTokenCreate(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 1,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	encodedMetadata, err := testIntegrationMptTokenCreationMetadata()
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
			flatIssuanceCreateTx := tc.MPTokenIssuanceCreate.Flatten()
			_, err := runner.TestTransaction(&flatIssuanceCreateTx, sender, "tesSUCCESS", nil)
			require.NoError(t, err)

			accountObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: sender.GetAddress(),
				Type:    account.MPTIssuanceObject,
			})
			require.NoError(t, err)
			require.Len(t, accountObjects.AccountObjects, 1)
			createdToken := accountObjects.AccountObjects[0]

			createdAssetScale := integration.TxFieldUint32(t, createdToken, "AssetScale")
			require.Equal(t, uint8(createdAssetScale), assetScale)

			createdMaximumAmount := createdToken["MaximumAmount"]
			require.Equal(t, createdMaximumAmount, maxAmount.String())
			require.Equal(t, createdToken["MPTokenMetadata"], encodedMetadata)
		})
	}
}

func TestIntegrationMPTokenIssuanceCreate_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationMptTokenCreate(t, client)
}

func TestIntegrationMPTokenIssuanceCreate_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationMptTokenCreate(t, client)
}
