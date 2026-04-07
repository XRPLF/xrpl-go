package ledger

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// PriceDataScaleMax is the maximum scale for a price data.
	PriceDataScaleMax uint8 = 10
)

// PriceDataWrapper represents a wrapper for the PriceData struct.
type PriceDataWrapper struct {
	PriceData PriceData
}

// A PriceData object represents the price information for a token pair.
type PriceData struct {
	// The primary asset in a trading pair. Any valid identifier, such as a stock symbol,
	// bond CUSIP, or currency code is allowed.
	BaseAsset string
	// The quote asset in a trading pair. The quote asset denotes the
	// price of one unit of the base asset.
	QuoteAsset string
	// The asset price after applying the Scale precision level. It's not included if
	// the last update transaction didn't include the BaseAsset/QuoteAsset pair.
	AssetPrice uint64 `json:",omitempty"`
	// The scaling factor to apply to an asset price. For example, if Scale is 6 and original price is 0.155,
	// then the scaled price is 155000. Valid scale ranges are 0-10.
	// It's not included if the last update transaction didn't include the BaseAsset/QuoteAsset pair.
	//
	// By default, the scale is 0.
	Scale uint8 `json:",omitempty"`
}

// Validate validates the price data.
func (priceData *PriceData) Validate() error {
	if len(priceData.BaseAsset) == 0 {
		return ErrPriceDataBaseAsset
	}

	if len(priceData.QuoteAsset) == 0 {
		return ErrPriceDataQuoteAsset
	}

	if priceData.Scale > PriceDataScaleMax {
		return ErrPriceDataScale{
			Value: priceData.Scale,
			Limit: PriceDataScaleMax,
		}
	}

	if (priceData.AssetPrice == 0) != (priceData.Scale == 0) {
		return ErrPriceDataAssetPriceAndScale
	}

	return nil
}

// Flatten returns a map containing the PriceData if it is set, or nil otherwise.
func (mw *PriceDataWrapper) Flatten() map[string]any {
	if mw.PriceData != (PriceData{}) {
		flattened := make(map[string]any)
		flattened["PriceData"] = mw.PriceData.Flatten()
		return flattened
	}
	return nil
}

// Flatten flattens the price data.
func (priceData *PriceData) Flatten() map[string]any {
	mapKeys := 2

	if priceData.Scale != 0 && priceData.AssetPrice != 0 {
		mapKeys = 4
	}

	flattened := make(map[string]any, mapKeys)

	if priceData.AssetPrice != 0 {
		// AssetPrice must be a hex string for the binary codec UInt64 type
		flattened["AssetPrice"] = fmt.Sprintf("%016X", priceData.AssetPrice)
	}
	if priceData.BaseAsset != "" {
		flattened["BaseAsset"] = priceData.BaseAsset
	}
	if priceData.QuoteAsset != "" {
		flattened["QuoteAsset"] = priceData.QuoteAsset
	}

	flattened["Scale"] = priceData.Scale

	return flattened
}

// Oracle ledger entry holds data associated with a single price oracle object.
// Requires PriceOracle amendment.
// Example:
// ```json
//
//	{
//	  "LedgerEntryType": "Oracle",
//	  "Owner": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
//	  "Provider": "70726F7669646572",
//	  "AssetClass": "63757272656E6379",
//	  "PriceDataSeries": [
//	    {
//	      "PriceData": {
//	        "BaseAsset": "XRP",
//	        "QuoteAsset": "USD",
//	        "AssetPrice": 740,
//	        "Scale": 3,
//	      }
//	    },
//	  ],
//	  "LastUpdateTime": 1724871860,
//	  "PreviousTxnID": "C53ECF838647FA5A4C780377025FEC7999AB4182590510CA461444B207AB74A9",
//	  "PreviousTxnLgrSeq": 3675418
//	}
//
// ```
type Oracle struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The XRPL account with update and delete privileges for the oracle.
	// It's recommended to set up multi-signing on this account.
	Owner types.Address
	// An arbitrary value that identifies an oracle provider, such as Chainlink, Band, or DIA.
	// This field is a string, up to 256 ASCII hex encoded characters (0x20-0x7E).
	Provider string
	// An array of up to 10 PriceData objects, each representing the price information for a token pair.
	// More than five PriceData objects require two owner reserves.
	PriceDataSeries []PriceDataWrapper
	// The time the data was last updated, represented in Unix time.
	LastUpdateTime uint32
	// An optional Universal Resource Identifier to reference price data off-chain.
	// This field is limited to 256 bytes.
	URI string `json:",omitempty"`
	// Describes the type of asset, such as "currency", "commodity", or "index". This field is a string,
	// up to 16 ASCII hex encoded characters (0x20-0x7E).
	AssetClass string
	// A hint indicating which page of the oracle owner's owner directory links to this entry,
	// in case the directory consists of multiple pages.
	OwnerNode uint64
	// The hash of the previous transaction that modified this entry.
	PreviousTxnID string
	// The ledger index that this object was most recently modified or created in.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the type of the ledger entry.
func (*Oracle) EntryType() EntryType {
	return OracleEntry
}
