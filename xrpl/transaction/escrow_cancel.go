package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// EscrowCancel returns escrowed XRP to the sender.
//
// Example:
//
// ```json
//
//	{
//	    "Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "TransactionType": "EscrowCancel",
//	    "Owner": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "OfferSequence": 7,
//	}
//
// ```
type EscrowCancel struct {
	BaseTx
	// Address of the source account that funded the escrow payment.
	Owner types.Address
	// Transaction sequence (or Ticket number) of EscrowCreate transaction that created the escrow to cancel.
	OfferSequence uint32
}

// TxType returns the transaction type for this transaction (EscrowCancel).
func (*EscrowCancel) TxType() TxType {
	return EscrowCancelTx
}

// Flatten returns the flattened map of the EscrowCancel transaction.
func (e *EscrowCancel) Flatten() FlatTransaction {
	flattened := e.BaseTx.Flatten()

	flattened["TransactionType"] = "EscrowCancel"

	if e.Owner != "" {
		flattened["Owner"] = e.Owner.String()
	}

	if e.OfferSequence != 0 {
		flattened["OfferSequence"] = e.OfferSequence
	}

	return flattened
}

// Validate checks if the EscrowCancel struct is valid.
func (e *EscrowCancel) Validate() (bool, error) {
	ok, err := e.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if !addresscodec.IsValidAddress(e.Owner.String()) {
		return false, ErrEscrowCancelMissingOwner
	}

	if e.OfferSequence == 0 {
		return false, ErrEscrowCancelMissingOfferSequence
	}

	return true, nil
}
