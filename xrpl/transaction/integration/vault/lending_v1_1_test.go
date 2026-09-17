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

// Create an empty closed-ended vault, check its ledger fields, then delete it
// with metadata. Requires LendingProtocolV1_1.
func testClosedVaultLifecycle(t *testing.T, client integration.Client) {
	t.Helper()
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 1})
	require.NoError(t, runner.Setup())
	defer runner.Teardown()
	owner := runner.GetWallet(0)

	// Use ledger time, not the host clock. Allow one hour for subscription and
	// one hour for investment so both dates remain valid during this test.
	const subscriptionWindow uint32 = 3600
	const investmentPeriod uint32 = 3600
	validatedLedger, err := client.GetLedger(&queryledger.Request{LedgerIndex: common.Validated})
	require.NoError(t, err)
	closeTime := validatedLedger.Ledger.CloseTime
	require.GreaterOrEqual(t, closeTime, 0)
	require.LessOrEqual(t, int64(closeTime), int64(math.MaxUint32-subscriptionWindow-investmentPeriod))
	subscriptionDate := uint32(closeTime) + subscriptionWindow //nolint:gosec // bounded above
	redemptionDate := subscriptionDate + investmentPeriod

	// Create the vault and wait until it is available in the validated ledger.
	kind := types.VaultKindClosed
	createVault := transaction.VaultCreate{
		BaseTx:           transaction.BaseTx{Account: owner.GetAddress()},
		Asset:            ledger.Asset{Currency: "XRP"},
		VaultKind:        &kind,
		SubscriptionDate: &subscriptionDate,
		RedemptionDate:   &redemptionDate,
	}
	createTx := createVault.Flatten()
	_, err = runner.TestSuccessfulTransactionAndWait(&createTx, owner, nil)
	require.NoError(t, err)

	vaultRequest := &account.ObjectsRequest{
		Account:     owner.GetAddress(),
		Type:        account.VaultObject,
		LedgerIndex: common.Validated,
	}
	vaultObjects, err := client.GetAccountObjects(vaultRequest)
	require.NoError(t, err)
	require.Len(t, vaultObjects.AccountObjects, 1)
	vault := integration.DecodeLedgerObject[ledger.Vault](t, vaultObjects.AccountObjects[0])
	require.Equal(t, &kind, vault.VaultKind)
	require.Equal(t, &subscriptionDate, vault.SubscriptionDate)
	require.Equal(t, &redemptionDate, vault.RedemptionDate)
	require.NotNil(t, vault.LEVersion)
	require.Equal(t, uint8(1), *vault.LEVersion)

	// Deletion must retain the top-level metadata and remove the vault object.
	metadata := "636C6F73696E67207661756C74" // "closing vault"
	deleteVault := transaction.VaultDelete{
		BaseTx:   transaction.BaseTx{Account: owner.GetAddress()},
		VaultID:  vault.Index,
		MemoData: &metadata,
	}
	deleteTx := deleteVault.Flatten()
	deleteResult, err := runner.TestSuccessfulTransactionAndWait(&deleteTx, owner, nil)
	require.NoError(t, err)
	require.Equal(t, metadata, deleteResult.TxJSON["MemoData"])

	vaultObjects, err = client.GetAccountObjects(vaultRequest)
	require.NoError(t, err)
	require.Empty(t, vaultObjects.AccountObjects)
}

