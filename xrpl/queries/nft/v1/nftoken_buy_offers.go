// Package v1 provides version 1 types and methods for NFT buy offers queries.
package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"

	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
)

// ############################################################################
// Request
// ############################################################################

// NFTokenBuyOffersRequest represents a request to retrieve all buy offers for a specified NFToken.
type NFTokenBuyOffersRequest struct {
	common.BaseRequest
	NFTokenID   types.NFTokenID        `json:"nft_id"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Marker      any                    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name for the NFTokenBuyOffersRequest.
func (*NFTokenBuyOffersRequest) Method() string {
	return "nft_buy_offers"
}

// APIVersion returns the supported API version for the NFTokenBuyOffersRequest.
func (*NFTokenBuyOffersRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks that the NFTokenBuyOffersRequest is correctly formed.
// TODO implement V2
func (*NFTokenBuyOffersRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTokenBuyOffersResponse contains the buy offers returned for a specified NFToken.
type NFTokenBuyOffersResponse struct {
	NFTokenID types.NFTokenID         `json:"nft_id"`
	Offers    []nfttypes.NFTokenOffer `json:"offers"`
	Limit     int                     `json:"limit,omitempty"`
	Marker    any                     `json:"marker,omitempty"`
}
