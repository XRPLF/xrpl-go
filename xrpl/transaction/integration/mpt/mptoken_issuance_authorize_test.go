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

type MPTokenIssuanceAuthorizeTest struct {
	Name                  string
	MPTokenIssuanceCreate *transaction.MPTokenIssuanceCreate
}

func testIntegrationMptTokenIssuanceAuthorize(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	encodedMetadata, err := testIntegrationMptTokenCreationMetadata()
	require.NoError(t, err)
	assetScale := uint8(2)
	maxAmount := types.XRPCurrencyAmount(1)

	tt := []MPTokenIssuanceAuthorizeTest{
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
			tc.MPTokenIssuanceCreate.SetMPTRequireAuthFlag()
			flatIssuanceCreateTx := tc.MPTokenIssuanceCreate.Flatten()
			_, err := runner.TestTransaction(&flatIssuanceCreateTx, sender, "tesSUCCESS", nil)
			require.NoError(t, err)

			accountObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: sender.GetAddress(),
				Type:    account.MPTIssuance,
			})
			require.NoError(t, err)
			require.Len(t, accountObjects.AccountObjects, 1)

			issuanceTokenID := accountObjects.AccountObjects[0]["mpt_issuance_id"].(string)
			authorizeTx := transaction.MPTokenAuthorize{
				BaseTx: transaction.BaseTx{
					Account: receiver.GetAddress(),
				},
				MPTokenIssuanceID: issuanceTokenID,
			}
			flatAuthorizeTx := authorizeTx.Flatten()
			_, err = runner.TestTransaction(&flatAuthorizeTx, receiver, "tesSUCCESS", nil)
			require.NoError(t, err)

			receiverObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: receiver.GetAddress(),
				Type:    account.MPToken,
			})
			require.NoError(t, err)
			require.Len(t, receiverObjects.AccountObjects, 1)

			receiverAddress := receiver.GetAddress()
			authorizeTx = transaction.MPTokenAuthorize{
				BaseTx: transaction.BaseTx{
					Account: sender.GetAddress(),
				},
				MPTokenIssuanceID: issuanceTokenID,
				Holder:            &receiverAddress,
			}
			flatAuthorizeTx = authorizeTx.Flatten()
			_, err = runner.TestTransaction(&flatAuthorizeTx, sender, "tesSUCCESS", nil)
			require.NoError(t, err)

			unauthorizeTx := &transaction.MPTokenAuthorize{
				BaseTx: transaction.BaseTx{
					Account: receiver.GetAddress(),
				},
				MPTokenIssuanceID: issuanceTokenID,
			}
			unauthorizeTx.SetMPTUnauthorizeFlag()
			unauthorizeFlatTx := unauthorizeTx.Flatten()
			_, err = runner.TestTransaction(&unauthorizeFlatTx, receiver, "tesSUCCESS", nil)
			require.NoError(t, err)

			finalReceiverObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: receiver.GetAddress(),
				Type:    account.MPToken,
			})
			require.NoError(t, err)
			require.Empty(t, finalReceiverObjects.AccountObjects)
		})
	}
}

func TestIntegrationMPTokenIssuanceAuthorize_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationMptTokenIssuanceAuthorize(t, client)
}

func TestIntegrationMPTokenIssuanceAuthorize_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationMptTokenIssuanceAuthorize(t, client)
}
