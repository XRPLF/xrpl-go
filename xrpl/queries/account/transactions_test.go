package account

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAccountTransactionsRequest(t *testing.T) {
	s := TransactionsRequest{
		Account:        "abc",
		LedgerIndexMin: 100,
		LedgerIndexMax: 120,
		LedgerHash:     "def",
		LedgerIndex:    common.LedgerIndex(10),
		Marker:         "123",
	}

	j := `{
	"account": "abc",
	"ledger_index_min": 100,
	"ledger_index_max": 120,
	"ledger_hash": "def",
	"ledger_index": 10,
	"marker": "123"
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestAccountTransactionsResponse(t *testing.T) {
	s := TransactionsResponse{
		Account:        "abc",
		LedgerIndexMin: 100,
		LedgerIndexMax: 120,
		Limit:          10,
		Marker:         "123",
		Transactions: []Transaction{
			{
				Hash:         "def",
				LedgerHash:   "ghi",
				LedgerIndex:  10,
				CloseTimeISO: "2021-01-01T00:00:00Z",
				Tx: map[string]any{
					"TransactionType": "Payment",
					"Account":         "abc",
					"Destination":     "def",
					"Amount":          "100",
				},
			},
		},
	}

	j := `{
	"account": "abc",
	"ledger_index_min": 100,
	"ledger_index_max": 120,
	"limit": 10,
	"marker": "123",
	"transactions": [
		{
			"close_time_iso": "2021-01-01T00:00:00Z",
			"hash": "def",
			"ledger_hash": "ghi",
			"ledger_index": 10,
			"meta": {
				"AffectedNodes": null,
				"TransactionIndex": 0,
				"TransactionResult": ""
			},
			"tx_json": {
				"Account": "abc",
				"Amount": "100",
				"Destination": "def",
				"TransactionType": "Payment"
			},
			"tx_blob": "",
			"validated": false
		}
	],
	"validated": false
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
