//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// BookOffer represents an orderbook offer ledger entry, including metadata such as flags, owner, and amounts to get and pay.
type BookOffer struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// The value 0x006F, mapped to the string Offer, indicates that this is an Offer entry.
	LedgerEntryType ledger.EntryType
	// The address of the account that owns this Offer.
	Account types.Address
	// The ID of the Offer Directory that links to this Offer.
	BookDirectory types.Hash256
	// A hint indicating which page of the offer directory links to this entry, in case the
	// directory consists of multiple pages.
	BookNode string
	// 	Indicates the time after which this Offer is considered unfunded. See Specifying Time for details.
	Expiration uint32 `json:",omitempty"`
	// A hint indicating which page of the owner directory links to this entry, in case the directory consists of multiple pages.
	OwnerNode string
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this object.
	PreviousTxnLgrSeq uint32
	// 	The Sequence value of the OfferCreate transaction that created this offer.
	// Used in combination with the Account to identify this offer.
	Sequence uint32
	// The remaining amount and type of currency requested by the Offer creator.
	TakerPays any
	// The remaining amount and type of currency being provided by the Offer creator.
	TakerGets  any
	OwnerFunds string `json:"owner_funds,omitempty"`
	// TakerGetsFunded types.CurrencyAmount `json:"taker_gets_funded,omitempty"`
	TakerGetsFunded any `json:"taker_gets_funded,omitempty"`
	// TakerPaysFunded types.CurrencyAmount `json:"taker_pays_funded,omitempty"`
	TakerPaysFunded any    `json:"taker_pays_funded,omitempty"`
	Quality         string `json:"quality,omitempty"`
}

// BookOfferCurrency represents a currency and optional issuer in a book offer.
type BookOfferCurrency struct {
	Currency string `json:"currency"`
	Issuer   string `json:"issuer,omitempty"`
}
