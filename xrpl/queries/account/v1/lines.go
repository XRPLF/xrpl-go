package v1

import (
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// LinesRequest returns information about an account's trust lines, including balances in non-XRP currencies and assets.
// All information is relative to a specific ledger version.
type LinesRequest struct {
	common.BaseRequest
	Account       types.Address          `json:"account"`
	LedgerHash    common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex   common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Peer          types.Address          `json:"peer,omitempty"`
	IgnoreDefault bool                   `json:"ignore_default,omitempty"`
	Limit         int                    `json:"limit,omitempty"`
	Marker        any                    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name for LinesRequest.
func (*LinesRequest) Method() string {
	return "account_lines"
}

// APIVersion returns the Rippled API version for LinesRequest.
func (*LinesRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks the LinesRequest parameters for validity.
func (*LinesRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// LinesResponse is the response returned by the account_lines method, containing trust line details.
type LinesResponse struct {
	Account            types.Address            `json:"account"`
	Lines              []accounttypes.TrustLine `json:"lines"`
	LedgerCurrentIndex common.LedgerIndex       `json:"ledger_current_index,omitempty"`
	LedgerIndex        common.LedgerIndex       `json:"ledger_index,omitempty"`
	LedgerHash         common.LedgerHash        `json:"ledger_hash,omitempty"`
	Marker             any                      `json:"marker,omitempty"`
	Limit              int                      `json:"limit,omitempty"`
}
