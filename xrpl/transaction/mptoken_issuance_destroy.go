package transaction

// MPTokenIssuanceDestroy transaction is used to remove an MPTokenIssuance object from the directory node
// in which it is being held, effectively removing the token from the ledger ("destroying" it).
//
// If this operation succeeds, the corresponding MPTokenIssuance is removed and the owner’s reserve requirement is reduced by one.
// This operation must fail if there are any holders of the MPT in question.
//
// ```json
//
//	 {
//	     "TransactionType": "MPTokenIssuanceDestroy",
//	     "Fee": "10",
//	     "MPTokenIssuanceID": "000004C463C52827307480341125DA0577DEFC38405B0E3E"
//	}
//
// ```
type MPTokenIssuanceDestroy struct {
	BaseTx
	// Identifies the MPTokenIssuance object to be removed by the transaction.
	MPTokenIssuanceID string
}

// TxType returns the type of the transaction (MPTokenIssuanceDestroy).
func (*MPTokenIssuanceDestroy) TxType() TxType {
	return MPTokenIssuanceDestroyTx
}

// Flatten returns the flattened map of the MPTokenIssuanceDestroy transaction.
func (m *MPTokenIssuanceDestroy) Flatten() FlatTransaction {
	flattened := m.BaseTx.Flatten()

	flattened["TransactionType"] = "MPTokenIssuanceDestroy"

	flattened["MPTokenIssuanceID"] = m.MPTokenIssuanceID

	return flattened
}

// Validate validates the MPTokenIssuanceDestroy transaction ensuring all fields are correct.
func (m *MPTokenIssuanceDestroy) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if _, ok := decodeMPTIssuanceID(m.MPTokenIssuanceID); !ok {
		return false, ErrInvalidMPTokenIssuanceIDDestroy
	}

	return true, nil
}
