package vault

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	querycommon "github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func integrationTestVaultClawback(t *testing.T, client integration.Client) {
	t.Run("pass - vault clawback test", func(t *testing.T) {
		runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
		err := runner.Setup()
		require.NoError(t, err)
		defer runner.Teardown()

		issuer := runner.GetWallet(0)
		vaultOwner := runner.GetWallet(1)
		holder := runner.GetWallet(2)
		issuerAccountSetDefaultRippleTx := &transaction.AccountSet{
			BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
		}
		issuerAccountSetDefaultRippleTx.SetFlag = transaction.AsfDefaultRipple
		flatIssuerAccountSetDefaultRippleTx := issuerAccountSetDefaultRippleTx.Flatten()
		_, err = runner.TestSuccessfulTransactionAndWait(&flatIssuerAccountSetDefaultRippleTx, issuer, nil)
		require.NoError(t, err)

		issuerAccountSetAllowTrustLineClawbackTx := &transaction.AccountSet{
			BaseTx: transaction.BaseTx{Account: issuer.GetAddress()},
		}
		issuerAccountSetAllowTrustLineClawbackTx.SetFlag = transaction.AsfAllowTrustLineClawback
		flatIssuerAccountSetAllowTrustLineClawbackTx := issuerAccountSetAllowTrustLineClawbackTx.Flatten()
		_, err = runner.TestSuccessfulTransactionAndWait(&flatIssuerAccountSetAllowTrustLineClawbackTx, issuer, nil)
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
		_, err = runner.TestSuccessfulTransactionAndWait(&flatSetTrustLineTx, holder, nil)
		require.NoError(t, err)

		paymentTx := &transaction.Payment{
			BaseTx:      transaction.BaseTx{Account: issuer.GetAddress()},
			Destination: holder.GetAddress(),
			Amount:      types.IssuedCurrencyAmount{Currency: "USD", Issuer: issuer.GetAddress(), Value: "100"},
		}
		flatPaymentTx := paymentTx.Flatten()
		_, err = runner.TestSuccessfulTransactionAndWait(&flatPaymentTx, issuer, nil)
		require.NoError(t, err)

		vaultCreateTx := &transaction.VaultCreate{
			BaseTx: transaction.BaseTx{Account: vaultOwner.GetAddress()},
			Asset:  ledger.Asset{Currency: "USD", Issuer: issuer.GetAddress()},
		}
		flatVaultCreateTx := vaultCreateTx.Flatten()
		_, err = runner.TestSuccessfulTransactionAndWait(&flatVaultCreateTx, vaultOwner, nil)
		require.NoError(t, err)

		vaultObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
			Account:     vaultOwner.GetAddress(),
			Type:        account.VaultObject,
			LedgerIndex: querycommon.Validated,
		})
		require.NoError(t, err)
		require.True(t, vaultObjects.Validated)
		require.Len(t, vaultObjects.AccountObjects, 1)

		vaultID := types.Hash256(vaultObjects.AccountObjects[0]["index"].(string))
		vaultDepositTx := &transaction.VaultDeposit{
			BaseTx:  transaction.BaseTx{Account: holder.GetAddress()},
			VaultID: vaultID,
			Amount:  types.IssuedCurrencyAmount{Currency: "USD", Issuer: issuer.GetAddress(), Value: "10"},
		}
		flatVaultDepositTx := vaultDepositTx.Flatten()
		_, err = runner.TestSuccessfulTransactionAndWait(&flatVaultDepositTx, holder, nil)
		require.NoError(t, err)

		vaultClawbackTx := &transaction.VaultClawback{
			BaseTx:  transaction.BaseTx{Account: issuer.GetAddress()},
			VaultID: vaultID,
			Holder:  holder.GetAddress(),
			Amount:  types.IssuedCurrencyAmount{Currency: "USD", Issuer: issuer.GetAddress(), Value: "10"},
		}
		flatVaultClawbackTx := vaultClawbackTx.Flatten()
		_, err = runner.TestSuccessfulTransactionAndWait(&flatVaultClawbackTx, issuer, nil)
		require.NoError(t, err)

		vaultObjects, err = client.GetAccountObjects(&account.ObjectsRequest{
			Account:     vaultOwner.GetAddress(),
			Type:        account.VaultObject,
			LedgerIndex: querycommon.Validated,
		})
		require.NoError(t, err)
		require.True(t, vaultObjects.Validated)
		require.Len(t, vaultObjects.AccountObjects, 1)
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
