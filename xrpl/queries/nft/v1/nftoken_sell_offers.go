package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// NFTokenSellOffersRequest represents a request to retrieve all sell offers for a specified NFToken.
type NFTokenSellOffersRequest struct {
	common.BaseRequest
	NFTokenID   types.NFTokenID        `json:"nft_id"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Marker      any                    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name for the NFTokenSellOffersRequest.
func (*NFTokenSellOffersRequest) Method() string {
	return "nft_sell_offers"
}

// APIVersion returns the supported API version for the NFTokenSellOffersRequest.
func (*NFTokenSellOffersRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks that the NFTokenSellOffersRequest is correctly formed.
// TODO implement V2
func (*NFTokenSellOffersRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTokenSellOffersResponse contains all sell offers returned for a specified NFToken.
type NFTokenSellOffersResponse struct {
	NFTokenID types.NFTokenID         `json:"nft_id"`
	Offers    []nfttypes.NFTokenOffer `json:"offers"`
	Limit     int                     `json:"limit,omitempty"`
	Marker    any                     `json:"marker,omitempty"`
}
