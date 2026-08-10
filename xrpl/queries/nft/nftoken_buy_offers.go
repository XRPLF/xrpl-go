// Package nft provides commands to query XRPL NFT-related methods.
package nft

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"

	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
)

// ############################################################################
// Request
// ############################################################################

// NFTokenBuyOffersRequest retrieves all buy offers for the specified NFT.
type NFTokenBuyOffersRequest struct {
	common.BaseRequest
	NFTokenID   types.NFTokenID        `json:"nft_id"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Marker      any                    `json:"marker,omitempty"`
}

// Method returns the XRPL JSON-RPC method name for NFTokenBuyOffersRequest.
func (*NFTokenBuyOffersRequest) Method() string {
	return "nft_buy_offers"
}

// APIVersion returns the XRPL API version for NFTokenBuyOffersRequest.
func (*NFTokenBuyOffersRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate ensures the NFTokenBuyOffersRequest is valid.
// TODO: Implement V2
func (*NFTokenBuyOffersRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTokenBuyOffersResponse is the expected response from the nft_buy_offers method.
type NFTokenBuyOffersResponse struct {
	NFTokenID types.NFTokenID         `json:"nft_id"`
	Offers    []nfttypes.NFTokenOffer `json:"offers"`
	Limit     int                     `json:"limit,omitempty"`
	Marker    any                     `json:"marker,omitempty"`
}
