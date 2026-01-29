package transaction

import (
	"errors"
	"strings"
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/stretchr/testify/assert"
)

func TestOracleSet_TxType(t *testing.T) {
	tx := &OracleSet{}
	assert.Equal(t, tx.TxType(), OracleSetTx)
}

func TestOracleSet_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *OracleSet
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &OracleSet{},
			expected: map[string]interface{}{
				"TransactionType":  OracleSetTx.String(),
				"OracleDocumentID": uint32(0),
				"LastUpdatedTime":  uint32(0),
			},
		},
		{
			name: "pass - complete",
			tx: &OracleSet{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				OracleDocumentID: 1,
				Provider:         "Chainlink",
				URI:              "https://example.com",
				LastUpdatedTime:  1715702400,
				AssetClass:       "currency",
				PriceDataSeries: []ledger.PriceDataWrapper{
					{
						PriceData: ledger.PriceData{
							BaseAsset:  "XRP",
							QuoteAsset: "USD",
							AssetPrice: 740,
							Scale:      3,
						},
					},
				},
			},
			expected: map[string]interface{}{
				"TransactionType":    OracleSetTx.String(),
				"Account":            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"OracleDocumentID":   uint32(1),
				"Provider":           "Chainlink",
				"URI":                "https://example.com",
				"LastUpdatedTime":    uint32(1715702400),
				"AssetClass":         "currency",
				"PriceDataSeries": []map[string]any{
					{
						"PriceData": map[string]any{
							"AssetPrice": "740",
							"BaseAsset":  "XRP",
							"QuoteAsset": "USD",
							"Scale":      uint8(3),
						},
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.tx.Flatten(), testcase.expected)
		})
	}
}

func TestOracleSet_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *OracleSet
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &OracleSet{
				BaseTx: BaseTx{
					TransactionType: OracleSetTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - provider length",
			tx: &OracleSet{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: OracleSetTx,
				},
				Provider: strings.Repeat("a", 257),
			},
			expected: ErrOracleProviderLength{
				Length: 257,
				Limit:  OracleSetProviderMaxLength,
			},
		},
		{
			name: "fail - price data series items",
			tx: &OracleSet{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: OracleSetTx,
				},
				PriceDataSeries: make([]ledger.PriceDataWrapper, 100),
			},
			expected: ErrOraclePriceDataSeriesItems{
				Length: 100,
				Limit:  OracleSetMaxPriceDataSeriesItems,
			},
		},
		{
			name: "fail - price data series item invalid",
			tx: &OracleSet{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: OracleSetTx,
				},
				PriceDataSeries: []ledger.PriceDataWrapper{
					{
						PriceData: ledger.PriceData{
							BaseAsset: "XRP",
						},
					},
				},
			},
			expected: ledger.ErrPriceDataQuoteAsset,
		},
		{
			name: "fail - price data series item scale",
			tx: &OracleSet{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: OracleSetTx,
				},
				PriceDataSeries: []ledger.PriceDataWrapper{
					{
						PriceData: ledger.PriceData{
							BaseAsset:  "XRP",
							QuoteAsset: "USD",
							Scale:      11,
						},
					},
				},
			},
			expected: ledger.ErrPriceDataScale{
				Value: 11,
				Limit: ledger.PriceDataScaleMax,
			},
		},
		{
			name: "fail - price data series item asset price and scale",
			tx: &OracleSet{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: OracleSetTx,
				},
				PriceDataSeries: []ledger.PriceDataWrapper{
					{
						PriceData: ledger.PriceData{
							BaseAsset:  "XRP",
							QuoteAsset: "USD",
							Scale:      10,
						},
					},
				},
			},
			expected: ledger.ErrPriceDataAssetPriceAndScale,
		},
		{
			name: "pass - complete",
			tx: &OracleSet{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: OracleSetTx,
				},
				OracleDocumentID: 1,
				Provider:         "Chainlink",
				URI:              "https://example.com",
				LastUpdatedTime:  1715702400,
				AssetClass:       "currency",
				PriceDataSeries: []ledger.PriceDataWrapper{
					{
						PriceData: ledger.PriceData{
							BaseAsset:  "XRP",
							QuoteAsset: "USD",
							AssetPrice: 740,
							Scale:      3,
						},
					},
				},
			},
			expected: nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ok, err := testcase.tx.Validate()
			assert.Equal(t, ok, testcase.expected == nil)
			assert.True(t, errors.Is(err, testcase.expected), "expected %v, got %v", testcase.expected, err)
		})
	}
}
