package account_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/require"
)

func TestAccountCurrenciesRequest(t *testing.T) {
	s := account.CurrenciesRequest{
		Account:     "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		Strict:      true,
		LedgerIndex: common.LedgerIndex(1234),
	}

	j := `{
	"account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"ledger_index": 1234,
	"strict": true
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func accountCurrenciesResponseFixture() (account.CurrenciesResponse, string) {
	s := account.CurrenciesResponse{
		LedgerHash:  "abc",
		LedgerIndex: 123,
		ReceiveCurrencies: []string{
			"USD",
			"JPY",
		},
		SendCurrencies: []string{
			"USD",
			"CAD",
		},
		Validated: true,
	}
	j := `{
	"ledger_hash": "abc",
	"ledger_index": 123,
	"receive_currencies": [
		"USD",
		"JPY"
	],
	"send_currencies": [
		"USD",
		"CAD"
	],
	"validated": true
}`
	return s, j
}

func TestAccountCurrenciesResponseSerialize(t *testing.T) {
	value, payload := accountCurrenciesResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestAccountCurrenciesResponseJSONDecode(t *testing.T) {
	want, payload := accountCurrenciesResponseFixture()
	var got account.CurrenciesResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestAccountCurrenciesResponseClientDecode(t *testing.T) {
	want, payload := accountCurrenciesResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got account.CurrenciesResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
