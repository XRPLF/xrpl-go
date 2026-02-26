package ledger

import (
	"encoding/json"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// LsfPassive the offer was placed as "passive". This has no effect after the offer is placed into
	// the ledger.
	LsfPassive uint32 = 0x00010000
	// LsfSell the offer was placed as a "Sell" offer. This has no effect after the offer is placed
	// in the ledger, because TfSell only matters if you get a better rate than you asked for
	// which can only happen when the offer is initially placed.
	LsfSell uint32 = 0x00020000
	// LsfHybrid indicates the offer is hybrid. (meaning it is part of both a domain and open order book)
	// This flag cannot be set if the offer doesn't have a DomainID.
	LsfHybrid uint32 = 0x00040000
)

// Book represents an offer book.
//
// ```json
//
//	{
//	    "BookDirectory": "ACC27DE91DBA86FC509069EAF4BC511D73128B780F2E54BF5E07A369E2446000",
//	    "BookNode": "0000000000000000"
//	}
//
// ```
type Book struct {
	Book struct {
		// The ID of the offer directory that links to this offer.
		BookDirectory string
		// A hint indicating which page of the offer directory links to this entry,
		// in case the directory consists of multiple pages.
		BookNode string
	}
}

// Offer ledger entry describes an Offer to exchange currencies in the XRP Ledger's
// decentralized exchange. (In finance, this is more traditionally known as an order.)
// An OfferCreate transaction only creates an Offer entry in the ledger when the Offer
// cannot be fully executed immediately by consuming other Offers already in the ledger.
//
// An Offer can become unfunded through other activities in the network, while remaining
// in the ledger. When processing transactions, the network automatically removes any
// unfunded Offers that those transactions come across. (Otherwise, unfunded Offers remain,
// because only transactions can change the ledger state.)
//
// ```json
//
//	{
//	    "Account": "rBqb89MRQJnMPq8wTwEbtz4kvxrEDfcYvt",
//	    "BookDirectory": "ACC27DE91DBA86FC509069EAF4BC511D73128B780F2E54BF5E07A369E2446000",
//	    "BookNode": "0000000000000000",
//	    "Flags": 131072,
//	    "LedgerEntryType": "Offer",
//	    "OwnerNode": "0000000000000000",
//	    "PreviousTxnID": "F0AB71E777B2DA54B86231E19B82554EF1F8211F92ECA473121C655BFC5329BF",
//	    "PreviousTxnLgrSeq": 14524914,
//	    "Sequence": 866,
//	    "TakerGets": {
//	        "currency": "XAG",
//	        "issuer": "r9Dr5xwkeLegBeXq6ujinjSBLQzQ1zQGjH",
//	        "value": "37"
//	    },
//	    "TakerPays": "79550000000",
//	    "index": "96F76F27D8A327FC48753167EC04A46AA0E382E6F57F32FD12274144D00F1797"
//	}
//
// ```
type Offer struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// The value 0x006F, mapped to the string Offer, indicates that this is an Offer entry.
	LedgerEntryType EntryType
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
	TakerPays types.CurrencyAmount
	// The remaining amount and type of currency being provided by the Offer creator.
	TakerGets types.CurrencyAmount
	// The domain that the offer must be a part of.
	DomainID *string `json:",omitempty"`
	// An additional list of order book directories that this offer belongs to.
	// Currently, this field only applicable to hybrid offers.
	AdditionalBooks []Book `json:",omitempty"`
}

// EntryType returns the type of the ledger entry.
func (*Offer) EntryType() EntryType {
	return OfferEntry
}

// SetLsfPassive sets the offer as passive.
func (o *Offer) SetLsfPassive() {
	o.Flags |= LsfPassive
}

// SetLsfSell sets the offer as a sell offer.
func (o *Offer) SetLsfSell() {
	o.Flags |= LsfSell
}

// SetLsfHybrid sets the hybrid flag.
func (o *Offer) SetLsfHybrid() {
	o.Flags |= LsfHybrid
}

// UnmarshalJSON unmarshals the offer from a JSON byte slice.
func (o *Offer) UnmarshalJSON(data []byte) error {
	type offerHelper struct {
		Account           types.Address
		BookDirectory     types.Hash256
		BookNode          string
		Expiration        uint32
		Flags             uint32
		LedgerEntryType   EntryType
		OwnerNode         string
		PreviousTxnID     types.Hash256
		PreviousTxnLgrSeq uint32
		Sequence          uint32
		TakerPays         json.RawMessage
		TakerGets         json.RawMessage
		DomainID          string `json:",omitempty"`
		AdditionalBooks   []Book `json:",omitempty"`
	}
	var h offerHelper
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}
	*o = Offer{
		Account:           h.Account,
		BookDirectory:     h.BookDirectory,
		BookNode:          h.BookNode,
		Expiration:        h.Expiration,
		Flags:             h.Flags,
		LedgerEntryType:   h.LedgerEntryType,
		OwnerNode:         h.OwnerNode,
		PreviousTxnID:     h.PreviousTxnID,
		PreviousTxnLgrSeq: h.PreviousTxnLgrSeq,
		Sequence:          h.Sequence,
		AdditionalBooks:   h.AdditionalBooks,
	}
	if h.DomainID != "" {
		o.DomainID = &h.DomainID
	}
	pays, err := types.UnmarshalCurrencyAmount(h.TakerPays)
	if err != nil {
		return err
	}
	gets, err := types.UnmarshalCurrencyAmount(h.TakerGets)
	if err != nil {
		return err
	}
	o.TakerPays = pays
	o.TakerGets = gets
	return nil
}
