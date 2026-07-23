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
