package vault

import (
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

type vaultWithdrawTest struct {
	Name          string
	VaultCreate   *transaction.VaultCreate
	VaultDeposit  *transaction.VaultDeposit
	VaultWithdraw *transaction.VaultWithdraw
}

func integrationTestVaultWithdraw(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []vaultWithdrawTest{
		{
			Name: "withdraw XRP from vault",
			VaultCreate: &transaction.VaultCreate{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
				Asset:  ledger.Asset{Currency: "XRP"},
			},
			VaultDeposit: &transaction.VaultDeposit{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
				Amount: types.XRPCurrencyAmount(1000000),
			},
			VaultWithdraw: &transaction.VaultWithdraw{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
				Amount: types.XRPCurrencyAmount(500000),
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
			tc.VaultDeposit.VaultID = types.Hash256(vaultObjects.AccountObjects[0]["index"].(string))

			flatVaultDepositTx := tc.VaultDeposit.Flatten()
			_, err = runner.TestTransaction(&flatVaultDepositTx, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			tc.VaultWithdraw.VaultID = tc.VaultDeposit.VaultID
			flatVaultWithdrawTx := tc.VaultWithdraw.Flatten()
			_, err = runner.TestTransaction(&flatVaultWithdrawTx, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
				Type:    account.VaultObject,
			})
			require.NoError(t, err)
			require.Equal(t, "500000", objects.AccountObjects[0]["AssetsTotal"])
		})
	}
}

func TestIntegrationVaultWithdraw_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	integrationTestVaultWithdraw(t, client)
}

func TestIntegrationVaultWithdraw_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	integrationTestVaultWithdraw(t, client)
}
