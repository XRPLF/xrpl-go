package transaction

// OfferCancel transaction removes an Offer object from the XRP Ledger.
//
// Example:
//
// ```json
//
//	{
//	    "TransactionType": "OfferCancel",
//	    "Account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
//	    "Fee": "12",
//	    "Flags": 0,
//	    "LastLedgerSequence": 7108629,
//	    "OfferSequence": 6,
//	    "Sequence": 7
//	}
//
// ```
type OfferCancel struct {
	BaseTx
	// The sequence number (or Ticket number) of a previous OfferCreate transaction.
	// If specified, cancel any offer object in the ledger that was created by that transaction. It is not considered an error if the offer specified does not exist.
	OfferSequence uint32
}

// TxType returns the transaction type for OfferCancel.
func (*OfferCancel) TxType() TxType {
	return OfferCancelTx
}

// Flatten returns the flattened map of the OfferCancel transaction.
func (o *OfferCancel) Flatten() FlatTransaction {
	flattened := o.BaseTx.Flatten()

	flattened["TransactionType"] = o.TxType().String()

	flattened["OfferSequence"] = o.OfferSequence
	return flattened
}

// Validate checks the OfferCancel transaction for correctness.
func (o *OfferCancel) Validate() (bool, error) {
	return o.BaseTx.Validate()
}
