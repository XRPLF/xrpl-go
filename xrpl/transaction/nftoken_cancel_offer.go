package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// NFTokenCancelOfferMetadata represents the resulting metadata of a succeeded NFTokenCancelOffer transaction.
// It extends from TxObjMeta.
type NFTokenCancelOfferMetadata struct {
	TxObjMeta

	// rippled 1.11.0 or later.
	NFTokenIDs []types.NFTokenID `json:"nftoken_ids,omitempty"`
}

// The NFTokenCancelOffer transaction can be used to cancel existing token offers created using NFTokenCreateOffer.
//
// Example:
//
// ```json
//
//	{
//		"TransactionType": "NFTokenCancelOffer",
//		"Account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
//		"NFTokenOffers": [
//			"9C92E061381C1EF37A8CDE0E8FC35188BFC30B1883825042A64309AC09F4C36D"
//		]
//	}
//
// ```
type NFTokenCancelOffer struct {
	BaseTx
	// An array of IDs of the NFTokenOffer objects to cancel (not the IDs of NFToken objects, but the IDs of the NFTokenOffer objects).
	// Each entry must be a different object ID of an NFTokenOffer object; the transaction is invalid if the array contains duplicate entries.
	NFTokenOffers []types.NFTokenID
}

// TxType returns the type of the transaction (NFTokenCancelOffer).
func (*NFTokenCancelOffer) TxType() TxType {
	return NFTokenCancelOfferTx
}

// Flatten returns a map of the NFTokenCancelOffer transaction fields.
func (n *NFTokenCancelOffer) Flatten() FlatTransaction {
	flattened := n.BaseTx.Flatten()

	flattened["TransactionType"] = "NFTokenCancelOffer"

	if len(n.NFTokenOffers) > 0 {
		flattenedOffers := make([]string, len(n.NFTokenOffers))
		for i, offer := range n.NFTokenOffers {
			flattenedOffers[i] = offer.String()
		}
		flattened["NFTokenOffers"] = flattenedOffers
	}

	return flattened
}

// Validate checks the validity of the NFTokenCancelOffer fields.
func (n *NFTokenCancelOffer) Validate() (bool, error) {
	ok, err := n.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if len(n.NFTokenOffers) == 0 || n.NFTokenOffers == nil {
		return false, ErrEmptyNFTokenOffers
	}

	return true, nil
}
