package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// OfferCreate transaction places an Offer in the decentralized exchange.
//
// Example:
//
// ```json
//
//	{
//	    "TransactionType": "OfferCreate",
//	    "Account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
//	    "Fee": "12",
//	    "Flags": 0,
//	    "LastLedgerSequence": 7108682,
//	    "Sequence": 8,
//	    "TakerGets": "6000000",
//	    "TakerPays": {
//	      "currency": "GKO",
//	      "issuer": "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
//	      "value": "2"
//	    }
//	}
//
// ```
type OfferCreate struct {
	BaseTx
	// (Optional) Time after which the Offer is no longer active, in seconds since the Ripple Epoch.
	Expiration uint32 `json:",omitempty"`
	// (Optional) An Offer to delete first, specified in the same way as OfferCancel.
	OfferSequence uint32 `json:",omitempty"`
	// The amount and type of currency being sold.
	TakerGets types.CurrencyAmount
	// The amount and type of currency being bought.
	TakerPays types.CurrencyAmount
	// The domain that the offer must be a part of.
	DomainID *string `json:",omitempty"`
}

// **********************************
// OfferCreate Flags
// **********************************

const (
	// TfPassive indicates that the offer is passive, meaning it does not consume offers that exactly match it, and instead waits to be consumed by an offer that exactly matches it.
	TfPassive uint32 = 65536
	// TfImmediateOrCancel treats the Offer as an Immediate or Cancel order. The Offer never creates an Offer object in the ledger: it only trades as much as it can by consuming existing Offers at the time the transaction is processed. If no Offers match, it executes "successfully" without trading anything. In this case, the transaction still uses the result code tesSUCCESS.
	TfImmediateOrCancel uint32 = 131072
	// TfFillOrKill treats the offer as a Fill or Kill order. The Offer never creates an Offer object in the ledger, and is canceled if it cannot be fully filled at the time of execution. By default, this means that the owner must receive the full TakerPays amount; if the TfSell flag is enabled, the owner must be able to spend the entire TakerGets amount instead.
	TfFillOrKill uint32 = 262144
	// TfSell indicates that the offer is selling, not buying.
	TfSell uint32 = 524288
	// TfHybrid indicates the offer is hybrid. (meaning it is part of both a domain and open order book)
	// This flag cannot be set if the offer doesn't have a DomainID
	TfHybrid uint32 = 0x00100000
)

// SetPassiveFlag sets the TfPassive flag, indicating the offer is passive and will not consume exactly matching offers.
func (o *OfferCreate) SetPassiveFlag() {
	o.Flags |= TfPassive
}

// SetImmediateOrCancelFlag sets the TfImmediateOrCancel flag, treating the offer as an Immediate or Cancel order.
// It executes against existing offers only and never creates a new ledger entry.
func (o *OfferCreate) SetImmediateOrCancelFlag() {
	o.Flags |= TfImmediateOrCancel
}

// SetFillOrKillFlag sets the TfFillOrKill flag, treating the offer as a Fill or Kill order.
// The offer is canceled if it cannot be fully filled immediately.
func (o *OfferCreate) SetFillOrKillFlag() {
	o.Flags |= TfFillOrKill
}

// SetSellFlag sets the TfSell flag, indicating the offer is selling rather than buying.
func (o *OfferCreate) SetSellFlag() {
	o.Flags |= TfSell
}

// SetHybridFlag sets the TfHybrid, indicating the offer is hybrid.
func (o *OfferCreate) SetHybridFlag() {
	o.Flags |= TfHybrid
}

// TxType returns the type of the transaction (OfferCreate).
func (*OfferCreate) TxType() TxType {
	return OfferCreateTx
}

// Flatten returns a map of the OfferCreate transaction fields.
func (o *OfferCreate) Flatten() FlatTransaction {
	flattened := o.BaseTx.Flatten()

	flattened["TransactionType"] = o.TxType().String()

	if o.Expiration != 0 {
		flattened["Expiration"] = o.Expiration
	}
	if o.OfferSequence != 0 {
		flattened["OfferSequence"] = o.OfferSequence
	}
	flattened["TakerGets"] = o.TakerGets.Flatten()
	flattened["TakerPays"] = o.TakerPays.Flatten()

	if o.DomainID != nil {
		flattened["DomainID"] = *o.DomainID
	}

	return flattened
}

// Validate validates the OfferCreate transaction.
func (o *OfferCreate) Validate() (bool, error) {
	_, err := o.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if ok, err := IsAmount(o.TakerGets, "TakerGets", true); !ok {
		return false, err
	}

	if ok, err := IsAmount(o.TakerPays, "TakerPays", true); !ok {
		return false, err
	}

	if o.DomainID == nil && flag.Contains(o.Flags, TfHybrid) {
		return false, ErrTfHybridCannotBeSetWithoutDomainID
	}

	if o.DomainID != nil {
		if ok := IsDomainID(*o.DomainID); !ok {
			return false, ErrInvalidDomainID
		}
	}

	return true, nil
}
