// Package nft provides commands to query XRPL NFT-related methods.
package nft

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// NFTokenSellOffersRequest retrieves all sell offers for the specified NFT.
type NFTokenSellOffersRequest struct {
	common.BaseRequest
	NFTokenID   types.NFTokenID        `json:"nft_id"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Marker      any                    `json:"marker,omitempty"`
}

// Method returns the XRPL JSON-RPC method name for NFTokenSellOffersRequest.
func (*NFTokenSellOffersRequest) Method() string {
	return "nft_sell_offers"
}

// APIVersion returns the XRPL API version for NFTokenSellOffersRequest.
func (*NFTokenSellOffersRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate ensures the NFTokenSellOffersRequest is valid.
func (*NFTokenSellOffersRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTokenSellOffersResponse is the expected response from the nft_sell_offers method.
type NFTokenSellOffersResponse struct {
	NFTokenID types.NFTokenID         `json:"nft_id"`
	Offers    []nfttypes.NFTokenOffer `json:"offers"`
	Limit     int                     `json:"limit,omitempty"`
	Marker    any                     `json:"marker,omitempty"`
}
