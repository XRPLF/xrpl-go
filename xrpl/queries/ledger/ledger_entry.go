package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// EntryRequest retrieves a specific ledger entry by its index.
type EntryRequest struct {
	common.BaseRequest
	Index       string                 `json:"index"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	Binary      bool                   `json:"binary,omitempty"`
}

// Method returns the JSON-RPC method name for EntryRequest.
func (*EntryRequest) Method() string {
	return "ledger_entry"
}

// APIVersion returns the Rippled API version for EntryRequest.
func (*EntryRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks the EntryRequest fields for validity.
func (*EntryRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// EntryResponse is the response returned by the ledger_entry method, containing a single ledger entry.
type EntryResponse struct {
	Index              string                  `json:"index"`
	LedgerIndex        common.LedgerIndex      `json:"ledger_index,omitempty"`
	LedgerCurrentIndex common.LedgerIndex      `json:"ledger_current_index,omitempty"`
	Node               ledger.FlatLedgerObject `json:"node"`
	Validated          bool                    `json:"validated"`
}
