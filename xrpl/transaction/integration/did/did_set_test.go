package integration

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type DIDSetTest struct {
	Name   string
	DIDSet *transaction.DIDSet
}

func integrationTestDIDSet(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	wallet := runner.GetWallet(0)
	tt := []DIDSetTest{
		{
			Name: "pass - base",
			DIDSet: &transaction.DIDSet{
				BaseTx:      transaction.BaseTx{Account: wallet.GetAddress()},
				Data:        "617474657374",
				DIDDocument: "646F63",
				URI:         "6469645F6578616D706C65",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatDIDSetTx := tc.DIDSet.Flatten()
			_, err := runner.TestTransaction(&flatDIDSetTx, wallet, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: wallet.GetAddress(),
				Type:    account.DIDObject,
			})
			require.NoError(t, err)
			require.Len(t, objects.AccountObjects, 1, "should be exactly one DID on the ledger after DIDSet")
		})
	}
}

func TestIntegrationDIDSet_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	integrationTestDIDSet(t, client)
}

func TestIntegrationDIDSet_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	integrationTestDIDSet(t, client)
}
