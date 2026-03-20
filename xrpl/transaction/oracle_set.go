package transaction

import (
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
)

const (
	// OracleSetMaxPriceDataSeriesItems is the maximum number of PriceData objects allowed in a PriceDataSeries array.
	OracleSetMaxPriceDataSeriesItems int = 10
	// OracleSetProviderMaxLength is the maximum length in bytes for the Provider field.
	OracleSetProviderMaxLength int = 256
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
//	        "AssetPrice": 740,
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
	// It is 0 by default.
	LastUpdatedTime uint32
	// (Variable) An arbitrary value that identifies an oracle provider, such as Chainlink, Band, or DIA. This field is a string, up to 256 ASCII hex encoded characters (0x20-0x7E).
	// This field is required when creating a new Oracle ledger entry, but is optional for updates.
	Provider string `json:",omitempty"`
	// (Optional) An optional Universal Resource Identifier to reference price data off-chain. This field is limited to 256 bytes.
	URI string `json:",omitempty"`
	// (Variable) Describes the type of asset, such as "currency", "commodity", or "index". This field is a string, up to 16 ASCII hex encoded characters (0x20-0x7E).
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
func (tx *OracleSet) Flatten() map[string]any {
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

	flattened["LastUpdatedTime"] = tx.LastUpdatedTime

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

	if len([]byte(tx.Provider)) > OracleSetProviderMaxLength {
		return false, ErrOracleProviderLength{
			Length: len([]byte(tx.Provider)),
			Limit:  OracleSetProviderMaxLength,
		}
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

	return true, nil
}
