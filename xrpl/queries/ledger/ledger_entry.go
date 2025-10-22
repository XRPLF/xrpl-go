package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// The `ledger_entry` method returns a single ledger object from the XRP Ledger
// in its raw format. Expects a response in the form of a EntryResponse.
type EntryRequest struct {
	common.BaseRequest
	MPTIssuance                     bool                    `json:"mpt_issuance,omitempty"`
	MPToken                         interface{}             `json:"mptoken,omitempty"`
	AMM                             types.EntryAssetPair    `json:"amm,omitempty"`
	IncludeDeleted                  bool                    `json:"include_deleted,omitempty"`
	Binary                          bool                    `json:"binary,omitempty"`
	Index                           string                  `json:"index,omitempty"`
	AccountRoot                     string                  `json:"account_root,omitempty"`
	Check                           string                  `json:"check,omitempty"`
	Credential                      interface{}             `json:"credential,omitempty"`
	DepositPreauth                  interface{}             `json:"deposit_preauth,omitempty"`
	Did                             string                  `json:"did,omitempty"`
	Directory                       interface{}             `json:"directory,omitempty"`
	Escrow                          interface{}             `json:"escrow,omitempty"`
	Offer                           interface{}             `json:"offer,omitempty"`
	PaymentChannel                  string                  `json:"payment_channel,omitempty"`
	RippleState                     types.EntryRippleState  `json:"ripple_state,omitempty"`
	Ticket                          interface{}             `json:"ticket,omitempty"`
	NFTPage                         string                  `json:"nft_page,omitempty"`
	BridgeAccount                   string                  `json:"bridge_account,omitempty"`
	Bridge                          types.EntryXChainBridge `json:"bridge,omitempty"`
	XChainOwnedClaimID              interface{}             `json:"xchain_owned_claim_id,omitempty"`
	XChainOwnedCreateAccountClaimID interface{}             `json:"xchain_owned_create_account_claim_id,omitempty"`
}

func (e *EntryRequest) Method() string {
	return "ledger_entry"
}

func (e *EntryRequest) APIVersion() int {
	return version.RippledAPIV2
}

func (e *EntryRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// The expected response from the ledger_entry method.
type EntryResponse struct {
	Index              string             `json:"index"`
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index"`
	Node               interface{}        `json:"node,omitempty"`
	NodeBinary         string             `json:"node_binary,omitempty"`
	Validated          bool               `json:"validated,omitempty"`
	DeletedLedgerIndex common.LedgerIndex `json:"deleted_ledger_index,omitempty"`
}
