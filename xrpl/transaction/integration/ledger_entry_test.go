package integration

import (
	"testing"

	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	testintegration "github.com/Peersyst/xrpl-go/xrpl/testutil/integration"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

func testIntegrationLedgerEntryTicketSelector(t *testing.T, client testintegration.Client) {
	t.Helper()

	runner := testintegration.NewRunner(t, client, &testintegration.RunnerConfig{WalletCount: 1})
	require.NoError(t, runner.Setup())
	defer runner.Teardown()

	account := runner.GetWallet(0)
	ticketSequence := submitTicketCreate(t, runner, client, account)

	response, err := client.GetLedgerEntry(&ledgerquery.EntryRequest{
		Ticket: ledgerquery.TicketSelector{
			Object: &ledgerquery.TicketSelectorFields{
				Account:   account.GetAddress(),
				TicketSeq: ticketSequence,
			},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, response.Index)
	require.Equal(t, "Ticket", response.Node["LedgerEntryType"])
	require.Equal(t, account.GetAddress().String(), response.Node["Account"])

	returnedSequence, err := uint32FromValue(response.Node["TicketSequence"])
	require.NoError(t, err)
	require.Equal(t, ticketSequence, returnedSequence)

	indexResponse, err := client.GetLedgerEntry(&ledgerquery.EntryRequest{
		Ticket: ledgerquery.TicketSelector{Index: response.Index},
	})
	require.NoError(t, err)
	require.Equal(t, response.Index, indexResponse.Index)
}

func TestIntegrationLedgerEntryTicketSelector_Websocket(t *testing.T) {
	env := testintegration.GetWebsocketEnv(t)
	client := websocket.NewClient(websocket.NewClientConfig().WithHost(env.Host).WithFaucetProvider(env.FaucetProvider))
	testIntegrationLedgerEntryTicketSelector(t, client)
}

func TestIntegrationLedgerEntryTicketSelector_RPCClient(t *testing.T) {
	env := testintegration.GetRPCEnv(t)
	clientCfg, err := rpc.NewClientConfig(env.Host, rpc.WithFaucetProvider(env.FaucetProvider))
	require.NoError(t, err)
	client := rpc.NewClient(clientCfg)
	testIntegrationLedgerEntryTicketSelector(t, client)
}
