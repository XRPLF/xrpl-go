package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// TxRequest is the request type for the tx command.
// It retrieves information on a single transaction by its identifying hash.
type TxRequest struct {
	common.BaseRequest
	Transaction string             `json:"transaction"`
	Binary      bool               `json:"binary,omitempty"`
	MinLedger   common.LedgerIndex `json:"min_ledger,omitempty"`
	MaxLedger   common.LedgerIndex `json:"max_ledger,omitempty"`
}

// Method returns the JSON-RPC method name for the TxRequest.
func (*TxRequest) Method() string {
	return "tx"
}

// APIVersion returns the API version for the TxRequest.
func (*TxRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate verifies the TxRequest parameters.
// TODO: implement V2 validation logic.
func (*TxRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// TxResponse is the response type returned by the tx command.
// It includes transaction details, metadata, and validation status.
type TxResponse struct {
	Date        uint                          `json:"date"`
	CtID        string                        `json:"ct_id,omitempty"`
	Hash        types.Hash256                 `json:"hash"`
	LedgerIndex common.LedgerIndex            `json:"ledger_index"`
	Meta        transaction.TxMetadataBuilder `json:"meta"`
	Validated   bool                          `json:"validated"`
	Tx          transaction.FlatTransaction   `json:",omitempty"`
}
