package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// LoanFlags represents flags for Loan ledger entries.
const (
	// lsfLoanDefault indicates that the Loan is defaulted.
	lsfLoanDefault uint32 = 0x00010000
	// lsfLoanImpaired indicates that the Loan is impaired.
	lsfLoanImpaired uint32 = 0x00020000
	// lsfLoanOverpayment indicates that the Loan supports overpayments.
	lsfLoanOverpayment uint32 = 0x00040000
)

// Loan represents a Loan ledger entry that captures various Loan terms on-chain.
// It is an agreement between the Borrower and the loan issuer.
//
// ```json
//
//	{
//	  "LedgerEntryType": "Loan",
//	  "Flags": 0,
//	  "LoanSequence": 1,
//	  "OwnerNode": "0000000000000000",
//	  "LoanBrokerNode": "0000000000000000",
//	  "LoanBrokerID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Borrower": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
//	  "PrincipalOutstanding": "100000",
//	  "PeriodicPayment": "10000",
//	  "TotalValueOutstanding": "150000",
//	  "StartDate": 1724871860,
//	  "PaymentInterval": 2592000,
//	  "GracePeriod": 604800,
//	  "NextPaymentDueDate": 1727463860,
//	  "PaymentRemaining": 10,
//	  "PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
//	  "PreviousTxnLgrSeq": 28991004
//	}
//
// ```
type Loan struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The value "Loan", mapped to the string Loan, indicates that this object is a Loan object.
	LedgerEntryType EntryType
	// Ledger object flags.
	Flags uint32
	// The sequence number of the Loan.
	LoanSequence uint32
	// Identifies the page where this item is referenced in the Borrower owner's directory.
	OwnerNode string
	// Identifies the page where this item is referenced in the LoanBrokers owner's directory.
	LoanBrokerNode string
	// The ID of the LoanBroker associated with this Loan Instance.
	LoanBrokerID types.Hash256
	// The address of the account that is the borrower.
	Borrower types.Address
	// The principal amount requested by the Borrower.
	PrincipalOutstanding types.XRPLNumber
	// The calculated periodic payment amount for each payment interval.
	PeriodicPayment types.XRPLNumber
	// The total outstanding value of the Loan, including all fees and interest.
	TotalValueOutstanding types.XRPLNumber
	// A nominal fee amount paid to the LoanBroker.Owner when the Loan is created.
	LoanOriginationFee *types.XRPLNumber `json:",omitempty"`
	// A nominal funds amount paid to the LoanBroker.Owner with every Loan payment.
	LoanServiceFee *types.XRPLNumber `json:",omitempty"`
	// A nominal funds amount paid to the LoanBroker.Owner when a payment is late.
	LatePaymentFee *types.XRPLNumber `json:",omitempty"`
	// A nominal funds amount paid to the LoanBroker.Owner when a full payment is made.
	ClosePaymentFee *types.XRPLNumber `json:",omitempty"`
	// A fee charged on over-payments in 1/10th basis points. Valid values are between 0 and 100000 inclusive. (0 - 100%)
	OverpaymentFee *types.XRPLNumber `json:",omitempty"`
	// Annualized interest rate of the Loan in 1/10th basis points.
	InterestRate *types.InterestRate `json:",omitempty"`
	// A premium is added to the interest rate for late payments in 1/10th basis points.
	// Valid values are between 0 and 100000 inclusive. (0 - 100%)
	LateInterestRate *types.InterestRate `json:",omitempty"`
	// An interest rate charged for repaying the Loan early in 1/10th basis points.
	// Valid values are between 0 and 100000 inclusive. (0 - 100%)
	CloseInterestRate *types.InterestRate `json:",omitempty"`
	// An interest rate charged on over-payments in 1/10th basis points. Valid values are between 0 and 100000 inclusive. (0 - 100%)
	OverpaymentInterestRate *types.InterestRate `json:",omitempty"`
	// The timestamp of when the Loan started Ripple Epoch.
	StartDate uint32
	// Number of seconds between Loan payments.
	PaymentInterval uint32
	// The number of seconds after the Payment Due Date that the Loan can be Defaulted.
	GracePeriod uint32
	// The timestamp of when the previous payment was made in Ripple Epoch.
	PreviousPaymentDate *types.PreviousPaymentDate `json:",omitempty"`
	// The timestamp of when the next payment is due in Ripple Epoch.
	NextPaymentDueDate uint32
	// The number of payments remaining on the Loan.
	PaymentRemaining uint32
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this entry.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the ledger entry type for Loan.
func (*Loan) EntryType() EntryType {
	return LoanEntry
}

// SetLsfLoanDefault sets the lsfLoanDefault flag, indicating that the Loan is defaulted.
func (l *Loan) SetLsfLoanDefault() {
	l.Flags |= lsfLoanDefault
}

// SetLsfLoanImpaired sets the lsfLoanImpaired flag, indicating that the Loan is impaired.
func (l *Loan) SetLsfLoanImpaired() {
	l.Flags |= lsfLoanImpaired
}

// SetLsfLoanOverpayment sets the lsfLoanOverpayment flag, indicating that the Loan supports overpayments.
func (l *Loan) SetLsfLoanOverpayment() {
	l.Flags |= lsfLoanOverpayment
}
