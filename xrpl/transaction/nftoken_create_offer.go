package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// **********************************
// NFTokenCreateOffer Flags
// **********************************

const (
	// If enabled, indicates that the offer is a sell offer. Otherwise, it is a buy offer.
	tfSellNFToken uint32 = 1
)

// NFTokenCreateOfferMetadata represents the resulting metadata of a succeeded NFTokenCreateOffer transaction.
// It extends from TxObjMeta.
type NFTokenCreateOfferMetadata struct {
	TxObjMeta

	// OfferID is a string of Amount is present.
	OfferID *types.OfferID `json:"offer_id,omitempty"`
}

// NFTokenCreateOffer creates either a new Sell offer for an NFToken owned by the account executing the transaction, or a new Buy offer for an NFToken owned by another account.
//
// If successful, the transaction creates a NFTokenOffer object. Each offer counts as one object towards the owner reserve of the account that placed the offer.
//
// Example:
//
// ```json
//
//	{
//		"TransactionType": "NFTokenCreateOffer",
//		"Account": "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
//		"NFTokenID": "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
//		"Amount": "1000000",
//		"Flags": 1
//	}
//
// ```
type NFTokenCreateOffer struct {
	BaseTx
	// (Optional) Who owns the corresponding NFToken.
	// If the offer is to buy a token, this field must be present and it must be different than the Account field (since an offer to buy a token one already holds is meaningless).
	// If the offer is to sell a token, this field must not be present, as the owner is, implicitly, the same as the Account (since an offer to sell a token one doesn't already hold is meaningless).
	Owner types.Address `json:",omitempty"`
	// Identifies the NFToken object that the offer references.
	NFTokenID types.NFTokenID
	// Indicates the amount expected or offered for the corresponding NFToken.
	// The amount must be non-zero, except where this is an offer to sell and the asset is XRP; then, it is legal to specify an amount of zero, which means that the current owner of the token is giving it away, gratis, either to anyone at all, or to the account identified by the Destination field.
	Amount types.CurrencyAmount
	// (Optional) Time after which the offer is no longer active, in seconds since the Ripple Epoch.
	Expiration uint32 `json:",omitempty"`
	// (Optional) If present, indicates that this offer may only be accepted by the specified account. Attempts by other accounts to accept this offer MUST fail.
	Destination types.Address `json:",omitempty"`
}

// SetSellNFTokenFlag sets the flag indicating the offer is a sell offer.
func (n *NFTokenCreateOffer) SetSellNFTokenFlag() {
	n.Flags |= tfSellNFToken
}

// TxType returns the type of the transaction (NFTokenCreateOffer).
func (*NFTokenCreateOffer) TxType() TxType {
	return NFTokenCreateOfferTx
}

// Flatten returns a map of the NFTokenCreateOffer transaction fields.
func (n *NFTokenCreateOffer) Flatten() FlatTransaction {
	flattened := n.BaseTx.Flatten()

	flattened["TransactionType"] = "NFTokenCreateOffer"

	if n.Owner != "" {
		flattened["Owner"] = n.Owner.String()
	}

	flattened["NFTokenID"] = n.NFTokenID.String()
	flattened["Amount"] = n.Amount.Flatten()

	if n.Expiration != 0 {
		flattened["Expiration"] = n.Expiration
	}

	if n.Destination != "" {
		flattened["Destination"] = n.Destination.String()
	}

	return flattened
}

// Validate checks the validity of the NFTokenCreateOffer fields.
func (n *NFTokenCreateOffer) Validate() (bool, error) {
	ok, err := n.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// check owner and account are not equal
	if n.Owner == n.Account {
		return false, ErrOwnerAccountConflict
	}

	// check account and destination are not equal
	if n.Destination == n.Account {
		return false, ErrDestinationAccountConflict
	}

	// check owner is a valid xrpl address
	if n.Owner != "" && !addresscodec.IsValidAddress(n.Owner.String()) {
		return false, ErrInvalidOwner
	}

	// check destination is a valid xrpl address
	if n.Destination != "" && !addresscodec.IsValidAddress(n.Destination.String()) {
		return false, ErrInvalidDestination
	}

	// validate Sell Offer Cases
	if types.IsFlagEnabled(n.Flags, tfSellNFToken) && n.Owner != "" {
		return false, ErrOwnerPresentForSellOffer
	}

	// validate Buy Offer Cases
	if !types.IsFlagEnabled(n.Flags, tfSellNFToken) && n.Owner == "" {
		return false, ErrOwnerNotPresentForBuyOffer
	}

	return true, nil
}
