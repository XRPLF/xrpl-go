package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// LoanBrokerCoverDeposit deposits First Loss Capital into the LoanBroker object.
//
// ```json
//
//	{
//	  "TransactionType": "LoanBrokerCoverDeposit",
//	  "Account": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
//	  "LoanBrokerID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Amount": "10000"
//	}
//
// ```
type LoanBrokerCoverDeposit struct {
	BaseTx
	// The Loan Broker ID to deposit First-Loss Capital.
	LoanBrokerID string
	// The First-Loss Capital amount to deposit.
	Amount types.CurrencyAmount
}

// TxType returns the TxType for LoanBrokerCoverDeposit transactions.
func (tx *LoanBrokerCoverDeposit) TxType() TxType {
	return LoanBrokerCoverDepositTx
}

// Flatten returns a map representation of the LoanBrokerCoverDeposit transaction for JSON-RPC submission.
func (tx *LoanBrokerCoverDeposit) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	flattened["LoanBrokerID"] = tx.LoanBrokerID

	if tx.Amount != nil {
		flattened["Amount"] = tx.Amount.Flatten()
	}

	return flattened
}

// Validate checks LoanBrokerCoverDeposit transaction fields and returns false with an error if invalid.
func (tx *LoanBrokerCoverDeposit) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.LoanBrokerID == "" {
		return false, ErrLoanBrokerCoverDepositLoanBrokerIDRequired
	}

	if !IsLedgerEntryID(tx.LoanBrokerID) {
		return false, ErrLoanBrokerCoverDepositLoanBrokerIDInvalid
	}

	if tx.Amount == nil {
		return false, ErrLoanBrokerCoverDepositAmountRequired
	}

	if ok, err := IsAmount(tx.Amount, "Amount", true); !ok {
		return false, err
	}

	return true, nil
}
