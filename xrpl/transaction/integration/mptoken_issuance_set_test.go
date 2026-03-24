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

type MPTokenIssuanceSetTest struct {
	Name                  string
	MPTokenIssuanceCreate *transaction.MPTokenIssuanceCreate
	ExpectedError         string
}

func mptTokenIssuanceSetTest(t *testing.T, client integration.Client) {
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

	tt := []MPTokenIssuanceSetTest{
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
			tc.MPTokenIssuanceCreate.SetMPTCanLockFlag()
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
			setTx := transaction.MPTokenIssuanceSet{
				BaseTx: transaction.BaseTx{
					Account: sender.GetAddress(),
				},
				MPTokenIssuanceID: issuanceTokenID,
			}
			setTx.SetMPTLockFlag()
			flatSetTx := setTx.Flatten()
			_, err = runner.TestTransaction(&flatSetTx, sender, "tesSUCCESS", nil)
			require.NoError(t, err)
		})
	}
}

func TestIntegrationMPTokenIssuanceSet_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	mptTokenIssuanceSetTest(t, client)
}

func TestIntegrationMPTokenIssuanceSet_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	mptTokenIssuanceSetTest(t, client)
}
