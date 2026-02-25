package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// LsfNFTokenBuyOffers this directory contains NFT buy offers.
	LsfNFTokenBuyOffers uint32 = 0x00000001
	// LsfNFTokenSellOffers this directory contains NFT sell offers.
	LsfNFTokenSellOffers uint32 = 0x00000002
)

// DirectoryNode ledger entry type provides a list of links to other entries in the ledger's state data.
// A single conceptual Directory takes the form of a doubly linked list, with one or more DirectoryNode entries
// each containing up to 32 IDs of other entries.
// The first DirectoryNode entry is called the root of the directory, and all entries other than the root can be added or deleted as necessary.
//
// DirectoryNode entries can have the following values in the Flags field:
// - LsfNFTokenBuyOffers: This directory contains NFT buy offers.
// - LsfNFTokenSellOffers: This directory contains NFT sell offers.
//
// Owner directories and offer directories for fungible tokens do not use flags; their Flags value is always 0.
//
// There are three kinds of directory:
// - Owner directories list other entries owned by an account, such as RippleState (trust line) or Offer entries.
// - Offer directories list the offers available in the decentralized exchange. A single Offer directory contains all the offers that have the same exchange rate for the same token (currency code and issuer).
// - NFT Offer directories list buy and sell offers for NFTs. Each NFT has up to two directories, one for buy offers, the other for sell offers.
//
// ```json
//
//	{
//	    "ExchangeRate": "4e133c40576f7c00",
//	    "Flags": 0,
//	    "Indexes": [
//	        "353E55E7A0B0E82D16DF6E748D48BDAFE4C56045DF5A8B0ED723FF3C38A4787A"
//	    ],
//	    "LedgerEntryType": "DirectoryNode",
//	    "PreviousTxnID": "0F79E60C8642A23658ECB29D939499EA0F28D804077B7EE16613BE0C813A2DD6",
//	    "PreviousTxnLgrSeq": 91448326,
//	    "RootIndex": "79C54A4EBD69AB2EADCE313042F36092BE432423CC6A4F784E133C40576F7C00",
//	    "TakerGetsCurrency": "0000000000000000000000000000000000000000",
//	    "TakerGetsIssuer": "0000000000000000000000000000000000000000",
//	    "TakerPaysCurrency": "0000000000000000000000005553440000000000",
//	    "TakerPaysIssuer": "2ADB0B3959D60A6E6991F729E1918B7163925230",
//	}
//
// ```
type DirectoryNode struct {
	// (Offer directories only) DEPRECATED. Do not use.
	ExchangeRate string `json:",omitempty"`
	// A bit-map of boolean flags enabled for this object. Currently, the protocol defines no flags
	// for DirectoryNode objects. The value is always 0.
	Flags uint32
	// The contents of this directory: an array of IDs of other objects.
	Indexes []types.Hash256
	// If this directory consists of multiple pages, this ID links to the next object in the chain, wrapping around at the end.
	IndexNext string `json:",omitempty"`
	// If this directory consists of multiple pages, this ID links to the previous object in the chain, wrapping around at the beginning.
	IndexPrevious string `json:",omitempty"`
	// The type of ledger entry. Always DirectoryNode.
	LedgerEntryType EntryType
	// (NFT offer directories only) ID of the NFT in a buy or sell offer.
	NFTokenID types.Hash256 `json:",omitempty"`
	// (Owner directories only) The address of the account that owns the objects in this directory.
	Owner types.Address `json:",omitempty"`
	// The identifying hash of the transaction that most recently modified this entry. (Added by the fixPreviousTxnID amendment.)
	PreviousTxnID types.Hash256 `json:",omitempty"`
	// The index of the ledger that contains the transaction that most recently modified this entry. (Added by the fixPreviousTxnID amendment.)
	PreviousTxnLgrSeq uint32 `json:",omitempty"`
	// The ID of root object for this directory.
	RootIndex types.Hash256
	// (Offer directories only) The currency code of the TakerGets amount from the offers in this directory.
	TakerGetsCurrency string `json:",omitempty"`
	// (Offer directories only) The issuer of the TakerGets amount from the offers in this directory.
	TakerGetsIssuer string `json:",omitempty"`
	// (Offer directories only) The currency code of the TakerPays amount from the offers in this directory.
	TakerPaysCurrency string `json:",omitempty"`
	// (Offer directories only) The issuer of the TakerPays amount from the offers in this directory.
	TakerPaysIssuer string `json:",omitempty"`
	// The domain that the offer directory is a part of.
	DomainID string `json:",omitempty"`
}

// EntryType returns the type of ledger entry.
func (*DirectoryNode) EntryType() EntryType {
	return DirectoryNodeEntry
}

// SetNFTokenBuyOffers sets the directory to contain NFT buy offers.
func (d *DirectoryNode) SetNFTokenBuyOffers() {
	d.Flags |= LsfNFTokenBuyOffers
}

// SetNFTokenSellOffers sets the directory to contain NFT sell offers.
func (d *DirectoryNode) SetNFTokenSellOffers() {
	d.Flags |= LsfNFTokenSellOffers
}
