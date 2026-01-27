package transaction

import (
	"errors"
	"strconv"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// LoanBrokerCoverClawback claws back the First-Loss Capital from the Loan Broker.
// The transaction can only be submitted by the Issuer of the Loan asset.
// Furthermore, the transaction can only clawback funds up to the minimum cover required for the current loans.
//
// ```json
//
//	{
//	  "TransactionType": "LoanBrokerCoverClawback",
//	  "Account": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
//	  "LoanBrokerID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Amount": {
//	    "currency": "USD",
//	    "issuer": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
//	    "value": "1000"
//	  }
//	}
//
// ```
type LoanBrokerCoverClawback struct {
	BaseTx
	// The Loan Broker ID from which to withdraw First-Loss Capital.
	// Must be provided if the Amount is an MPT, or Amount is an IOU
	// and issuer is specified as the Account submitting the transaction.
	LoanBrokerID *types.LoanBrokerID `json:",omitempty"`
	// The First-Loss Capital amount to clawback.
	// If the amount is 0 or not provided, clawback funds up to LoanBroker.DebtTotal * LoanBroker.CoverRateMinimum.
	Amount types.CurrencyAmount `json:",omitempty"`
}

// TxType returns the TxType for LoanBrokerCoverClawback transactions.
func (tx *LoanBrokerCoverClawback) TxType() TxType {
	return LoanBrokerCoverClawbackTx
}

// Flatten returns a map representation of the LoanBrokerCoverClawback transaction for JSON-RPC submission.
func (tx *LoanBrokerCoverClawback) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	if tx.LoanBrokerID != nil {
		flattened["LoanBrokerID"] = string(*tx.LoanBrokerID)
	}

	if tx.Amount != nil {
		flattened["Amount"] = tx.Amount.Flatten()
	}

	return flattened
}

// Validate checks LoanBrokerCoverClawback transaction fields and returns false with an error if invalid.
func (tx *LoanBrokerCoverClawback) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.LoanBrokerID != nil && *tx.LoanBrokerID != "" {
		if !IsLedgerEntryID(tx.LoanBrokerID.Value()) {
			return false, errors.New("loanBrokerCoverClawback: LoanBrokerID must be 64 characters hexadecimal string")
		}
	}

	if tx.Amount != nil {
		if !IsTokenAmount(tx.Amount) {
			return false, errors.New("loanBrokerCoverClawback: Amount must be an IssuedCurrencyAmount or MPTCurrencyAmount")
		}

		// Check that Amount value is >= 0
		switch amt := tx.Amount.(type) {
		case types.IssuedCurrencyAmount:
			val, err := strconv.ParseFloat(amt.Value, 64)
			if err != nil || val < 0 {
				return false, errors.New("loanBrokerCoverClawback: Amount must be >= 0")
			}
		case types.MPTCurrencyAmount:
			val, err := strconv.ParseFloat(amt.Value, 64)
			if err != nil || val < 0 {
				return false, errors.New("loanBrokerCoverClawback: Amount must be >= 0")
			}
		}
	}

	// At least one of LoanBrokerID or Amount must be provided
	if (tx.LoanBrokerID == nil || *tx.LoanBrokerID == "") && tx.Amount == nil {
		return false, errors.New("loanBrokerCoverClawback: Either LoanBrokerID or Amount is required")
	}

	return true, nil
}
