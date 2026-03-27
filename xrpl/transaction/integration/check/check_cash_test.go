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

type CheckCashTest struct {
	Name          string
	CheckCreate   *transaction.CheckCreate
	ExpectedError string
}

func checkCashTest(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	amount := types.XRPCurrencyAmount(500)
	tt := []CheckCashTest{
		{
			Name: "pass - base",
			CheckCreate: &transaction.CheckCreate{
				BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
				Destination: receiver.GetAddress(),
				SendMax:     amount,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flat := tc.CheckCreate.Flatten()
			_, err := runner.TestTransaction(&flat, sender, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: sender.GetAddress(),
				Type:    account.CheckObject,
			})
			require.NoError(t, err)
			require.Len(t, objects.AccountObjects, 1, "there should be exactly one check on the ledger")
			checkID := objects.AccountObjects[0]["index"].(string)

			checkCashTx := &transaction.CheckCash{
				BaseTx:  transaction.BaseTx{Account: receiver.GetAddress()},
				CheckID: types.Hash256(checkID),
				Amount:  amount,
			}
			flat = checkCashTx.Flatten()
			_, err = runner.TestTransaction(&flat, receiver, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err = client.GetAccountObjects(&account.ObjectsRequest{
				Account: sender.GetAddress(),
				Type:    account.CheckObject,
			})
			require.NoError(t, err)
			require.Empty(t, objects.AccountObjects, "there should be no checks on the ledger")
		})
	}
}

func TestIntegrationCheckCash_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	checkCashTest(t, client)
}

func TestIntegrationCheckCash_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	checkCashTest(t, client)
}
