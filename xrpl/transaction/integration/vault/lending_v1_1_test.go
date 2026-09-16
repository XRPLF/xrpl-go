package vault

import (
	"math"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	queryledger "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

// These scenarios require LendingProtocolV1_1, Credentials, DepositAuth,
// DepositPreauth, and fixCleanup3_4_0. The pinned localnet config enables them, rather than
// silently skipping scenarios on an incompatible node.
func testClosedVaultLifecycle(t *testing.T, client integration.Client) {
	t.Helper()
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	require.NoError(t, runner.Setup())
	defer runner.Teardown()
	owner := runner.GetWallet(0)

	latest, err := client.GetLedger(&queryledger.Request{LedgerIndex: common.Validated})
	require.NoError(t, err)
	require.GreaterOrEqual(t, latest.Ledger.CloseTime, 0)
	require.LessOrEqual(t, int64(latest.Ledger.CloseTime), int64(math.MaxUint32-7200))
	subscription := uint32(latest.Ledger.CloseTime) + 3600 //nolint:gosec // bounded above
	redemption := subscription + 3600
	kind := types.VaultKindClosed
	create := transaction.VaultCreate{
		BaseTx:           transaction.BaseTx{Account: owner.GetAddress()},
		Asset:            ledger.Asset{Currency: "XRP"},
		VaultKind:        &kind,
		SubscriptionDate: &subscription,
		RedemptionDate:   &redemption,
	}
	flat := create.Flatten()
	_, err = runner.TestSuccessfulTransactionAndWait(&flat, owner, nil)
	require.NoError(t, err)

	objects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: owner.GetAddress(), Type: account.VaultObject, LedgerIndex: common.Validated,
	})
	require.NoError(t, err)
	require.Len(t, objects.AccountObjects, 1)
	vault := integration.DecodeLedgerObject[ledger.Vault](t, objects.AccountObjects[0])
	require.Equal(t, &kind, vault.VaultKind)
	require.Equal(t, &subscription, vault.SubscriptionDate)
	require.Equal(t, &redemption, vault.RedemptionDate)
	require.NotNil(t, vault.LEVersion)
	require.Equal(t, uint8(1), *vault.LEVersion)

	metadata := "636C6F73696E67207661756C74"
	deletion := transaction.VaultDelete{
		BaseTx: transaction.BaseTx{Account: owner.GetAddress()}, VaultID: vault.Index, MemoData: &metadata,
	}
	flat = deletion.Flatten()
	response, err := runner.TestSuccessfulTransactionAndWait(&flat, owner, nil)
	require.NoError(t, err)
	require.Equal(t, metadata, response.TxJSON["MemoData"])
	objects, err = client.GetAccountObjects(&account.ObjectsRequest{
		Account: owner.GetAddress(), Type: account.VaultObject, LedgerIndex: common.Validated,
	})
	require.NoError(t, err)
	require.Empty(t, objects.AccountObjects)
}

