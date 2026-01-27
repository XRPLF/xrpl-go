package transaction

import (
	"errors"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// LoanBrokerCoverWithdraw withdraws the First-Loss Capital from the LoanBroker.
//
// ```json
//
//	{
//	  "TransactionType": "LoanBrokerCoverWithdraw",
//	  "Account": "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
//	  "LoanBrokerID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Amount": "10000",
//	  "Destination": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5"
//	}
//
// ```
type LoanBrokerCoverWithdraw struct {
	BaseTx
	// The Loan Broker ID from which to withdraw First-Loss Capital.
	LoanBrokerID string
	// The First-Loss Capital amount to withdraw.
	Amount types.CurrencyAmount
	// An account to receive the assets. It must be able to receive the asset.
	Destination *types.Address `json:",omitempty"`
}

// TxType returns the TxType for LoanBrokerCoverWithdraw transactions.
func (tx *LoanBrokerCoverWithdraw) TxType() TxType {
	return LoanBrokerCoverWithdrawTx
}

// Flatten returns a map representation of the LoanBrokerCoverWithdraw transaction for JSON-RPC submission.
func (tx *LoanBrokerCoverWithdraw) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	flattened["LoanBrokerID"] = tx.LoanBrokerID

	if tx.Amount != nil {
		flattened["Amount"] = tx.Amount.Flatten()
	}

	if tx.Destination != nil {
		flattened["Destination"] = tx.Destination.String()
	}

	return flattened
}

// Validate checks LoanBrokerCoverWithdraw transaction fields and returns false with an error if invalid.
func (tx *LoanBrokerCoverWithdraw) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.LoanBrokerID == "" {
		return false, errors.New("loanBrokerCoverWithdraw: LoanBrokerID is required")
	}

	if !IsLedgerEntryID(tx.LoanBrokerID) {
		return false, errors.New("loanBrokerCoverWithdraw: LoanBrokerID must be 64 characters hexadecimal string")
	}

	if tx.Amount == nil {
		return false, errors.New("loanBrokerCoverWithdraw: Amount is required")
	}

	if ok, err := IsAmount(tx.Amount, "Amount", true); !ok {
		return false, err
	}

	if tx.Destination != nil {
		if !addresscodec.IsValidAddress(tx.Destination.String()) {
			return false, ErrInvalidDestination
		}
	}

	return true, nil
}
