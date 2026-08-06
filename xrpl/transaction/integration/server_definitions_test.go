package integration

import (
	"testing"

	serverquery "github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	testintegration "github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationServerDefinitions(t *testing.T, client testintegration.Client) {
	t.Helper()

	if connectable, ok := client.(testintegration.Connectable); ok {
		require.NoError(t, connectable.Connect())
		defer connectable.Disconnect()
	}

	response, err := client.GetServerDefinitions(&serverquery.DefinitionsRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, response.Hash)
	require.NotEmpty(t, response.Fields)
	require.Contains(t, response.Types, "UInt16")
	require.Contains(t, response.LedgerEntryTypes, "AccountRoot")
	require.Contains(t, response.TransactionTypes, "Payment")
	require.Contains(t, response.TransactionResults, "tesSUCCESS")

	unchangedResponse, err := client.GetServerDefinitions(&serverquery.DefinitionsRequest{Hash: response.Hash})
	require.NoError(t, err)
	require.Equal(t, response.Hash, unchangedResponse.Hash)
	require.Empty(t, unchangedResponse.Fields)
	require.Empty(t, unchangedResponse.Types)
	require.Empty(t, unchangedResponse.LedgerEntryTypes)
	require.Empty(t, unchangedResponse.TransactionTypes)
	require.Empty(t, unchangedResponse.TransactionResults)
}

func TestIntegrationServerDefinitions_Websocket(t *testing.T) {
	env := testintegration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationServerDefinitions(t, client)
}

func TestIntegrationServerDefinitions_RPCClient(t *testing.T) {
	env := testintegration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationServerDefinitions(t, client)
}
