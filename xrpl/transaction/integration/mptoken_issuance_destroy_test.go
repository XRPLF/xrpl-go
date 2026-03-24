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

type MPTokenIssuanceDestroyTest struct {
	Name                  string
	MPTokenIssuanceCreate *transaction.MPTokenIssuanceCreate
	ExpectedError         string
}

func mptIssuanceDestroyTest(t *testing.T, client integration.Client) {
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

	tt := []MPTokenIssuanceDestroyTest{
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

			issuanceTokenID := accountObjects.AccountObjects[0]["mpt_issuance_id"].(string)
			deleteTx := transaction.MPTokenIssuanceDestroy{
				BaseTx: transaction.BaseTx{
					Account: sender.GetAddress(),
				},
				MPTokenIssuanceID: issuanceTokenID,
			}
			flattenedDeleteTx := deleteTx.Flatten()
			_, err = runner.TestTransaction(&flattenedDeleteTx, sender, "tesSUCCESS", nil)
			require.NoError(t, err)
			accountObjectsFinal, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: sender.GetAddress(),
				Type:    account.MPTIssuance,
			})
			require.NoError(t, err)
			require.Len(t, accountObjectsFinal.AccountObjects, 0)
		})
	}
}

func TestIntegrationMPTokenIssuanceDestroy_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	mptIssuanceDestroyTest(t, client)
}

func TestIntegrationMPTokenIssuanceDestroy_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	mptIssuanceDestroyTest(t, client)
}
