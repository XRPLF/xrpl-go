package ledger

import (
	"encoding/json"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// NFTokenOffer entry represents an offer to buy, sell or transfer an NFT.
// (Added by the NonFungibleTokensV1_1 amendment.)
//
// ```json
//
//	{
//	    "Amount": "1000000",
//	    "Flags": 1,
//	    "LedgerEntryType": "NFTokenOffer",
//	    "NFTokenID": "00081B5825A08C22787716FA031B432EBBC1B101BB54875F0002D2A400000000",
//	    "NFTokenOfferNode": "0",
//	    "Owner": "rhRxL3MNvuKEjWjL7TBbZSDacb8PmzAd7m",
//	    "OwnerNode": "17",
//	    "PreviousTxnID": "BFA9BE27383FA315651E26FDE1FA30815C5A5D0544EE10EC33D3E92532993769",
//	    "PreviousTxnLgrSeq": 75443565,
//	    "index": "AEBABA4FAC212BF28E0F9A9C3788A47B085557EC5D1429E7A8266FB859C863B3"
//	}
//
// ```
type NFTokenOffer struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// The value 0x0037, mapped to the string NFTokenOffer, indicates that this is an offer to trade a NFToken.
	LedgerEntryType EntryType
	// Amount expected or offered for the NFToken. If the token has the LsfOnlyXRP flag set,
	// the amount must be specified in XRP. Sell offers that specify assets other than XRP
	// must specify a non-zero amount. Sell offers that specify XRP can be 'free' (that is,
	// the Amount field can be equal to "0").
	Amount types.CurrencyAmount
	// The AccountID for which this offer is intended. If present, only that account can accept the offer.
	Destination types.Address `json:",omitempty"`
	// The time after which the offer is no longer active. The value is the number of seconds
	// since the Ripple Epoch.
	Expiration uint32 `json:",omitempty"`
	// The NFTokenID of the NFToken object referenced by this offer.
	NFTokenID types.Hash256
	// Internal bookkeeping, indicating the page inside the token buy or sell offer directory,
	// as appropriate, where this token is being tracked. This field allows the efficient
	// deletion of offers.
	NFTokenOfferNode string `json:",omitempty"`
	// Owner of the account that is creating and owns the offer. Only the current Owner of an
	// NFToken can create an offer to sell an NFToken, but any account can create an offer
	// to buy an NFToken.
	Owner types.Address
	// Internal bookkeeping, indicating the page inside the owner directory where this token is
	// being tracked. This field allows the efficient deletion of offers.
	OwnerNode string `json:",omitempty"`
	// Identifying hash of the transaction that most recently modified this object.
	PreviousTxnID types.Hash256
	// Index of the ledger that contains the transaction that most recently modified this object.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the type of the ledger entry.
func (*NFTokenOffer) EntryType() EntryType {
	return NFTokenOfferEntry

}

// UnmarshalJSON implements custom JSON unmarshalling for NFTokenOffer.
func (n *NFTokenOffer) UnmarshalJSON(data []byte) error {
	type nftHelper struct {
		Amount            json.RawMessage
		Destination       types.Address
		Expiration        uint32
		Flags             uint32
		LedgerEntryType   EntryType
		NFTokenID         types.Hash256
		NFTokenOfferNode  string
		Owner             types.Address
		OwnerNode         string
		PreviousTxnID     types.Hash256
		PreviousTxnLgrSeq uint32
	}
	var h nftHelper
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}
	*n = NFTokenOffer{
		Destination:       h.Destination,
		Expiration:        h.Expiration,
		Flags:             h.Flags,
		LedgerEntryType:   h.LedgerEntryType,
		NFTokenID:         h.NFTokenID,
		NFTokenOfferNode:  h.NFTokenOfferNode,
		Owner:             h.Owner,
		OwnerNode:         h.OwnerNode,
		PreviousTxnID:     h.PreviousTxnID,
		PreviousTxnLgrSeq: h.PreviousTxnLgrSeq,
	}
	amnt, err := types.UnmarshalCurrencyAmount(h.Amount)
	if err != nil {
		return err
	}
	n.Amount = amnt
	return nil
}
