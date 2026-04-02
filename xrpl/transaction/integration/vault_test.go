package integration

import (
	"encoding/hex"
	"fmt"
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

// ############################################################################
// VaultCreate
// ############################################################################

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

// ############################################################################
// VaultDeposit
// ############################################################################

type vaultDepositTest struct {
	Name         string
	VaultCreate  *transaction.VaultCreate
	VaultDeposit *transaction.VaultDeposit
}

func integrationTestVaultDeposit(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []vaultDepositTest{
		{
			Name: "deposit XRP into vault",
			VaultCreate: &transaction.VaultCreate{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
				Asset:  ledger.Asset{Currency: "XRP"},
			},
			VaultDeposit: &transaction.VaultDeposit{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
				Amount: types.XRPCurrencyAmount(1000000),
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

			objects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
				Type:    account.VaultObject,
			})
			require.NoError(t, err)
			require.Equal(t, "1000000", objects.AccountObjects[0]["AssetsTotal"].(string))
		})
	}
}

func TestIntegrationVaultDeposit_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	integrationTestVaultDeposit(t, client)
}

func TestIntegrationVaultDeposit_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	integrationTestVaultDeposit(t, client)
}

// ############################################################################
// VaultWithdraw
// ############################################################################

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

// ############################################################################
// VaultSet
// ############################################################################

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

// ############################################################################
// VaultDelete
// ############################################################################

type vaultDeleteTest struct {
	Name        string
	VaultCreate *transaction.VaultCreate
	VaultDelete *transaction.VaultDelete
}

func integrationTestVaultDelete(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	owner := runner.GetWallet(0)

	tt := []vaultDeleteTest{
		{
			Name: "delete empty vault",
			VaultCreate: &transaction.VaultCreate{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
				Asset:  ledger.Asset{Currency: "XRP"},
			},
			VaultDelete: &transaction.VaultDelete{
				BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
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

			tc.VaultDelete.VaultID = types.Hash256(vaultObjects.AccountObjects[0]["index"].(string))
			flatVaultDeleteTx := tc.VaultDelete.Flatten()
			_, err = runner.TestTransaction(&flatVaultDeleteTx, owner, "tesSUCCESS", nil)
			require.NoError(t, err)

			objects, err := client.GetAccountObjects(&account.ObjectsRequest{
				Account: owner.GetAddress(),
				Type:    account.VaultObject,
			})
			require.NoError(t, err)
			require.Empty(t, objects.AccountObjects)
		})
	}
}

func TestIntegrationVaultDelete_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	integrationTestVaultDelete(t, client)
}

func TestIntegrationVaultDelete_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	integrationTestVaultDelete(t, client)
}

// ############################################################################
// VaultClawback
// ############################################################################

func integrationTestVaultClawback(t *testing.T, client integration.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	issuer := runner.GetWallet(0)
	vaultOwner := runner.GetWallet(1)
	holder := runner.GetWallet(2)
	t.Run("pass - vault clawback test", func(t *testing.T) {
		issuerAccountSetDefaultRippleTx := &transaction.AccountSet{
			BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
		}
		issuerAccountSetDefaultRippleTx.SetFlag = transaction.AsfDefaultRipple
		flatIssuerAccountSetDefaultRippleTx := issuerAccountSetDefaultRippleTx.Flatten()
		_, err = runner.TestTransaction(&flatIssuerAccountSetDefaultRippleTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		issuerAccountSetAllowTrustLineClawbackTx := &transaction.AccountSet{
			BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
		}
		issuerAccountSetAllowTrustLineClawbackTx.SetFlag = transaction.AsfAllowTrustLineClawback
		flatIssuerAccountSetAllowTrustLineClawbackTx := issuerAccountSetAllowTrustLineClawbackTx.Flatten()
		_, err = runner.TestTransaction(&flatIssuerAccountSetAllowTrustLineClawbackTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		setTrustLineTx := &transaction.TrustSet{
			BaseTx: transaction.BaseTx{Account: holder.GetAddress()},
			LimitAmount: types.IssuedCurrencyAmount{
				Currency: "USD",
				Issuer:   issuer.GetAddress(),
				Value:    "9999999999",
			},
		}
		flatSetTrustLineTx := setTrustLineTx.Flatten()
		_, err = runner.TestTransaction(&flatSetTrustLineTx, holder, "tesSUCCESS", nil)
		require.NoError(t, err)

		paymentTx := &transaction.Payment{
			BaseTx:      transaction.BaseTx{Account: issuer.GetAddress()},
			Destination: holder.GetAddress(),
			Amount:      types.IssuedCurrencyAmount{Currency: "USD", Issuer: issuer.GetAddress(), Value: "100"},
		}
		flatPaymentTx := paymentTx.Flatten()
		_, err = runner.TestTransaction(&flatPaymentTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		vaultCreateTx := &transaction.VaultCreate{
			BaseTx: transaction.BaseTx{Account: vaultOwner.GetAddress()},
			Asset:  ledger.Asset{Currency: "USD", Issuer: issuer.GetAddress()},
		}
		flatVaultCreateTx := vaultCreateTx.Flatten()
		_, err = runner.TestTransaction(&flatVaultCreateTx, vaultOwner, "tesSUCCESS", nil)
		fmt.Println(err)
		require.NoError(t, err)

		vaultObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
			Account: vaultOwner.GetAddress(),
			Type:    account.VaultObject,
		})
		require.NoError(t, err)
		require.Len(t, vaultObjects.AccountObjects, 1)

		vaultID := types.Hash256(vaultObjects.AccountObjects[0]["index"].(string))
		vaultDepositTx := &transaction.VaultDeposit{
			BaseTx:  transaction.BaseTx{Account: holder.GetAddress()},
			VaultID: vaultID,
			Amount:  types.IssuedCurrencyAmount{Currency: "USD", Issuer: issuer.GetAddress(), Value: "10"},
		}
		flatVaultDepositTx := vaultDepositTx.Flatten()
		_, err = runner.TestTransaction(&flatVaultDepositTx, holder, "tesSUCCESS", nil)
		require.NoError(t, err)

		vaultClawbackTx := &transaction.VaultClawback{
			BaseTx:  transaction.BaseTx{Account: issuer.GetAddress()},
			VaultID: vaultID,
			Holder:  holder.GetAddress(),
			Amount:  types.IssuedCurrencyAmount{Currency: "USD", Issuer: issuer.GetAddress(), Value: "10"},
		}
		flatVaultClawbackTx := vaultClawbackTx.Flatten()
		_, err = runner.TestTransaction(&flatVaultClawbackTx, issuer, "tesSUCCESS", nil)
		require.NoError(t, err)

		vaultObjects, err = client.GetAccountObjects(&account.ObjectsRequest{
			Account: vaultOwner.GetAddress(),
			Type:    account.VaultObject,
		})
		require.NoError(t, err)
		require.NotContains(t, vaultObjects.AccountObjects[0], "AssetsTotal")
	})
}

func TestIntegrationVaultClawback_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	integrationTestVaultClawback(t, client)
}

func TestIntegrationVaultClawback_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	integrationTestVaultClawback(t, client)
}
