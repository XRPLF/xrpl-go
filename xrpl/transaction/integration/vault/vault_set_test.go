package vault

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

type vaultSetTest struct {
	Name        string
	VaultCreate *transaction.VaultCreate
	VaultSet    *transaction.VaultSet
}

func integrationTestVaultSet(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	updatedData := types.Data(hex.EncodeToString([]byte("updated vault metadata")))
	updatedMaximum := types.XRPLNumber("3000000")

	tt := []vaultSetTest{
		{
			Name: "update vault Data and AssetsMaximum",
			VaultCreate: &transaction.VaultCreate{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
				Asset:  ledger.Asset{Currency: "XRP"},
			},
			VaultSet: &transaction.VaultSet{
				BaseTx:        transaction.BaseTx{Account: owner.GetAddress()},
				Data:          &updatedData,
				AssetsMaximum: &updatedMaximum,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			flatVaultCreateTx := tc.VaultCreate.Flatten()
			_, err = runner.TestTransaction(&flatVaultCreateTx, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			vaultObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
				Type:    account.VaultObject,
			})
			require.NoError(t, err)
			require.Len(t, vaultObjects.AccountObjects, 1)

			tc.VaultSet.VaultID = types.Hash256(vaultObjects.AccountObjects[0]["index"].(string))
			flatVaultSetTx := tc.VaultSet.Flatten()
			_, err = runner.TestTransaction(&flatVaultSetTx, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
				Type:    account.VaultObject,
			})
			require.NoError(t, err)

			vault := objects.AccountObjects[0]
			require.Equal(t, strings.ToLower(string(updatedData)), strings.ToLower(vault["Data"].(string)))
			require.Equal(t, "3000000", vault["AssetsMaximum"].(string))
		})
	}
}

func TestIntegrationVaultSet_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	integrationTestVaultSet(t, client)
}

func TestIntegrationVaultSet_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	integrationTestVaultSet(t, client)
}
