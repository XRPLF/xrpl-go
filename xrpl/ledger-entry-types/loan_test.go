package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestLoan(t *testing.T) {
	var s Object = &Loan{
		LedgerEntryType:       LoanEntry,
		Flags:                 0,
		LoanSequence:          1,
		OwnerNode:             "0000000000000000",
		LoanBrokerNode:        "0000000000000000",
		LoanBrokerID:          "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
		Borrower:              "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
		PrincipalOutstanding:  types.XRPLNumber("100000"),
		PeriodicPayment:       types.XRPLNumber("10000"),
		TotalValueOutstanding: types.XRPLNumber("150000"),
		StartDate:             1724871860,
		PaymentInterval:       2592000,
		GracePeriod:           604800,
		NextPaymentDueDate:    1727463860,
		PaymentRemaining:      10,
		PreviousTxnID:         "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		PreviousTxnLgrSeq:     28991004,
	}

	j := `{
	"LedgerEntryType": "Loan",
	"Flags": 0,
	"LoanSequence": 1,
	"OwnerNode": "0000000000000000",
	"LoanBrokerNode": "0000000000000000",
	"LoanBrokerID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
	"Borrower": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
	"PrincipalOutstanding": "100000",
	"PeriodicPayment": "10000",
	"TotalValueOutstanding": "150000",
	"StartDate": 1724871860,
	"PaymentInterval": 2592000,
	"GracePeriod": 604800,
	"NextPaymentDueDate": 1727463860,
	"PaymentRemaining": 10,
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestLoan_EntryType(t *testing.T) {
	s := &Loan{}
	require.Equal(t, s.EntryType(), LoanEntry)
}

func TestLoan_SetLsfLoanDefault(t *testing.T) {
	l := &Loan{}
	l.SetLsfLoanDefault()
	require.Equal(t, l.Flags, LsfLoanDefault)
}

func TestLoan_SetLsfLoanImpaired(t *testing.T) {
	l := &Loan{}
	l.SetLsfLoanImpaired()
	require.Equal(t, l.Flags, LsfLoanImpaired)
}

func TestLoan_SetLsfLoanOverpayment(t *testing.T) {
	l := &Loan{}
	l.SetLsfLoanOverpayment()
	require.Equal(t, l.Flags, LsfLoanOverpayment)
}

func TestLoan_WithOptionalFields(t *testing.T) {
	loanOriginationFee := types.XRPLNumber("1000")
	loanServiceFee := types.XRPLNumber("500")
	latePaymentFee := types.XRPLNumber("2000")
	closePaymentFee := types.XRPLNumber("1500")
	overpaymentFee := types.XRPLNumber("300")
	interestRate := types.InterestRate(5000)
	lateInterestRate := types.InterestRate(1000)
	closeInterestRate := types.InterestRate(2000)
	overpaymentInterestRate := types.InterestRate(500)
	previousPaymentDate := types.PreviousPaymentDate(1724871860)

	var s Object = &Loan{
		LedgerEntryType:         LoanEntry,
		Flags:                   0,
		LoanSequence:            1,
		OwnerNode:               "0000000000000000",
		LoanBrokerNode:          "0000000000000000",
		LoanBrokerID:            "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
		Borrower:                "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
		LoanOriginationFee:      &loanOriginationFee,
		LoanServiceFee:          &loanServiceFee,
		LatePaymentFee:          &latePaymentFee,
		ClosePaymentFee:         &closePaymentFee,
		OverpaymentFee:          &overpaymentFee,
		InterestRate:            &interestRate,
		LateInterestRate:        &lateInterestRate,
		CloseInterestRate:       &closeInterestRate,
		OverpaymentInterestRate: &overpaymentInterestRate,
		StartDate:               1724871860,
		PaymentInterval:         2592000,
		GracePeriod:             604800,
		PreviousPaymentDate:     &previousPaymentDate,
		NextPaymentDueDate:      1727463860,
		PaymentRemaining:        10,
		PrincipalOutstanding:    types.XRPLNumber("100000"),
		PeriodicPayment:         types.XRPLNumber("10000"),
		TotalValueOutstanding:   types.XRPLNumber("150000"),
		PreviousTxnID:           "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
		PreviousTxnLgrSeq:       28991004,
	}

	j := `{
	"LedgerEntryType": "Loan",
	"Flags": 0,
	"LoanSequence": 1,
	"OwnerNode": "0000000000000000",
	"LoanBrokerNode": "0000000000000000",
	"LoanBrokerID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
	"Borrower": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
	"PrincipalOutstanding": "100000",
	"PeriodicPayment": "10000",
	"TotalValueOutstanding": "150000",
	"LoanOriginationFee": "1000",
	"LoanServiceFee": "500",
	"LatePaymentFee": "2000",
	"ClosePaymentFee": "1500",
	"OverpaymentFee": "300",
	"InterestRate": 5000,
	"LateInterestRate": 1000,
	"CloseInterestRate": 2000,
	"OverpaymentInterestRate": 500,
	"StartDate": 1724871860,
	"PaymentInterval": 2592000,
	"GracePeriod": 604800,
	"PreviousPaymentDate": 1724871860,
	"NextPaymentDueDate": 1727463860,
	"PaymentRemaining": 10,
	"PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
	"PreviousTxnLgrSeq": 28991004
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
