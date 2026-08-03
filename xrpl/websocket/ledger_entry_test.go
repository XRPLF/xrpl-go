package websocket

import (
	"testing"

	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/stretchr/testify/require"
)

func TestClient_GetLedgerEntry(t *testing.T) {
	tests := []struct {
		name     string
		message  map[string]any
		request  *ledgerquery.EntryRequest
		expected *ledgerquery.EntryResponse
	}{
		{
			name: "json response",
			message: map[string]any{
				"id": 1,
				"result": map[string]any{
					"index":        "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
					"ledger_hash":  "31850E8E48E76D1064651DF39DF4E9542E8C90A9A9B629F4DE339EB3FA74F726",
					"ledger_index": uint32(61966146),
					"node": map[string]any{
						"Account":         "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
						"Balance":         "424021949",
						"LedgerEntryType": "AccountRoot",
					},
					"deleted_ledger_index": uint32(61966150),
					"validated":            true,
				},
			},
			request: &ledgerquery.EntryRequest{
				AccountRoot:    "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
				IncludeDeleted: true,
			},
			expected: &ledgerquery.EntryResponse{
				Index:       "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerHash:  "31850E8E48E76D1064651DF39DF4E9542E8C90A9A9B629F4DE339EB3FA74F726",
				LedgerIndex: common.LedgerIndex(61966146),
				Node: ledgerentry.FlatLedgerObject{
					"Account":         "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"Balance":         "424021949",
					"LedgerEntryType": "AccountRoot",
				},
				DeletedLedgerIndex: common.LedgerIndex(61966150),
				Validated:          true,
			},
		},
		{
			name: "binary response",
			message: map[string]any{
				"id": 1,
				"result": map[string]any{
					"index":        "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
					"ledger_index": uint32(61966146),
					"node_binary":  "1100612200000000",
					"validated":    true,
				},
			},
			request: &ledgerquery.EntryRequest{
				Index:  "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				Binary: true,
			},
			expected: &ledgerquery.EntryResponse{
				Index:       "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerIndex: common.LedgerIndex(61966146),
				NodeBinary:  "1100612200000000",
				Validated:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestClient(t, []map[string]any{tt.message})
			defer cleanup()

			response, err := client.GetLedgerEntry(tt.request)
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}