func testVaultWithdrawalCredentials(t *testing.T, client integration.Client) {
	t.Helper()
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	require.NoError(t, runner.Setup())
	defer runner.Teardown()
	owner, depositor, receiver := runner.GetWallet(0), runner.GetWallet(1), runner.GetWallet(2)
	submit := func(flat transaction.FlatTransaction, signer *wallet.Wallet) {
		t.Helper()
		_, err := runner.TestSuccessfulTransactionAndWait(&flat, signer, nil)
		require.NoError(t, err)
	}

	credentialType := types.CredentialType("6C702D6B7963")
	createCredential := transaction.CredentialCreate{
		BaseTx: transaction.BaseTx{Account: owner.GetAddress()}, Subject: depositor.GetAddress(), CredentialType: credentialType,
	}
	submit(createCredential.Flatten(), owner)
	acceptCredential := transaction.CredentialAccept{
		BaseTx: transaction.BaseTx{Account: depositor.GetAddress()}, Issuer: owner.GetAddress(), CredentialType: credentialType,
	}
	submit(acceptCredential.Flatten(), depositor)

	// Only the receiver's deposit authorization requires CredentialIDs. A private
	// vault's domain does not require them for a withdrawal to the depositor.
	depositAuth := transaction.AccountSet{BaseTx: transaction.BaseTx{Account: receiver.GetAddress()}}
	depositAuth.SetAsfDepositAuth()
	submit(depositAuth.Flatten(), receiver)
	preauth := transaction.DepositPreauth{
		BaseTx: transaction.BaseTx{Account: receiver.GetAddress()},
		AuthorizeCredentials: []types.AuthorizeCredentialsWrapper{{Credential: types.AuthorizeCredentials{
			Issuer: owner.GetAddress(), CredentialType: credentialType,
		}}},
	}
	submit(preauth.Flatten(), receiver)

	createVault := transaction.VaultCreate{
		BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
		Asset:  ledger.Asset{Currency: "XRP"},
	}
	submit(createVault.Flatten(), owner)
	vaults, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: owner.GetAddress(), Type: account.VaultObject, LedgerIndex: common.Validated,
	})
	require.NoError(t, err)
	require.Len(t, vaults.AccountObjects, 1)
	vaultID, ok := vaults.AccountObjects[0]["index"].(string)
	require.True(t, ok)
	credentials, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: depositor.GetAddress(), Type: account.CredentialObject, LedgerIndex: common.Validated,
	})
	require.NoError(t, err)
	require.Len(t, credentials.AccountObjects, 1)
	credentialID, ok := credentials.AccountObjects[0]["index"].(string)
	require.True(t, ok)

	deposit := transaction.VaultDeposit{
		BaseTx:  transaction.BaseTx{Account: depositor.GetAddress()},
		VaultID: types.Hash256(vaultID), Amount: types.XRPCurrencyAmount(1_000_000),
	}
	submit(deposit.Flatten(), depositor)
	destination := receiver.GetAddress()
	balanceBefore, err := client.GetXrpDropsBalanceValidated(destination)
	require.NoError(t, err)
	withdraw := transaction.VaultWithdraw{
		BaseTx:  transaction.BaseTx{Account: depositor.GetAddress()},
		VaultID: types.Hash256(vaultID), Amount: types.XRPCurrencyAmount(500_000),
		Destination: &destination,
	}

	// Even with the accepted credential and preauthorization on the ledger,
	// omitting CredentialIDs must fail. Wait for the claimed-fee result before
	// submitting again so the next transaction uses the updated sequence.
	flat := withdraw.Flatten()
	require.NoError(t, client.Autofill(&flat))
	blob, _, err := depositor.Sign(flat)
	require.NoError(t, err)
	response, err := client.SubmitTxBlobAndWait(blob, false)
	require.NoError(t, err)
	require.True(t, response.Validated)
	require.Equal(t, "tecNO_PERMISSION", response.Meta.TransactionResult)
	balanceAfterRejection, err := client.GetXrpDropsBalanceValidated(destination)
	require.NoError(t, err)
	require.Equal(t, balanceBefore, balanceAfterRejection)

	// Adding only CredentialIDs must authorize the same withdrawal. This must
	// fail if the IDs are dropped during flattening, signing, or submission.
	withdraw.CredentialIDs = types.CredentialIDs{credentialID}
	submit(withdraw.Flatten(), depositor)
	balanceAfterWithdrawal, err := client.GetXrpDropsBalanceValidated(destination)
	require.NoError(t, err)
	require.Equal(t, balanceBefore+types.XRPCurrencyAmount(500_000), balanceAfterWithdrawal)
	vaults, err = client.GetAccountObjects(&account.ObjectsRequest{
		Account: owner.GetAddress(), Type: account.VaultObject, LedgerIndex: common.Validated,
	})
	require.NoError(t, err)
	require.Len(t, vaults.AccountObjects, 1)
	require.Equal(t, "500000", vaults.AccountObjects[0]["AssetsTotal"])
}

func TestIntegrationClosedVaultLifecycle_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	config, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	testClosedVaultLifecycle(t, rpc.NewClient(config))
}

func TestIntegrationClosedVaultLifecycle_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testClosedVaultLifecycle(t, client)
}

func TestIntegrationVaultWithdrawalCredentials_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	config, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	testVaultWithdrawalCredentials(t, rpc.NewClient(config))
}

func TestIntegrationVaultWithdrawalCredentials_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testVaultWithdrawalCredentials(t, client)
}
