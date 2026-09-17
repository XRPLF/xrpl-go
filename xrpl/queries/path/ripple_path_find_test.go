package path

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"

	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestRipplePathFindRequest(t *testing.T) {
	s := RipplePathFindRequest{
		SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		DestinationAmount: types.IssuedCurrencyAmount{
			Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
			Currency: "USD",
			Value:    "0.001",
		},
		SourceCurrencies: []pathtypes.RipplePathFindCurrency{
			{
				Currency: "XRP",
			},
			{
				Currency: "USD",
			},
		},
	}

	j := `{
	"source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"destination_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"destination_amount": {
		"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
		"currency": "USD",
		"value": "0.001"
	},
	"source_currencies": [
		{
			"currency": "XRP"
		},
		{
			"currency": "USD"
		}
	]
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestRipplePathFindRequestWithDomain(t *testing.T) {
	domain := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	s := RipplePathFindRequest{
		SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		DestinationAmount: types.IssuedCurrencyAmount{
			Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
			Currency: "USD",
			Value:    "0.001",
		},
		SourceCurrencies: []pathtypes.RipplePathFindCurrency{
			{
				Currency: "XRP",
			},
			{
				Currency: "USD",
			},
		},
		Domain: &domain,
	}

	j := `{
	"source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"destination_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"destination_amount": {
		"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
		"currency": "USD",
		"value": "0.001"
	},
	"source_currencies": [
		{
			"currency": "XRP"
		},
		{
			"currency": "USD"
		}
	],
	"domain": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func ripplePathFindResponseFixture() (RipplePathFindResponse, string) {
	s := RipplePathFindResponse{
		Alternatives: []pathtypes.RippleAlternative{
			{
				PathsComputed: [][]transaction.PathStep{
					{
						{
							Currency: "USD",
							Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
						},
						{
							Account: "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
						},
					},
					{
						{
							Currency: "USD",
							Issuer:   "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1",
						},
						{
							Account: "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1",
						},
						{
							Account: "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
						},
					},
					{
						{
							Currency: "USD",
							Issuer:   "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1",
						},
						{
							Account: "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1",
						},
						{
							Account: "rLpq4LgabRfm1xEX5dpWfJovYBH6g7z99q",
						},
						{
							Account: "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
						},
					},
					{
						{
							Currency: "USD",
							Issuer:   "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1",
						},
						{
							Account: "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1",
						},
						{
							Account: "rPuBoajMjFoDjweJBrtZEBwUMkyruxpwwV",
						},
						{
							Account: "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
						},
					},
				},
				SourceAmount: "256987",
			},
		},
		DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		SourceAccount:      "rMAZ5ZnK73nyNUL4foAvaxdreczCkG3vA6",
		FullReply:          true,
		DestinationCurrencies: []string{
			"015841551A748AD2C1F76FF6ECB0CCCD00000000",
			"JOE",
			"DYM",
			"EUR",
			"CNY",
			"MXN",
			"BTC",
			"USD",
			"XRP",
		},
	}
	j := `{
	"alternatives": [
		{
			"paths_computed": [
				[
					{
						"currency": "USD",
						"issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
					},
					{
						"account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
					}
				],
				[
					{
						"currency": "USD",
						"issuer": "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1"
					},
					{
						"account": "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1"
					},
					{
						"account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
					}
				],
				[
					{
						"currency": "USD",
						"issuer": "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1"
					},
					{
						"account": "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1"
					},
					{
						"account": "rLpq4LgabRfm1xEX5dpWfJovYBH6g7z99q"
					},
					{
						"account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
					}
				],
				[
					{
						"currency": "USD",
						"issuer": "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1"
					},
					{
						"account": "rrpNnNLKrartuEqfJGpqyDwPj1AFPg9vn1"
					},
					{
						"account": "rPuBoajMjFoDjweJBrtZEBwUMkyruxpwwV"
					},
					{
						"account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
					}
				]
			],
			"source_amount": "256987"
		}
	],
	"destination_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"destination_currencies": [
		"015841551A748AD2C1F76FF6ECB0CCCD00000000",
		"JOE",
		"DYM",
		"EUR",
		"CNY",
		"MXN",
		"BTC",
		"USD",
		"XRP"
	],
	"full_reply": true,
	"source_account": "rMAZ5ZnK73nyNUL4foAvaxdreczCkG3vA6",
	"validated": false
}`
	return s, j
}

func TestRipplePathFindResponseSerialize(t *testing.T) {
	value, payload := ripplePathFindResponseFixture()
	value.Alternatives[0].SourceAmount = types.XRPCurrencyAmount(256987)
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestRipplePathFindResponseJSONDecode(t *testing.T) {
	want, payload := ripplePathFindResponseFixture()
	var got RipplePathFindResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestRipplePathFindResponseClientDecode(t *testing.T) {
	want, payload := ripplePathFindResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got RipplePathFindResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}

func ripplePathFindResponseIssuedSourceAmountFixture() (RipplePathFindResponse, string) {
	expected := RipplePathFindResponse{
		Alternatives: []pathtypes.RippleAlternative{
			{
				PathsComputed: [][]transaction.PathStep{
					{
						{
							Currency: "USD",
							Issuer:   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
							Account:  "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
						},
					},
				},
				SourceAmount: map[string]any{
					"currency": "USD",
					"issuer":   "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
					"value":    "100",
				},
			},
		},
		DestinationAccount:    "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		DestinationCurrencies: []string{"USD"},
		SourceAccount:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		FullReply:             true,
	}
	payload := `{
  "alternatives": [
    {
      "paths_computed": [
        [
          {
            "currency": "USD",
            "issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
            "account": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B"
          }
        ]
      ],
      "source_amount": {
        "currency": "USD",
        "issuer": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
        "value": "100"
      }
    }
  ],
  "destination_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
  "destination_currencies": [
    "USD"
  ],
  "source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
  "full_reply": true,
  "validated": false
}`
	return expected, payload
}

func TestRipplePathFindResponseIssuedSourceAmountSerialize(t *testing.T) {
	value, payload := ripplePathFindResponseIssuedSourceAmountFixture()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, payload, string(encoded))
}

func TestRipplePathFindResponseIssuedSourceAmountJSONDecode(t *testing.T) {
	want, payload := ripplePathFindResponseIssuedSourceAmountFixture()
	var got RipplePathFindResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestRipplePathFindResponseIssuedSourceAmountClientDecode(t *testing.T) {
	want, payload := ripplePathFindResponseIssuedSourceAmountFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got RipplePathFindResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
