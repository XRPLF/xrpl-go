package path

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// RipplePathFindRequest is a simplified version of the path_find request that provides a single usable payment path.
type RipplePathFindRequest struct {
	common.BaseRequest
	SourceAccount      types.Address                      `json:"source_account"`
	DestinationAccount types.Address                      `json:"destination_account"`
	DestinationAmount  types.CurrencyAmount               `json:"destination_amount"`
	SendMax            types.CurrencyAmount               `json:"send_max,omitempty"`
	SourceCurrencies   []pathtypes.RipplePathFindCurrency `json:"source_currencies,omitempty"`
	LedgerHash         common.LedgerHash                  `json:"ledger_hash,omitempty"`
	LedgerIndex        common.LedgerSpecifier             `json:"ledger_index,omitempty"`
	Domain             *string                            `json:"domain,omitempty"`
}

// Method returns the JSON-RPC method name for the RipplePathFindRequest.
func (*RipplePathFindRequest) Method() string {
	return "ripple_path_find"
}

// APIVersion returns the supported API version for the RipplePathFindRequest.
func (*RipplePathFindRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks that the RipplePathFindRequest is correctly formed.
// TODO implement V2
func (*RipplePathFindRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// RipplePathFindResponse contains the result returned for a RipplePathFindRequest, including alternative paths.
type RipplePathFindResponse struct {
	Alternatives          []pathtypes.RippleAlternative `json:"alternatives"`
	DestinationAccount    types.Address                 `json:"destination_account"`
	DestinationCurrencies []string                      `json:"destination_currencies"`
	FullReply             bool                          `json:"full_reply,omitempty"`
	LedgerCurrentIndex    int                           `json:"ledger_current_index,omitempty"`
	SourceAccount         types.Address                 `json:"source_account"`
	Validated             bool                          `json:"validated"`
}
