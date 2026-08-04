package ledger

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// PriceDataScaleMax is the maximum scale for a price data.
	PriceDataScaleMax uint8 = 10
)

// AssetPrice returns a pointer to an asset price value.
func AssetPrice(value uint64) *uint64 {
	return &value
}

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
	// On the wire this is a base-16 string (sfAssetPrice carries no base-ten flag in
	// rippled). The custom JSON (un)marshallers below convert to and from *uint64.
	// A nil value means the field is absent. A non-nil zero value is an explicit zero price.
	// Use AssetPrice to create a non-nil value.
	AssetPrice *uint64
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

	if priceData.AssetPrice == nil && priceData.Scale != 0 {
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
	flattened := make(map[string]any, 4)

	if priceData.AssetPrice != nil {
		// AssetPrice must be a hex string for the binary codec UInt64 type
		flattened["AssetPrice"] = fmt.Sprintf("%016X", *priceData.AssetPrice)
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

// MarshalJSON serializes the price data with AssetPrice in its hexadecimal wire form.
func (priceData PriceData) MarshalJSON() ([]byte, error) {
	type priceDataWire struct {
		BaseAsset  string
		QuoteAsset string
		AssetPrice string `json:",omitempty"`
		Scale      uint8  `json:",omitempty"`
	}
	wire := priceDataWire{
		BaseAsset:  priceData.BaseAsset,
		QuoteAsset: priceData.QuoteAsset,
		Scale:      priceData.Scale,
	}
	if priceData.AssetPrice != nil {
		wire.AssetPrice = strconv.FormatUint(*priceData.AssetPrice, 16)
	}
	return json.Marshal(wire)
}

// UnmarshalJSON decodes the price data, accepting AssetPrice as the hexadecimal
// string rippled emits or as a plain base-10 JSON number.
func (priceData *PriceData) UnmarshalJSON(data []byte) error {
	type priceDataRaw struct {
		BaseAsset  string
		QuoteAsset string
		AssetPrice json.RawMessage
		Scale      uint8
	}
	var raw priceDataRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*priceData = PriceData{
		BaseAsset:  raw.BaseAsset,
		QuoteAsset: raw.QuoteAsset,
		Scale:      raw.Scale,
	}

	if len(raw.AssetPrice) == 0 || string(raw.AssetPrice) == "null" {
		return nil
	}

	token := string(raw.AssetPrice)
	base := 10
	if raw.AssetPrice[0] == '"' {
		if err := json.Unmarshal(raw.AssetPrice, &token); err != nil {
			return err
		}
		base = 16
	}
	value, err := strconv.ParseUint(token, base, 64)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrPriceDataAssetPrice, token)
	}
	priceData.AssetPrice = AssetPrice(value)
	return nil
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
//	        "AssetPrice": "2e4",
//	        "Scale": 3,
//	      }
//	    },
//	  ],
//	  "LastUpdateTime": 1724871860,
//	  "OwnerNode": "0",
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
	// The type of ledger entry.
	LedgerEntryType EntryType
	// Set of bit-flags for this ledger entry.
	Flags uint32
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
	// A hexadecimal hint indicating which page of the oracle owner's owner directory links to this entry,
	// in case the directory consists of multiple pages.
	OwnerNode string
	// The hash of the previous transaction that modified this entry.
	PreviousTxnID string
	// The ledger index that this object was most recently modified or created in.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the type of the ledger entry.
func (*Oracle) EntryType() EntryType {
	return OracleEntry
}
