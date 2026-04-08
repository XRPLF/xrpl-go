// Package amm contains amm-related queries for XRPL.
package amm

import (
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// InfoRequest retrieves information about an AMM instance.
type InfoRequest struct {
	common.BaseRequest
	// The definition for one of the two assets this AMM holds.
	Asset ledger.Asset `json:"asset"`
	// The definition for the other asset this AMM holds.
	Asset2 ledger.Asset `json:"asset2"`
	// (Optional) The AMM Account to look up.
	AMMAccount types.Address `json:"amm_account,omitempty"`
}

// Method returns the JSON-RPC method name for InfoRequest.
func (*InfoRequest) Method() string {
	return "amm_info"
}

// APIVersion returns the API version supported by InfoRequest.
func (*InfoRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate performs validation on InfoRequest.
func (*InfoRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// AuctionSlotInfo represents the auction slot details in an amm_info response.
type AuctionSlotInfo struct {
	// The current owner of this auction slot.
	Account types.Address `json:"account"`
	// A list of at most 4 additional accounts authorized to trade at the discounted fee.
	AuthAccounts []AuthAccountInfo `json:"auth_accounts,omitempty"`
	// The trading fee charged to the auction owner.
	DiscountedFee uint16 `json:"discounted_fee"`
	// The amount the auction owner paid to win this slot, in LP Tokens.
	Price types.IssuedCurrencyAmount `json:"price"`
	// The time when this slot expires. Returned as a formatted date string by the server.
	Expiration string `json:"expiration,omitempty"`
}

// AuthAccountInfo represents an account authorized to trade at the discounted fee.
type AuthAccountInfo struct {
	// The authorized account address.
	Account types.Address `json:"account"`
}

// VoteSlotInfo represents one vote slot in an amm_info response.
type VoteSlotInfo struct {
	// The account that cast the vote.
	Account types.Address `json:"account"`
	// The proposed trading fee, in units of 1/100,000.
	TradingFee uint16 `json:"trading_fee"`
	// The weight of the vote, in units of 1/100,000.
	VoteWeight uint32 `json:"vote_weight"`
}

// Info represents the AMM data returned by the amm_info method.
type Info struct {
	// The address of the special account that holds this AMM's assets.
	Account types.Address `json:"account"`
	// The first asset this AMM holds. For XRP this is a string (drops), for IOU it is a map (currency/issuer/value).
	// Use any to accommodate both mapstructure (map[string]any / string) and JSON decoding.
	Amount any `json:"amount"`
	// The second asset this AMM holds. Same polymorphic type as Amount.
	Amount2 any `json:"amount2"`
	// Details of the current auction slot owner.
	AuctionSlot *AuctionSlotInfo `json:"auction_slot,omitempty"`
	// The total outstanding balance of LP tokens from this AMM instance.
	LPToken types.IssuedCurrencyAmount `json:"lp_token"`
	// The percentage fee for trades, in units of 1/100,000.
	TradingFee uint16 `json:"trading_fee"`
	// A list of vote objects representing votes on the pool's trading fee.
	VoteSlots []VoteSlotInfo `json:"vote_slots,omitempty"`
}

// InfoResponse is the response from the amm_info method.
type InfoResponse struct {
	// The AMM data.
	AMM Info `json:"amm"`
	// The identifying hash of the ledger used to generate this response.
	LedgerHash common.LedgerHash `json:"ledger_hash,omitempty"`
	// The ledger index of the ledger version used to generate this response.
	LedgerIndex common.LedgerIndex `json:"ledger_index,omitempty"`
	// If true, the information comes from a validated ledger version.
	Validated bool `json:"validated,omitempty"`
}