// A deposit-authorized receiver must reject a withdrawal without CredentialIDs
// and accept the same withdrawal with them. Requires Credentials, DepositAuth,
// DepositPreauth, and fixCleanup3_4_0.
func testVaultWithdrawalCredentials(t *testing.T, client integration.Client) {
	t.Helper()
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{WalletCount: 3})
	require.NoError(t, runner.Setup())
	defer runner.Teardown()

	owner := runner.GetWallet(0) // Also issues the depositor's credential.
	depositor := runner.GetWallet(1)
	receiver := runner.GetWallet(2)
	receiverAddress := receiver.GetAddress()

	submitAndWait := func(tx transaction.FlatTransaction, signer *wallet.Wallet) {
		t.Helper()
		_, err := runner.TestSuccessfulTransactionAndWait(&tx, signer, nil)
		require.NoError(t, err)
	}

	// Give the depositor an accepted credential from the vault owner.
	credentialType := types.CredentialType("6C702D6B7963") // "lp-kyc"
	createCredential := transaction.CredentialCreate{
		BaseTx:         transaction.BaseTx{Account: owner.GetAddress()},
		Subject:        depositor.GetAddress(),
		CredentialType: credentialType,
	}
	submitAndWait(createCredential.Flatten(), owner)

	acceptCredential := transaction.CredentialAccept{
		BaseTx:         transaction.BaseTx{Account: depositor.GetAddress()},
		Issuer:         owner.GetAddress(),
		CredentialType: credentialType,
	}
	submitAndWait(acceptCredential.Flatten(), depositor)

	// Require this credential to deposit funds into the receiver's account.
	// The vault is public, so only the receiver's deposit authorization is tested.
	depositAuth := transaction.AccountSet{
		BaseTx: transaction.BaseTx{Account: receiverAddress},
	}
	depositAuth.SetAsfDepositAuth()
	submitAndWait(depositAuth.Flatten(), receiver)

	preauthorizeCredential := transaction.DepositPreauth{
		BaseTx: transaction.BaseTx{Account: receiverAddress},
		AuthorizeCredentials: []types.AuthorizeCredentialsWrapper{
			{
				Credential: types.AuthorizeCredentials{
					Issuer:         owner.GetAddress(),
					CredentialType: credentialType,
				},
			},
		},
	}
	submitAndWait(preauthorizeCredential.Flatten(), receiver)

	// Create a public XRP vault and find the ledger IDs used by the withdrawal.
	createVault := transaction.VaultCreate{
		BaseTx: transaction.BaseTx{Account: owner.GetAddress()},
		Asset:  ledger.Asset{Currency: "XRP"},
	}
	submitAndWait(createVault.Flatten(), owner)

	vaultRequest := &account.ObjectsRequest{
		Account:     owner.GetAddress(),
		Type:        account.VaultObject,
		LedgerIndex: common.Validated,
	}
	vaultObjects, err := client.GetAccountObjects(vaultRequest)
	require.NoError(t, err)
	require.Len(t, vaultObjects.AccountObjects, 1)
	vault := integration.DecodeLedgerObject[ledger.Vault](t, vaultObjects.AccountObjects[0])
	require.NotEmpty(t, vault.Index)

	credentialObjects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account:     depositor.GetAddress(),
		Type:        account.CredentialObject,
		LedgerIndex: common.Validated,
	})
	require.NoError(t, err)
	require.Len(t, credentialObjects.AccountObjects, 1)
	credential := integration.DecodeLedgerObject[ledger.Credential](t, credentialObjects.AccountObjects[0])
	require.NotEmpty(t, credential.Index)

	// The depositor funds the vault, then tries to send half to the receiver.
	const depositAmount types.XRPCurrencyAmount = 1_000_000
	const withdrawalAmount types.XRPCurrencyAmount = 500_000
	deposit := transaction.VaultDeposit{
		BaseTx:  transaction.BaseTx{Account: depositor.GetAddress()},
		VaultID: vault.Index,
		Amount:  depositAmount,
	}
	submitAndWait(deposit.Flatten(), depositor)

	balanceBefore, err := client.GetXrpDropsBalanceValidated(receiverAddress)
	require.NoError(t, err)
	withdraw := transaction.VaultWithdraw{
		BaseTx:      transaction.BaseTx{Account: depositor.GetAddress()},
		VaultID:     vault.Index,
		Amount:      withdrawalAmount,
		Destination: &receiverAddress,
	}

	// Without CredentialIDs, the withdrawal must fail even though the credential
	// and preauthorization exist. Wait for validation because the fee is claimed
	// and the depositor's sequence changes before the retry.
	withdrawTx := withdraw.Flatten()
	require.NoError(t, client.Autofill(&withdrawTx))
	withdrawBlob, _, err := depositor.Sign(withdrawTx)
	require.NoError(t, err)
	rejectedWithdrawal, err := client.SubmitTxBlobAndWait(withdrawBlob, false)
	require.NoError(t, err)
	require.True(t, rejectedWithdrawal.Validated)
	require.Equal(t, "tecNO_PERMISSION", rejectedWithdrawal.Meta.TransactionResult)

	balanceAfterRejection, err := client.GetXrpDropsBalanceValidated(receiverAddress)
	require.NoError(t, err)
	require.Equal(t, balanceBefore, balanceAfterRejection)

	// Change only CredentialIDs. The receiver must now get the funds.
	withdraw.CredentialIDs = types.CredentialIDs{credential.Index.String()}
	submitAndWait(withdraw.Flatten(), depositor)

	balanceAfterWithdrawal, err := client.GetXrpDropsBalanceValidated(receiverAddress)
	require.NoError(t, err)
	require.Equal(t, balanceBefore+withdrawalAmount, balanceAfterWithdrawal)

	vaultObjects, err = client.GetAccountObjects(vaultRequest)
	require.NoError(t, err)
	require.Len(t, vaultObjects.AccountObjects, 1)
	vault = integration.DecodeLedgerObject[ledger.Vault](t, vaultObjects.AccountObjects[0])
	require.NotNil(t, vault.AssetsTotal)
	require.Equal(t, types.XRPLNumber("500000"), *vault.AssetsTotal)
}

func TestIntegrationClosedVaultLifecycle_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	config, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	testClosedVaultLifecycle(t, rpc.NewClient(config))
}

func TestIntegrationClosedVaultLifecycle_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	config := websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider)
	testClosedVaultLifecycle(t, websocket.NewClient(config))
}

func TestIntegrationVaultWithdrawalCredentials_RPCClient(t *testing.T) {
	env := integration.GetRPCEnv(t)
	config, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	testVaultWithdrawalCredentials(t, rpc.NewClient(config))
}

func TestIntegrationVaultWithdrawalCredentials_Websocket(t *testing.T) {
	env := integration.GetWebsocketEnv(t)
	config := websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider)
	testVaultWithdrawalCredentials(t, websocket.NewClient(config))
}
