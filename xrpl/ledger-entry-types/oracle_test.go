package ledger

import (
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
			expected: map[string]any{
				"Scale": uint8(0),
			},
		},
		{
			name: "pass - complete",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				AssetPrice: 740,
				Scale:      3,
			},
			expected: map[string]any{
				"BaseAsset":  "XRP",
				"QuoteAsset": "USD",
				"AssetPrice": "740",
				"Scale":      uint8(3),
			},
		},
		{
			name: "pass - complete with currency more than 3 characters",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "ACGBD",
				AssetPrice: 740,
				Scale:      3,
			},
			expected: map[string]any{
				"BaseAsset":  "XRP",
				"QuoteAsset": "ACGBD",
				"AssetPrice": "740",
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
					AssetPrice: 740,
					Scale:      3,
				},
			},
			expected: map[string]any{
				"PriceData": map[string]any{
					"BaseAsset":  "XRP",
					"QuoteAsset": "USD",
					"AssetPrice": "740",
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
					AssetPrice: 740,
					Scale:      3,
				},
			},
			expected: map[string]any{
				"PriceData": map[string]any{
					"BaseAsset":  "XRP",
					"QuoteAsset": "ACGBD",
					"AssetPrice": "740",
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
			name: "fail - asset price and scale not set together",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				AssetPrice: 740,
			},
			expected: ErrPriceDataAssetPriceAndScale,
		},
		{
			name: "pass - complete",
			priceData: &PriceData{
				BaseAsset:  "XRP",
				QuoteAsset: "USD",
				AssetPrice: 740,
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
