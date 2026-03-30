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

type CheckCancelTest struct {
	Name        string
	CheckCreate *transaction.CheckCreate
}

func integrationTestCheckCancel(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 2})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	receiver := runner.GetWallet(1)

	tt := []CheckCancelTest{
		{
			Name: "pass - base check cancel",
			CheckCreate: &transaction.CheckCreate{
				BaseTx:      transaction.BaseTx{Account: sender.GetAddress()},
				Destination: receiver.GetAddress(),
				SendMax:     types.XRPCurrencyAmount(50),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatCheckCreateTx := tc.CheckCreate.Flatten()
			_, err := runner.TestTransaction(&flatCheckCreateTx, sender, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: sender.GetAddress(),
				Type:    account.CheckObject,
			})
			require.NoError(t, err)
			require.Len(t, objects.AccountObjects, 1, "should be exactly one check on the ledger")
			checkID := objects.AccountObjects[0]["index"].(string)

			checkCancelTx := &transaction.CheckCancel{
				BaseTx:  transaction.BaseTx{Account: sender.GetAddress()},
				CheckID: types.Hash256(checkID),
			}
			flatCheckCancelTx := checkCancelTx.Flatten()
			_, err = runner.TestTransaction(&flatCheckCancelTx, sender, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err = client.GetAccountObjects(&account.ObjectsRequest{
				Account: sender.GetAddress(),
				Type:    account.CheckObject,
			})
			require.NoError(t, err)
			require.Empty(t, objects.AccountObjects, "should be no checks on the ledger")
		})
	}
}

func TestIntegrationCheckCancel_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	integrationTestCheckCancel(t, client)
}

func TestIntegrationCheckCancel_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	integrationTestCheckCancel(t, client)
}
