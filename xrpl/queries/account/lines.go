package account

import (
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// LinesRequest returns information about an account's trust lines, including balances in all non-XRP currencies and assets for a specific ledger version.
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

// APIVersion returns the API version supported by LinesRequest.
func (*LinesRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate performs validation on LinesRequest.
// TODO: implement V2.
func (*LinesRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// LinesResponse represents the response for the account_lines method, containing account trust lines and pagination data.
type LinesResponse struct {
	Account            types.Address            `json:"account"`
	Lines              []accounttypes.TrustLine `json:"lines"`
	LedgerCurrentIndex common.LedgerIndex       `json:"ledger_current_index,omitempty"`
	LedgerIndex        common.LedgerIndex       `json:"ledger_index,omitempty"`
	LedgerHash         common.LedgerHash        `json:"ledger_hash,omitempty"`
	Marker             any                      `json:"marker,omitempty"`
	Limit              int                      `json:"limit,omitempty"`
}
