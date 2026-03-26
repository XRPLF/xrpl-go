package integration

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type AccountDeleteTest struct {
	Name          string
	AccountDelete *transaction.AccountDelete
	ExpectedError string
}

func testIntegrationAccountDelete(t *testing.T, client integration.Client) {
	// TODO: Re-enable test once the required 256 ledgers between operations can be skipped
	t.Skip("account delete test requires 256 ledgers between operations, skipping")

	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	destination := runner.GetWallet(1)

	tt := []AccountDeleteTest{
		{
			Name: "pass - delete account",
			AccountDelete: &transaction.AccountDelete{
				BaseTx: transaction.BaseTx{
					Account: sender.GetAddress(),
				},
				Destination: destination.GetAddress(),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatTx := tc.AccountDelete.Flatten()
			_, err := runner.TestTransaction(&flatTx, sender, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIntegrationAccountDelete_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationAccountDelete(t, client)
}

func TestIntegrationAccountDelete_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationAccountDelete(t, client)
}
