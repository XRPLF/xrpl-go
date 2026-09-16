package websocket

import (
	"testing"

	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/stretchr/testify/require"
)

func TestClient_GetLedgerEntry(t *testing.T) {
	const index = "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8"
	for _, tt := range []struct {
		name     string
		binary   bool
		response ledgerquery.EntryResponse
	}{
		{name: "JSON", response: ledgerquery.EntryResponse{Node: ledgerentry.FlatLedgerObject{"LedgerEntryType": "AccountRoot"}}},
		{name: "binary", binary: true, response: ledgerquery.EntryResponse{NodeBinary: "1100612200000000"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client, request := setupQueryTestClient(t, &tt.response)
			response, err := client.GetLedgerEntry(&ledgerquery.EntryRequest{Index: index, Binary: tt.binary})
			require.NoError(t, err)
			require.Equal(t, &tt.response, response)
			got := request()
			require.Equal(t, "ledger_entry", got["command"])
			require.Equal(t, index, got["index"])
			if tt.binary {
				require.Equal(t, true, got["binary"])
			} else {
				require.NotContains(t, got, "binary")
			}
		})
	}
}
