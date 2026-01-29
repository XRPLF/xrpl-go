package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// An AMM ledger entry describes a single Automated Market Maker (AMM) instance.
// This is always paired with a special AccountRoot entry. https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/amm#amm
//
// Example:
//
//	{
//	    "Account" : "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
//	    "Asset" : {
//	      "currency" : "XRP"
//	    },
//	    "Asset2" : {
//	      "currency" : "TST",
//	      "issuer" : "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
//	    },
//	    "AuctionSlot" : {
//	      "Account" : "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	      "AuthAccounts" : [
//	          {
//	            "AuthAccount" : {
//	                "Account" : "rMKXGCbJ5d8LbrqthdG46q3f969MVK2Qeg"
//	            }
//	          },
//	          {
//	            "AuthAccount" : {
//	                "Account" : "rBepJuTLFJt3WmtLXYAxSjtBWAeQxVbncv"
//	            }
//	          }
//	      ],
//	      "DiscountedFee" : 60,
//	      "Expiration" : 721870180,
//	      "Price" : {
//	          "currency" : "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
//	          "issuer" : "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
//	          "value" : "0.8696263565463045"
//	      }
//	    },
//	    "Flags" : 0,
//	    "LPTokenBalance" : {
//	      "currency" : "039C99CD9AB0B70B32ECDA51EAAE471625608EA2",
//	      "issuer" : "rE54zDvgnghAoPopCgvtiqWNq3dU5y836S",
//	      "value" : "71150.53584131501"
//	    },
//	    "TradingFee" : 600,
//	    "VoteSlots" : [
//	      {
//	          "VoteEntry" : {
//	            "Account" : "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	            "TradingFee" : 600,
//	            "VoteWeight" : 100000
//	          }
//	      }
//	    ]
//	}
type AMM struct {
	// The unique ID for this ledger entry.
	// In JSON, this field is represented with different names depending on the context and API method.
	// (Note, even though this is specified as "optional" in the code, every ledger entry should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The type of ledger entry. Valid ledger entry types include AccountRoot, Offer, RippleState, and others.
	LedgerEntryType string `json:",omitempty"`
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// The address of the special account that holds this AMM's assets.
	Account types.Address
	// The definition for one of the two assets this AMM holds. In JSON, this is an object with currency and issuer fields.
	Asset Asset
	// The definition for the other asset this AMM holds. In JSON, this is an object with currency and issuer fields.
	Asset2 Asset
	// Details of the current owner of the auction slot, as an Auction Slot object.
	AuctionSlot AuctionSlot `json:",omitempty"`
	// The total outstanding balance of liquidity provider tokens from this AMM instance.
	// The holders of these tokens can vote on the AMM's trading fee in proportion to their holdings, or redeem the tokens for a share of the AMM's assets which grows with the trading fees collected.
	LPTokenBalance types.CurrencyAmount
	// The percentage fee to be charged for trades against this AMM instance, in units of 1/100,000. The maximum value is 1000, for a 1% fee.
	TradingFee uint16
	// A list of vote objects, representing votes on the pool's trading fee.
	VoteSlots []VoteSlots `json:",omitempty"`
	// The identifying hash of the transaction that most recently modified this entry. (Added by the fixPreviousTxnID amendment.)
	PreviousTxnID types.Hash256 `json:",omitempty"`
	// The index of the ledger that contains the transaction that most recently modified this entry. (Added by the fixPreviousTxnID amendment.)
	PreviousTxnLgrSeq uint32 `json:",omitempty"`
}

// ---------------------------------------------
// Asset Object
// ---------------------------------------------

// Asset defines one of the two assets held by the AMM, with currency and optional issuer fields.
type Asset struct {
	Currency string        `json:"currency"`
	Issuer   types.Address `json:"issuer,omitempty"`
}

// Flatten returns the flattened representation of the Asset.
func (a *Asset) Flatten() map[string]interface{} {
	flattened := make(map[string]interface{})

	if a.Issuer.String() != "" {
		flattened["issuer"] = a.Issuer
	}

	if a.Currency != "" {
		flattened["currency"] = a.Currency
	}

	return flattened
}

// AuctionSlot represents the auction slot details for AMM fee discounts, including account, auth accounts, fee, price, and expiration.
type AuctionSlot struct {
	// The current owner of this auction slot.
	Account types.Address
	// A list of at most 4 additional accounts that are authorized to trade at the discounted fee for this AMM instance.
	AuthAccounts []AuthAccounts `json:",omitempty"`
	// The trading fee to be charged to the auction owner, in the same format as TradingFee. Normally, this is 1/10 of the normal fee for this AMM.
	DiscountedFee uint16
	// The amount the auction owner paid to win this slot, in LP Tokens.
	Price types.CurrencyAmount
	// The time when this slot expires, in seconds since the Ripple Epoch. https://xrpl.org/docs/references/protocol/data-types/basic-data-types#specifying-time.
	Expiration uint32
}

// ---------------------------------------------
// AuthAccounts Object
// ---------------------------------------------

// AuthAccounts represents a list of accounts authorized to trade at the discounted fee.
type AuthAccounts struct {
	AuthAccount AuthAccount
}

// Flatten returns the flattened representation of AuthAccounts.
func (a *AuthAccounts) Flatten() map[string]interface{} {
	flattened := make(map[string]interface{})
	flattened["AuthAccount"] = a.AuthAccount.Flatten()
	return flattened
}

// AuthAccount represents an account authorized to trade at the discounted fee for an AMM instance.
type AuthAccount struct {
	// Authorized account to trade at the discounted fee for this AMM instance.
	Account types.Address
}

// Flatten returns the flattened representation of AuthAccount.
func (a *AuthAccount) Flatten() map[string]interface{} {
	flattened := make(map[string]interface{})
	flattened["Account"] = a.Account
	return flattened
}

// ---------------------------------------------
// VoteSlots / Vote Entry Objects
// ---------------------------------------------

// VoteSlots represents one liquidity provider's vote entry in the AMM voting list.
type VoteSlots struct {
	VoteEntry VoteEntry
}

// VoteEntry represents a liquidity provider's vote for setting the AMM trading fee, including account, fee, and weight.
type VoteEntry struct {
	// The account that cast the vote.
	Account types.Address
	// The proposed trading fee, in units of 1/100,000; a value of 1 is equivalent to 0.001%. The maximum value is 1000, indicating a 1% fee.
	TradingFee uint16
	// The weight of the vote, in units of 1/100,000. For example, a value of 1234 means this vote counts as 1.234% of the weighted total vote.
	// The weight is determined by the percentage of this AMM's LP Tokens the account owns. The maximum value is 100000.
	VoteWeight uint32
}

// EntryType returns the ledger entry type for AMM.
func (*AMM) EntryType() EntryType {
	return AMMEntry
}
