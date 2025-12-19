package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// NFTokenAcceptOfferMetadata represents the resulting metadata of a succeeded NFTokenAcceptOffer transaction.
// It extends from TxObjMeta.
type NFTokenAcceptOfferMetadata struct {
	TxObjMeta

	// rippled 1.11.0 or later.
	NFTokenID *types.NFTokenID `json:"nftoken_id,omitempty"`
}

// NFTokenAcceptOffer transaction is used to accept offers to buy or sell an NFToken. It can either:
//
// - Allow one offer to be accepted. This is called direct mode.
//
// - Allow two distinct offers, one offering to buy a given NFToken and the other offering to sell the same NFToken, to be accepted in an atomic fashion. This is called brokered mode.
//
// Example:
//
// ```json
//
//	{
//		"Account": "r9spUPhPBfB6kQeF6vPhwmtFwRhBh2JUCG",
//		"Fee": "12",
//		"LastLedgerSequence": 75447550,
//		"Memos": [
//		  {
//			"Memo": {
//			  "MemoData": "61356534373538372D633134322D346663382D616466362D393666383562356435386437"
//			}
//		  }
//		],
//		"NFTokenSellOffer": "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
//		"Sequence": 68549302,
//		"TransactionType": "NFTokenAcceptOffer"
//	  }
//
// / ```
type NFTokenAcceptOffer struct {
	BaseTx
	// (Optional) Identifies the NFTokenOffer that offers to sell the NFToken.
	NFTokenSellOffer types.Hash256 `json:",omitempty"`
	// (Optional) Identifies the NFTokenOffer that offers to buy the NFToken.
	NFTokenBuyOffer types.Hash256 `json:",omitempty"`
	// (Optional) This field is only valid in brokered mode, and specifies the amount that the broker keeps as part of their fee for bringing the two offers together; the remaining amount is sent to the seller of the NFToken being bought.
	// If specified, the fee must be such that, before applying the transfer fee, the amount that the seller would receive is at least as much as the amount indicated in the sell offer.
	NFTokenBrokerFee types.CurrencyAmount `json:",omitempty"`
}

// TxType returns the type of the transaction (NFTokenAcceptOffer).
func (*NFTokenAcceptOffer) TxType() TxType {
	return NFTokenAcceptOfferTx
}

// Flatten returns a map of the NFTokenAcceptOffer transaction fields.
func (n *NFTokenAcceptOffer) Flatten() FlatTransaction {
	flattened := n.BaseTx.Flatten()

	flattened["TransactionType"] = "NFTokenAcceptOffer"

	if n.NFTokenSellOffer != "" {
		flattened["NFTokenSellOffer"] = n.NFTokenSellOffer.String()
	}

	if n.NFTokenBuyOffer != "" {
		flattened["NFTokenBuyOffer"] = n.NFTokenBuyOffer.String()
	}

	if n.NFTokenBrokerFee != nil {
		flattened["NFTokenBrokerFee"] = n.NFTokenBrokerFee.Flatten()
	}

	return flattened
}

// Validate checks the validity of the NFTokenAcceptOffer fields.
func (n *NFTokenAcceptOffer) Validate() (bool, error) {
	ok, err := n.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if xrpAmount, ok := n.NFTokenBrokerFee.(types.XRPCurrencyAmount); ok {
		if xrpAmount.Uint64() == 0 {
			return false, ErrNFTokenBrokerFeeZero
		}
	} else if issuedAmount, ok := n.NFTokenBrokerFee.(types.IssuedCurrencyAmount); ok {
		if issuedAmount.Value == "0" {
			return false, ErrNFTokenBrokerFeeZero
		}
	}

	// if NFTokenBrokerFee is set, then both NFTokenSellOffer and NFTokenBuyOffer must be set
	if n.NFTokenBrokerFee != nil && (n.NFTokenSellOffer == "" || n.NFTokenBuyOffer == "") {
		return false, ErrMissingBothOffers
	}

	// check either NFTokenSellOffer or NFTokenBuyOffer is set
	if n.NFTokenSellOffer == "" && n.NFTokenBuyOffer == "" {
		return false, ErrMissingOffer
	}

	return true, nil
}
