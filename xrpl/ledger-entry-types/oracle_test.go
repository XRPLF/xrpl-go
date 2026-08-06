package ledger

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOracle_EntryType(t *testing.T) {
	oracle := &Oracle{}
	assert.Equal(t, OracleEntry, oracle.EntryType())
}

func TestPriceData_Flatten(t *testing.T) {
	testcases := []struct {
		name      string
		priceData *PriceData
		expected  map[string]any
	}{
		{
			name:      "pass - empty",
			priceData: &PriceData{},
			expected:  map[string]any{},
		},
		{
			name: "pass - absent price omits scale",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
			},
			expected: map[string]any{
				"BaseAsset":  "XRP",
				"QuoteAsset": "USD",
			},
		},
		{
			name: "pass - explicit zero price",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				AssetPrice: AssetPrice(0),
			},
			expected: map[string]any{
				"BaseAsset":  "XRP",
				"QuoteAsset": "USD",
				"AssetPrice": "0000000000000000",
				"Scale":      uint8(0),
			},
		},
		{
			name: "pass - complete",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				AssetPrice: AssetPrice(740),
				Scale:      3,
			},
			expected: map[string]any{
				"BaseAsset":  "XRP",
				"QuoteAsset": "USD",
				"AssetPrice": "00000000000002E4",
				"Scale":      uint8(3),
			},
		},
		{
			name: "pass - complete with currency more than 3 characters",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "ACGBD",
				AssetPrice: AssetPrice(740),
				Scale:      3,
			},
			expected: map[string]any{
				"BaseAsset":  "XRP",
				"QuoteAsset": "ACGBD",
				"AssetPrice": "00000000000002E4",
				"Scale":      uint8(3),
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.expected, testcase.priceData.Flatten())
		})
	}
}

func TestPriceDataWrapper_Flatten(t *testing.T) {
	testcases := []struct {
		name      string
		priceData *PriceDataWrapper
		expected  map[string]any
	}{
		{
			name:      "pass - empty",
			priceData: &PriceDataWrapper{},
			expected:  nil,
		},
		{
			name: "pass - complete",
			priceData: &PriceDataWrapper{
				PriceData: PriceData{
					BaseAsset:  "XRP",
					QuoteAsset: "USD",
					AssetPrice: AssetPrice(740),
					Scale:      3,
				},
			},
			expected: map[string]any{
				"PriceData": map[string]any{
					"BaseAsset":  "XRP",
					"QuoteAsset": "USD",
					"AssetPrice": "00000000000002E4",
					"Scale":      uint8(3),
				},
			},
		},
		{
			name: "pass - complete with currency more than 3 characters",
			priceData: &PriceDataWrapper{
				PriceData: PriceData{
					BaseAsset:  "XRP",
					QuoteAsset: "ACGBD",
					AssetPrice: AssetPrice(740),
					Scale:      3,
				},
			},
			expected: map[string]any{
				"PriceData": map[string]any{
					"BaseAsset":  "XRP",
					"QuoteAsset": "ACGBD",
					"AssetPrice": "00000000000002E4",
					"Scale":      uint8(3),
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.expected, testcase.priceData.Flatten())
		})
	}
}

