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

// NFTsRequest retrieves all NFTs currently owned by the specified account.
type NFTsRequest struct {
	common.BaseRequest
	Account     types.Address          `json:"account"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Marker      any                    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name for NFTsRequest.
func (*NFTsRequest) Method() string {
	return "account_nfts"
}

// APIVersion returns the Rippled API version for NFTsRequest.
func (*NFTsRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks the NFTsRequest parameters for validity.
// TODO implement v2
func (*NFTsRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTsResponse is the response returned by the account_nfts method,
// containing NFT records owned by the account.
type NFTsResponse struct {
	Account            types.Address      `json:"account"`
	AccountNFTs        []accounttypes.NFT `json:"account_nfts"`
	LedgerIndex        common.LedgerIndex `json:"ledger_index,omitempty"`
	LedgerHash         common.LedgerHash  `json:"ledger_hash,omitempty"`
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index,omitempty"`
	Validated          bool               `json:"validated"`
	Marker             any                `json:"marker,omitempty"`
	Limit              int                `json:"limit,omitempty"`
}
