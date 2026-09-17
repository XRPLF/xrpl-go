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

func TestGatewayBalancesRequest(t *testing.T) {
	s := account.GatewayBalancesRequest{
		Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		Strict:      true,
		HotWallet:   []string{"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu"},
		LedgerHash:  "ABC123",
		LedgerIndex: common.LedgerTitle("validated"),
	}

	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"strict": true,
	"hotwallet": [
		"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu"
	],
	"ledger_hash": "ABC123",
	"ledger_index": "validated"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func gatewayBalancesResponseFixture() (account.GatewayBalancesResponse, string) {
	s := account.GatewayBalancesResponse{
		Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		Obligations: map[string]string{
			"USD": "100",
			"EUR": "200",
		},
		Balances: map[string][]account.GatewayBalance{
			"ra7JkEzrgeKHdzKgo4EUUVBnxggY4z37kt": {{Currency: "USD", Value: "12345.9"}},
			"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu": {
				{
					Currency: "USD",
					Value:    "50",
				},
				{
					Currency: "EUR",
					Value:    "100",
				},
			},
		},
		Assets: map[string][]account.GatewayBalance{
			"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu": {
				{
					Currency: "USD",
					Value:    "5444166510000000e-26",
				},
			},
		},
		LedgerHash:         "ABC123",
		LedgerCurrentIndex: 54321,
		LedgerIndex:        12345,
	}
	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"obligations": {
		"EUR": "200",
		"USD": "100"
	},
	"balances": {
		"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu": [
			{
				"currency": "USD",
				"value": "50"
			},
			{
				"currency": "EUR",
				"value": "100"
			}
		],
		"ra7JkEzrgeKHdzKgo4EUUVBnxggY4z37kt": [
			{
				"currency": "USD",
				"value": "12345.9"
			}
		]
	},
	"assets": {
		"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu": [
			{
				"currency": "USD",
				"value": "5444166510000000e-26"
			}
		]
	},
	"ledger_hash": "ABC123",
	"ledger_current_index": 54321,
	"ledger_index": 12345
}`
	return s, j
}

func TestGatewayBalancesResponseSerialize(t *testing.T) {
	value, payload := gatewayBalancesResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestGatewayBalancesResponseJSONDecode(t *testing.T) {
	want, payload := gatewayBalancesResponseFixture()
	var got account.GatewayBalancesResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestGatewayBalancesResponseClientDecode(t *testing.T) {
	want, payload := gatewayBalancesResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got account.GatewayBalancesResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
