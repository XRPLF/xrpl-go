package transaction

import (
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	rippletime "github.com/Peersyst/xrpl-go/xrpl/time"
)

const (
	// OracleSetMaxPriceDataSeriesItems is the maximum number of PriceData objects allowed in a PriceDataSeries array.
	OracleSetMaxPriceDataSeriesItems int = 10
	// OracleSetProviderMaxLength is the maximum decoded length in bytes for the Provider field.
	OracleSetProviderMaxLength int = 256
	// OracleSetURIMaxLength is the maximum decoded length in bytes for the URI field.
	OracleSetURIMaxLength int = 256
	// OracleSetAssetClassMaxLength is the maximum decoded length in bytes for the AssetClass field.
	OracleSetAssetClassMaxLength int = 16
)

// OracleSet creates a new Oracle ledger entry or updates the fields of an existing one using the Oracle ID.
//
// The oracle provider must complete these steps before submitting this transaction:
// 1. Create or own the XRPL account in the Owner field and have enough XRP to meet the reserve and transaction fee requirements.
// 2. Publish the XRPL account public key, so it can be used for verification by dApps.
// 3. Publish a registry of available price oracles with their unique OracleDocumentID.
//
// ```json
//
//	{
//	  "TransactionType": "OracleSet",
//	  "Account": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
//	  "OracleDocumentID": 34,
//	  "Provider": "70726F7669646572",
//	  "LastUpdateTime": 1724871860,
//	  "AssetClass": "63757272656E6379",
//	  "PriceDataSeries": [
//	    {
//	      "PriceData": {
//	        "BaseAsset": "XRP",
//	        "QuoteAsset": "USD",
//	        "AssetPrice": "00000000000002E4",
//	        "Scale": 3
//	      }
//	    }
//	  ]
//	}
//
// ```
type OracleSet struct {
	BaseTx
	// A unique identifier of the price oracle for the Account. It is 0 by default.
	OracleDocumentID uint32
	// The time the data was last updated, in seconds since the UNIX Epoch.
	// It must be at or after the Ripple epoch (2000-01-01T00:00:00Z).
	LastUpdateTime uint32
	// (Variable) Hex-encoded data that identifies an oracle provider, limited to 256 decoded bytes.
	// This field is required when creating a new Oracle ledger entry, but is optional for updates.
	Provider string `json:",omitempty"`
	// (Optional) Hex-encoded URI data that references price data off-chain, limited to 256 decoded bytes.
	URI string `json:",omitempty"`
	// (Variable) Hex-encoded asset classification data, limited to 16 decoded bytes.
	// This field is required when creating a new Oracle ledger entry, but is optional for updates.
	AssetClass string `json:",omitempty"`
	// An array of up to 10 PriceData objects, each representing the price information for a token pair. More than five PriceData objects require two owner reserves.
	PriceDataSeries []ledger.PriceDataWrapper
}

// TxType returns the TxType for OracleSet transactions.
func (tx *OracleSet) TxType() TxType {
	return OracleSetTx
}

// Flatten returns a map representation of the OracleSet transaction for JSON-RPC submission.
func (tx *OracleSet) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	flattened["OracleDocumentID"] = tx.OracleDocumentID

	if tx.Provider != "" {
		flattened["Provider"] = tx.Provider
	}
	if tx.URI != "" {
		flattened["URI"] = tx.URI
	}

	flattened["LastUpdateTime"] = tx.LastUpdateTime

	if tx.AssetClass != "" {
		flattened["AssetClass"] = tx.AssetClass
	}

	if len(tx.PriceDataSeries) > 0 {
		flattenedPriceDataSeries := make([]map[string]any, len(tx.PriceDataSeries))
		for i, priceDataWrapper := range tx.PriceDataSeries {
			flattenedPriceDataSeries[i] = priceDataWrapper.Flatten()
		}
		flattened["PriceDataSeries"] = flattenedPriceDataSeries
	}

	return flattened
}

// Validate checks OracleSet transaction fields and returns false with an error if invalid.
func (tx *OracleSet) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.Provider != "" {
		if !typecheck.IsHexBlob(tx.Provider) {
			return false, ErrOracleProviderInvalid
		}
		if decodedLength := len(tx.Provider) / 2; decodedLength > OracleSetProviderMaxLength {
			return false, ErrOracleProviderLength{
				Length: decodedLength,
				Limit:  OracleSetProviderMaxLength,
			}
		}
	}

	if tx.URI != "" && (!typecheck.IsHexBlob(tx.URI) || len(tx.URI)/2 > OracleSetURIMaxLength) {
		return false, ErrOracleURIInvalid
	}

	if tx.AssetClass != "" && (!typecheck.IsHexBlob(tx.AssetClass) || len(tx.AssetClass)/2 > OracleSetAssetClassMaxLength) {
		return false, ErrOracleAssetClassInvalid
	}

	if len(tx.PriceDataSeries) > OracleSetMaxPriceDataSeriesItems {
		return false, ErrOraclePriceDataSeriesItems{
			Length: len(tx.PriceDataSeries),
			Limit:  OracleSetMaxPriceDataSeriesItems,
		}
	}

	for _, priceDataWrapper := range tx.PriceDataSeries {
		if err := priceDataWrapper.PriceData.Validate(); err != nil {
			return false, err
		}
	}

	// The ledger-time window and update monotonicity require ledger state and remain server-side.
	if int64(tx.LastUpdateTime) < rippletime.RippleEpochDiff {
		return false, ErrOracleLastUpdateTimeInvalid
	}

	return true, nil
}
