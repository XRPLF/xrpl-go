package integration

import (
	"os"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
	txqueries "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

const (
	accountDeleteLedgerWait    = 256
	accountDeleteLedgerAccepts = accountDeleteLedgerWait + 1
)

type AccountDeleteTest struct {
	Name          string
	AccountDelete *transaction.AccountDelete
	ExpectedError string
}

func testIntegrationAccountDelete(t *testing.T, client integration.Client, ledgerClient *rpc.Client) {
	runner := integration.NewRunner(t, client, &integration.RunnerConfig{
		WalletCount: 2,
	})

	err := runner.Setup()
	require.NoError(t, err)
	defer runner.Teardown()

	sender := runner.GetWallet(0)
	destination := runner.GetWallet(1)

	advanceAccountDeleteLedgerWindow(t, ledgerClient)

	tt := []AccountDeleteTest{
		{
			Name: "pass - delete account and report deleted balance",
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
			deletedBalance, err := ledgerClient.GetXrpDropsBalanceValidated(sender.GetAddress())
			require.NoError(t, err)

			deletedBalanceXRP, err := currency.DropsToXrp(deletedBalance.String())
			require.NoError(t, err)

			expectedBalance := transaction.Balance{
				Value:    "-" + deletedBalanceXRP,
				Currency: "XRP",
			}

			flatTx := tc.AccountDelete.Flatten()
			tx, err := runner.TestTransaction(&flatTx, sender, "tesSUCCESS", nil)
			if tc.ExpectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedError)
				return
			}
			require.NoError(t, err)

			acceptLedger(t, ledgerClient)

			txHash, ok := tx.Tx["hash"].(string)
			require.True(t, ok)

			resp := getValidatedTransaction(t, ledgerClient, txHash)
			require.Equal(t, transaction.TesSUCCESS.String(), resp.Meta.TransactionResult)

			meta := resp.Meta.AsTxObjMeta()
			balanceChanges, err := transaction.GetBalanceChanges(&meta)
			require.NoError(t, err)

			requireAccountBalanceChange(t, balanceChanges, sender.GetAddress(), expectedBalance)
		})
	}
}

func TestIntegrationAccountDelete_Websocket(t *testing.T) {
	requireLocalnetAccountDelete(t)

	env := integration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))

	rpcEnv := integration.GetRPCEnv(t)
	rpcClientCfg, err := rpc.NewClientConfig(rpcEnv.Host, rpc.WithFaucetProvider(rpcEnv.FaucetProvider))
	require.NoError(t, err)
	ledgerClient := rpc.NewClient(rpcClientCfg)

	testIntegrationAccountDelete(t, client, ledgerClient)
}

func TestIntegrationAccountDelete_RPCClient(t *testing.T) {
	requireLocalnetAccountDelete(t)

	env := integration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationAccountDelete(t, client, client)
}

func getValidatedTransaction(t *testing.T, client *rpc.Client, txHash string) txqueries.TxResponse {
	t.Helper()

	res, err := client.Request(&txqueries.TxRequest{
		Transaction: txHash,
	})
	require.NoError(t, err)

	var resp txqueries.TxResponse
	err = res.GetResult(&resp)
	require.NoError(t, err)
	require.True(t, resp.Validated)
	return resp
}

func requireLocalnetAccountDelete(t *testing.T) {
	t.Helper()

	if os.Getenv(integration.IntegrationEnvVar) != string(integration.LocalnetEnv) {
		t.Skip("account delete tests require localnet ledger_accept")
	}
}

func advanceAccountDeleteLedgerWindow(t *testing.T, client *rpc.Client) {
	t.Helper()

	startLedger, err := client.GetLedgerIndex()
	require.NoError(t, err)

	// The first accept can validate setup funding, then the account needs 256 more ledgers.
	for range accountDeleteLedgerAccepts {
		acceptLedger(t, client)
	}

	currentLedger, err := client.GetLedgerIndex()
	require.NoError(t, err)
	require.GreaterOrEqual(t, currentLedger.Uint32(), startLedger.Uint32()+accountDeleteLedgerAccepts)
}

func requireAccountBalanceChange(
	t *testing.T,
	balanceChanges []transaction.AccountBalanceChanges,
	account types.Address,
	balance transaction.Balance,
) {
	t.Helper()

	for _, accountChanges := range balanceChanges {
		if accountChanges.Account == account {
			require.Contains(t, accountChanges.Balances, balance)
			return
		}
	}

	require.Failf(t, "missing account balance change", "account %s was not present", account)
}
