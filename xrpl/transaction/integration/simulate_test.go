package integration

import (
	"testing"

	transactionsquery "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	testintegration "github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationSimulate(t *testing.T, client testintegration.Client) {
	t.Helper()

	runner := testintegration.NewRunner(t, client, &testintegration.RunnerConfig{WalletCount: 1})
	require.NoError(t, runner.Setup())
	defer runner.Teardown()

	account := runner.GetWallet(0)
	accountSet := &transaction.AccountSet{
		BaseTx:  transaction.BaseTx{Account: account.GetAddress()},
		SetFlag: transaction.AsfDefaultRipple,
	}
	txJSON := accountSet.Flatten()

	tests := []struct {
		name   string
		binary bool
	}{
		{name: "JSON response"},
		{name: "binary response", binary: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.Simulate(&transactionsquery.SimulateRequest{
				TxJSON: txJSON,
				Binary: tt.binary,
			})
			require.NoError(t, err)
			require.False(t, response.Applied)
			require.Equal(t, "tesSUCCESS", response.EngineResult)
			require.Zero(t, response.EngineResultCode)
			require.NotEmpty(t, response.EngineResultMessage)
			require.NotZero(t, response.LedgerIndex)

			if tt.binary {
				require.NotEmpty(t, response.TxBlob)
				require.NotEmpty(t, response.MetaBlob)
				require.Empty(t, response.TxJSON)
				require.Nil(t, response.Meta)
				return
			}

			require.NotEmpty(t, response.TxJSON)
			require.NotNil(t, response.Meta)
			require.Empty(t, response.TxBlob)
			require.Empty(t, response.MetaBlob)
			require.Equal(t, account.GetAddress().String(), response.TxJSON["Account"])
			require.Equal(t, "AccountSet", response.TxJSON["TransactionType"])
		})
	}
}

func TestIntegrationSimulate_Websocket(t *testing.T) {
	env := testintegration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationSimulate(t, client)
}

func TestIntegrationSimulate_RPCClient(t *testing.T) {
	env := testintegration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationSimulate(t, client)
}
