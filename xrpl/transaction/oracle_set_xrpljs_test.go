package transaction

import (
	"encoding/json"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	codectypes "github.com/Peersyst/xrpl-go/binary-codec/types"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/stretchr/testify/require"
)

func TestOracleSet_AssetPriceEncodingMatchesXRPLJS(t *testing.T) {
	// Generated with xrpl.js v5.0.0 from the same transaction. xrpl.js accepts
	// AssetPrice as the decimal number 740, but its decoder returns the canonical
	// UInt64 JSON form "00000000000002E4".
	const xrplJSBlob = "12003324000000012F66CF74B420330000002268400000000000000C701C0863757272656E6379701D0870726F7669646572811494AE4477CF81EA0D6FC33DD82EC2D499206A8A89F018E020301700000000000002E4041003011A0000000000000000000000000000000000000000021A0000000000000000000000005553440000000000E1F1"

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
	require.Equal(t, xrplJSBlob, encoded)

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
