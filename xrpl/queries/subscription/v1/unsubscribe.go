package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// UnsubscribeOrderBook specifies an order book to unsubscribe from.
type UnsubscribeOrderBook struct {
	TakerGets types.IssuedCurrencyAmount `json:"taker_gets,omitempty"`
	TakerPays types.IssuedCurrencyAmount `json:"taker_pays,omitempty"`
	Both      bool                       `json:"both,omitempty"`
}

// ############################################################################
// Request
// ############################################################################

// UnsubscribeRequest stops the server from sending messages for specified streams, accounts, or order books.
type UnsubscribeRequest struct {
	Streams          []string               `json:"streams,omitempty"`
	Accounts         []types.Address        `json:"accounts,omitempty"`
	AccountsProposed []types.Address        `json:"accounts_proposed,omitempty"`
	Books            []UnsubscribeOrderBook `json:"books,omitempty"`
}

// Method returns the JSON-RPC method name for UnsubscribeRequest.
func (*UnsubscribeRequest) Method() string {
	return "unsubscribe"
}

// Validate performs validation on UnsubscribeRequest.
// TODO: implement V2.
func (*UnsubscribeRequest) Validate() error {
	return nil
}

// APIVersion returns the API version supported by UnsubscribeRequest.
func (*UnsubscribeRequest) APIVersion() int {
	return version.RippledAPIV1
}

// ############################################################################
// Response
// ############################################################################

// UnsubscribeResponse represents the response from the unsubscribe method.
type UnsubscribeResponse struct{}