func TestOracle_JSONRoundTrip(t *testing.T) {
	// Shaped like a real ledger entry: AssetPrice and OwnerNode are hex strings.
	raw := `{
	"index": "1B7B2F5D1D2C4E5A1B7B2F5D1D2C4E5A1B7B2F5D1D2C4E5A1B7B2F5D1D2C4E5A",
	"LedgerEntryType": "Oracle",
	"Flags": 0,
	"Owner": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
	"Provider": "70726F7669646572",
	"AssetClass": "63757272656E6379",
	"PriceDataSeries": [
		{"PriceData": {"BaseAsset": "XRP", "QuoteAsset": "USD", "AssetPrice": "2e4", "Scale": 3}},
		{"PriceData": {"BaseAsset": "XRP", "QuoteAsset": "EUR", "AssetPrice": "00000000000002E4", "Scale": 3}},
		{"PriceData": {"BaseAsset": "XRP", "QuoteAsset": "JPY", "AssetPrice": "0"}},
		{"PriceData": {"BaseAsset": "XRP", "QuoteAsset": "INR"}}
	],
	"LastUpdateTime": 1724871860,
	"OwnerNode": "1f",
	"PreviousTxnID": "C53ECF838647FA5A4C780377025FEC7999AB4182590510CA461444B207AB74A9",
	"PreviousTxnLgrSeq": 3675418
}`

	var oracle Oracle
	require.NoError(t, json.Unmarshal([]byte(raw), &oracle))
	require.Equal(t, OracleEntry, oracle.EntryType())
	require.Equal(t, EntryType("Oracle"), oracle.LedgerEntryType)
	require.Equal(t, "1f", oracle.OwnerNode)
	require.Len(t, oracle.PriceDataSeries, 4)
	require.Equal(t, AssetPrice(740), oracle.PriceDataSeries[0].PriceData.AssetPrice)
	require.Equal(t, AssetPrice(740), oracle.PriceDataSeries[1].PriceData.AssetPrice)
	require.Equal(t, AssetPrice(0), oracle.PriceDataSeries[2].PriceData.AssetPrice)
	require.Nil(t, oracle.PriceDataSeries[3].PriceData.AssetPrice)

	encoded, err := json.Marshal(&oracle)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"AssetPrice":"2e4"`)
	require.Contains(t, string(encoded), `"AssetPrice":"0"`)

	var decoded Oracle
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	require.Equal(t, oracle, decoded)
}

func TestPriceData_AssetPriceJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected *uint64
		wantErr  error
	}{
		{
			name:     "pass - minimal lowercase hex string",
			json:     `{"BaseAsset": "XRP", "QuoteAsset": "USD", "AssetPrice": "2e4", "Scale": 3}`,
			expected: AssetPrice(740),
		},
		{
			name:     "pass - padded uppercase hex string",
			json:     `{"BaseAsset": "XRP", "QuoteAsset": "USD", "AssetPrice": "00000000000002E4", "Scale": 3}`,
			expected: AssetPrice(740),
		},
		{
			name:     "pass - maximum value",
			json:     `{"BaseAsset": "XRP", "QuoteAsset": "USD", "AssetPrice": "ffffffffffffffff", "Scale": 3}`,
			expected: AssetPrice(0xffffffffffffffff),
		},
		{
			name:     "pass - base-10 JSON number",
			json:     `{"BaseAsset": "XRP", "QuoteAsset": "USD", "AssetPrice": 740, "Scale": 3}`,
			expected: AssetPrice(740),
		},
		{
			name:     "pass - explicit zero",
			json:     `{"BaseAsset": "XRP", "QuoteAsset": "USD", "AssetPrice": "0"}`,
			expected: AssetPrice(0),
		},
		{
			name: "pass - omitted",
			json: `{"BaseAsset": "XRP", "QuoteAsset": "INR"}`,
		},
		{
			name: "pass - null",
			json: `{"BaseAsset": "XRP", "QuoteAsset": "INR", "AssetPrice": null}`,
		},
		{
			name:    "fail - non-hex string",
			json:    `{"BaseAsset": "XRP", "QuoteAsset": "USD", "AssetPrice": "xyz", "Scale": 3}`,
			wantErr: ErrPriceDataAssetPrice,
		},
		{
			name:    "fail - overflowing hex string",
			json:    `{"BaseAsset": "XRP", "QuoteAsset": "USD", "AssetPrice": "10000000000000000", "Scale": 3}`,
			wantErr: ErrPriceDataAssetPrice,
		},
		{
			name:    "fail - negative number",
			json:    `{"BaseAsset": "XRP", "QuoteAsset": "USD", "AssetPrice": -5, "Scale": 3}`,
			wantErr: ErrPriceDataAssetPrice,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var priceData PriceData
			err := json.Unmarshal([]byte(test.json), &priceData)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.expected, priceData.AssetPrice)
		})
	}
}

func TestPriceData_Validate(t *testing.T) {
	testcases := []struct {
		name      string
		priceData *PriceData
		expected  error
	}{
		{
			name:      "fail - empty",
			priceData: &PriceData{},
			expected:  ErrPriceDataBaseAsset,
		},
		{
			name: "fail - empty quote asset",
			priceData: &PriceData{
				BaseAsset: "XRP",
			},
			expected: ErrPriceDataQuoteAsset,
		},
		{
			name: "fail - scale greater than max",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				Scale:      11,
			},
			expected: ErrPriceDataScale{
				Value: 11,
				Limit: PriceDataScaleMax,
			},
		},
		{
			name: "fail - scale without asset price",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				Scale:      3,
			},
			expected: ErrPriceDataAssetPriceAndScale,
		},
		{
			name: "pass - asset price with default scale",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				AssetPrice: AssetPrice(740),
			},
			expected: nil,
		},
		{
			name: "pass - explicit zero price",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				AssetPrice: AssetPrice(0),
			},
			expected: nil,
		},
		{
			name: "pass - complete",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				AssetPrice: AssetPrice(740),
				Scale:      3,
			},
			expected: nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			err := testcase.priceData.Validate()
			if testcase.expected == nil {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testcase.expected)
			}
		})
	}
}
