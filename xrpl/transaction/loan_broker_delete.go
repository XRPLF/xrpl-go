package transaction

// LoanBrokerDelete deletes LoanBroker ledger object.
//
// ```json
//
//	{
//	  "TransactionType": "LoanBrokerDelete",
//	  "Account": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
//	  "LoanBrokerID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"
//	}
//
// ```
type LoanBrokerDelete struct {
	BaseTx
	// The Loan Broker ID that the transaction is deleting.
	LoanBrokerID string
}

// TxType returns the TxType for LoanBrokerDelete transactions.
func (tx *LoanBrokerDelete) TxType() TxType {
	return LoanBrokerDeleteTx
}

// Flatten returns a map representation of the LoanBrokerDelete transaction for JSON-RPC submission.
func (tx *LoanBrokerDelete) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	flattened["LoanBrokerID"] = tx.LoanBrokerID

	return flattened
}

// Validate checks LoanBrokerDelete transaction fields and returns false with an error if invalid.
func (tx *LoanBrokerDelete) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.LoanBrokerID == "" {
		return false, ErrLoanBrokerDeleteLoanBrokerIDRequired
	}

	if !IsLedgerEntryID(tx.LoanBrokerID) {
		return false, ErrLoanBrokerDeleteLoanBrokerIDInvalid
	}

	return true, nil
}
