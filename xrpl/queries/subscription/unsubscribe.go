package subscribe

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// UnsubscribeOrderBook represents an order book subscription filter to stop receiving updates.
type UnsubscribeOrderBook struct {
	TakerGets types.IssuedCurrencyAmount `json:"taker_gets,omitzero"`
	TakerPays types.IssuedCurrencyAmount `json:"taker_pays,omitzero"`
	Both      bool                       `json:"both,omitempty"`
}

// ############################################################################
// Request
// ############################################################################

// UnsubscribeRequest tells the server to stop sending messages for specific subscriptions.
type UnsubscribeRequest struct {
	common.BaseRequest
	Streams          []string               `json:"streams,omitempty"`
	Accounts         []types.Address        `json:"accounts,omitempty"`
	AccountsProposed []types.Address        `json:"accounts_proposed,omitempty"`
	Books            []UnsubscribeOrderBook `json:"books,omitempty"`
}

// Method returns the XRPL JSON-RPC method name for UnsubscribeRequest.
func (*UnsubscribeRequest) Method() string {
	return "unsubscribe"
}

// Validate ensures the UnsubscribeRequest is valid.
func (*UnsubscribeRequest) Validate() error {
	return nil
}

// APIVersion returns the XRPL API version for UnsubscribeRequest.
func (*UnsubscribeRequest) APIVersion() int {
	return version.RippledAPIV2
}

// ############################################################################
// Response
// ############################################################################

// UnsubscribeResponse is the expected response from the unsubscribe method.
type UnsubscribeResponse struct{}
