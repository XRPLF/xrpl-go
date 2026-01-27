package transaction

import (
	"errors"
)

// LoanDelete deletes an existing Loan object.
//
// ```json
//
//	{
//	  "TransactionType": "LoanDelete",
//	  "Account": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
//	  "LoanID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"
//	}
//
// ```
type LoanDelete struct {
	BaseTx
	// The ID of the Loan object to be deleted.
	LoanID string
}

// TxType returns the TxType for LoanDelete transactions.
func (tx *LoanDelete) TxType() TxType {
	return LoanDeleteTx
}

// Flatten returns a map representation of the LoanDelete transaction for JSON-RPC submission.
func (tx *LoanDelete) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	flattened["LoanID"] = tx.LoanID

	return flattened
}

// Validate checks LoanDelete transaction fields and returns false with an error if invalid.
func (tx *LoanDelete) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.LoanID == "" {
		return false, errors.New("loanDelete: LoanID is required")
	}

	if !IsLedgerEntryID(tx.LoanID) {
		return false, errors.New("loanDelete: LoanID must be 64 characters hexadecimal string")
	}

	return true, nil
}
