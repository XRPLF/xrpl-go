// Package oracle contains oracle-related queries for XRPL.
package oracle

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/oracle/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// GetAggregatePriceRequest is a request to retrieve the aggregate price of specified Oracle objects,
// returning mean, median, and trimmed mean statistics.
type GetAggregatePriceRequest struct {
	common.BaseRequest
	// The currency code of the asset to be priced.
	BaseAsset string `json:"base_asset"`
	// The currency code of the asset to quote the price of the base asset.
	QuoteAsset string `json:"quote_asset"`
	// The oracles identifiers
	Oracles []types.Oracle `json:"oracles"`
	// The percentage of outliers to trim. Valid trim range is 1-25. If included, the API
	// returns statistics for the trimmed mean.
	Trim uint32 `json:"trim,omitempty"`
	// Defines a time range in seconds for filtering out older price data. Default value is 0,
	// which doesn't filter any data.
	TrimThreshold uint32 `json:"trim_threshold,omitempty"`
}

// Method returns the JSON-RPC method name for GetAggregatePriceRequest.
func (r *GetAggregatePriceRequest) Method() string {
	return "get_aggregate_price"
}

// APIVersion returns the Rippled JSON-RPC API version for GetAggregatePriceRequest.
func (r *GetAggregatePriceRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate ensures the GetAggregatePriceRequest contains valid fields.
func (r *GetAggregatePriceRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// GetAggregatePriceResponse is the response returned by the get_aggregate_price method,
// containing aggregate price statistics from the queried oracles.
type GetAggregatePriceResponse struct {
	// The statistics from the collected oracle prices.
	EntireSet types.Set `json:"entire_set"`
	// The trimmed statistics from the collected oracle prices. Only appears if the trim field was specified in the request.
	TrimmedSet types.Set `json:"trimmed_set,omitzero"`
	// The median of the collected oracle prices.
	Median string `json:"median"`
	// The most recent timestamp out of all LastUpdateTime values.
	Time uint `json:"time"`
	// The ledger index of the ledger version that was used to generate this
	// response.
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index"`
	// If included and set to true, the information in this response comes from
	// a validated ledger version. Otherwise, the information is subject to
	// change.
	Validated bool `json:"validated"`
}
