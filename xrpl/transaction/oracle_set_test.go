package transaction

import (
	"encoding/json"
	"strings"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	codectypes "github.com/Peersyst/xrpl-go/binary-codec/types"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOracleSet_TxType(t *testing.T) {
	tx := &OracleSet{}
	assert.Equal(t, OracleSetTx, tx.TxType())
}

func TestOracleSet_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *OracleSet
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &OracleSet{},
			expected: map[string]any{
				"TransactionType":  OracleSetTx.String(),
				"OracleDocumentID": uint32(0),
				"LastUpdateTime":   uint32(0),
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
				LastUpdateTime:   1715702400,
				AssetClass:       "currency",
				PriceDataSeries: []ledger.PriceDataWrapper{
					{
						PriceData: ledger.PriceData{
							BaseAsset:  "XRP",
							QuoteAsset: "USD",
							AssetPrice: ledger.AssetPrice(740),
							Scale:      3,
						},
					},
				},
			},
			expected: map[string]any{
				"TransactionType":    OracleSetTx.String(),
				"Account":            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"OracleDocumentID":   uint32(1),
				"Provider":           "Chainlink",
				"URI":                "https://example.com",
				"LastUpdateTime":     uint32(1715702400),
				"AssetClass":         "currency",
				"PriceDataSeries": []map[string]any{
					{
						"PriceData": map[string]any{
							"AssetPrice": "00000000000002E4",
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
			assert.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestOracleSet_AssetPriceEncoding(t *testing.T) {
	// AssetPrice accepts decimal JSON input and uses the canonical UInt64
	// hexadecimal form in flattened and decoded transaction data.
	const expectedEncoding = "12003324000000012F66CF74B420330000002268400000000000000C701C0863757272656E6379701D0870726F7669646572811494AE4477CF81EA0D6FC33DD82EC2D499206A8A89F018E020301700000000000002E4041003011A0000000000000000000000000000000000000000021A0000000000000000000000005553440000000000E1F1"

	var priceData ledger.PriceData
	require.NoError(t, json.Unmarshal([]byte(`{
		"BaseAsset": "XRP",
		"QuoteAsset": "USD",
		"AssetPrice": 740,
		"Scale": 3
	}`), &priceData))

	tx := &OracleSet{
		BaseTx: BaseTx{
			Account:  "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
			Fee:      12,
			Sequence: 1,
		},
		OracleDocumentID: 34,
		Provider:         "70726F7669646572",
		LastUpdateTime:   1724871860,
		AssetClass:       "63757272656E6379",
		PriceDataSeries:  []ledger.PriceDataWrapper{{PriceData: priceData}},
	}

	flattened := tx.Flatten()
	priceDataSeries := flattened["PriceDataSeries"].([]map[string]any)
	flattenedPriceData := priceDataSeries[0]["PriceData"].(map[string]any)
	require.Equal(t, "00000000000002E4", flattenedPriceData["AssetPrice"])

	encoded, err := binarycodec.Encode(flattened)
	require.NoError(t, err)
	require.Equal(t, expectedEncoding, encoded)

	decoded, err := binarycodec.Decode(encoded)
	require.NoError(t, err)
	decodedPriceDataSeries := decoded["PriceDataSeries"].([]any)
	decodedPriceDataWrapper := decodedPriceDataSeries[0].(map[string]any)
	decodedPriceData := decodedPriceDataWrapper["PriceData"].(map[string]any)
	require.Equal(t, "00000000000002E4", decodedPriceData["AssetPrice"])

	t.Run("previous numeric flattened value is rejected", func(t *testing.T) {
		previous := tx.Flatten()
		previousSeries := previous["PriceDataSeries"].([]map[string]any)
		previousPriceData := previousSeries[0]["PriceData"].(map[string]any)
		previousPriceData["AssetPrice"] = uint64(740)

		_, err := binarycodec.Encode(previous)
		require.ErrorIs(t, err, codectypes.ErrInvalidUInt64String)
	})
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
				LastUpdateTime:   1715702400,
				AssetClass:       "currency",
				PriceDataSeries: []ledger.PriceDataWrapper{
					{
						PriceData: ledger.PriceData{
							BaseAsset:  "XRP",
							QuoteAsset: "USD",
							AssetPrice: ledger.AssetPrice(740),
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
			assert.ErrorIs(t, err, testcase.expected, "expected %v, got %v", testcase.expected, err)
		})
	}
}
