package transaction

import (
	flag "github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TfLoanPayOverpayment indicates that remaining payment amount should be treated as an overpayment.
	TfLoanPayOverpayment uint32 = 0x00010000
	// TfLoanPayFullPayment indicates that the borrower is making a full early repayment.
	TfLoanPayFullPayment uint32 = 0x00020000
	// TfLoanPayLatePayment indicates that the borrower is making a late loan payment.
	TfLoanPayLatePayment uint32 = 0x00040000
)

// LoanPay allows the Borrower to submit a payment on the Loan.
//
// ```json
//
//	{
//	  "TransactionType": "LoanPay",
//	  "Account": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
//	  "LoanID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Amount": "10000"
//	}
//
// ```
type LoanPay struct {
	BaseTx
	// The ID of the Loan object to be paid to.
	LoanID string
	// The amount of funds to pay.
	Amount types.CurrencyAmount
}

// TxType returns the TxType for LoanPay transactions.
func (tx *LoanPay) TxType() TxType {
	return LoanPayTx
}

// Flatten returns a map representation of the LoanPay transaction for JSON-RPC submission.
func (tx *LoanPay) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	flattened["LoanID"] = tx.LoanID

	if tx.Amount != nil {
		flattened["Amount"] = tx.Amount.Flatten()
	}

	return flattened
}

// Validate checks LoanPay transaction fields and returns false with an error if invalid.
func (tx *LoanPay) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.LoanID == "" {
		return false, ErrLoanPayLoanIDRequired
	}

	if !IsLedgerEntryID(tx.LoanID) {
		return false, ErrLoanPayLoanIDInvalid
	}

	if tx.Amount == nil {
		return false, ErrLoanPayAmountRequired
	}

	if ok, err := IsAmount(tx.Amount, "Amount", true); !ok {
		return false, err
	}

	// Validate mutually exclusive flags: at most one of the three payment type flags can be set.
	paymentFlags := []uint32{TfLoanPayOverpayment, TfLoanPayFullPayment, TfLoanPayLatePayment}
	setCount := 0
	for _, f := range paymentFlags {
		if flag.Contains(tx.Flags, f) {
			setCount++
		}
	}
	if setCount > 1 {
		return false, ErrLoanPayMutuallyExclusiveFlags
	}

	return true, nil
}
