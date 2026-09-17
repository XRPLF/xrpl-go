package account_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

func TestAccountTransactionsRequest(t *testing.T) {
	s := account.TransactionsRequest{
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

func accountTransactionsResponseFixture() (account.TransactionsResponse, string) {
	s := account.TransactionsResponse{
		Account:        "abc",
		LedgerIndexMin: 100,
		LedgerIndexMax: 120,
		Limit:          10,
		Validated:      true,
		Marker:         "123",
		Transactions: []account.Transaction{
			{
				Hash:         "def",
				LedgerHash:   "ghi",
				LedgerIndex:  10,
				CloseTimeISO: "2021-01-01T00:00:00Z",
				Meta:         transaction.TxObjMeta{TransactionIndex: 1, TransactionResult: "tesSUCCESS"},
				Validated:    true,
				Tx: map[string]any{
					"TransactionType": "Payment",
					"Account":         "abc",
					"Destination":     "def",
					"Amount":          "100",
					"Fee":             "12",
					"Flags":           json.Number("2147483648"),
					"Sequence":        json.Number("1"),
					"SigningPubKey":   "0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020",
					"TxnSignature":    "3045022100A7CCD1B5F67D76C7F2C0B6C199C9D2F3721A14A3C69F3CB134E9BF9D9DD6F8B002206A5F974C4F4D07B4F5A99391BF3E93D9B0A7346861E12E924B6451B6D8ED4F09",
					"hash":            "2E2DDBF5B8F29AEED7494CC0A863A93A4BD3C066BF880A577C5C466EE2C637DF",
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
				"TransactionIndex": 1,
				"TransactionResult": "tesSUCCESS"
			},
			"tx_json": {
				"Account": "abc",
				"Amount": "100",
				"Destination": "def",
				"Fee": "12",
				"Flags": 2147483648,
				"Sequence": 1,
				"SigningPubKey": "0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020",
				"TransactionType": "Payment",
				"TxnSignature": "3045022100A7CCD1B5F67D76C7F2C0B6C199C9D2F3721A14A3C69F3CB134E9BF9D9DD6F8B002206A5F974C4F4D07B4F5A99391BF3E93D9B0A7346861E12E924B6451B6D8ED4F09",
				"hash": "2E2DDBF5B8F29AEED7494CC0A863A93A4BD3C066BF880A577C5C466EE2C637DF"
			},
			"tx_blob": "",
			"validated": true
		}
	],
	"validated": true
}`
	return s, j
}

func TestAccountTransactionsResponseSerialize(t *testing.T) {
	value, payload := accountTransactionsResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestAccountTransactionsResponseJSONDecode(t *testing.T) {
	want, payload := accountTransactionsResponseFixture()
	var got account.TransactionsResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestAccountTransactionsResponseClientDecode(t *testing.T) {
	want, payload := accountTransactionsResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got account.TransactionsResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
