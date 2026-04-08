package vault

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type vaultCreateTest struct {
	Name        string
	VaultCreate *transaction.VaultCreate
}

func integrationTestVaultCreate(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []vaultCreateTest{
		{
			Name: "pass - base vault create",
			VaultCreate: &transaction.VaultCreate{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
				Asset:  ledger.Asset{Currency: "XRP"},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatVaultCreateTx := tc.VaultCreate.Flatten()
			_, err := runner.TestTransaction(&flatVaultCreateTx, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			vaultObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
				Type:    account.VaultObject,
			})
			require.NoError(t, err)
			require.Len(t, vaultObjects.AccountObjects, 1)

			vault := vaultObjects.AccountObjects[0]
			require.Equal(t, string(owner.GetAddress()), vault["Owner"].(string))
			require.Equal(t, tc.VaultCreate.Asset.Currency, vault["Asset"].(map[string]any)["currency"].(string))
		})
	}
}

func TestIntegrationVaultCreate_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	integrationTestVaultCreate(t, client)
}

func TestIntegrationVaultCreate_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	integrationTestVaultCreate(t, client)
}
