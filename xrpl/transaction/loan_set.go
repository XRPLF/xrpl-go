package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// LoanSetMaxDataLength is the maximum length in characters for the Data field.
	LoanSetMaxDataLength = 512
	// LoanSetMaxOverPaymentFeeRate is the maximum value for OverpaymentFee (100000 = 100%).
	LoanSetMaxOverPaymentFeeRate = 100_000
	// LoanSetMaxInterestRate is the maximum value for InterestRate (100000 = 100%).
	LoanSetMaxInterestRate = 100_000
	// LoanSetMaxLateInterestRate is the maximum value for LateInterestRate (100000 = 100%).
	LoanSetMaxLateInterestRate = 100_000
	// LoanSetMaxCloseInterestRate is the maximum value for CloseInterestRate (100000 = 100%).
	LoanSetMaxCloseInterestRate = 100_000
	// LoanSetMaxOverPaymentInterestRate is the maximum value for OverpaymentInterestRate (100000 = 100%).
	LoanSetMaxOverPaymentInterestRate = 100_000
	// LoanSetMinPaymentInterval is the minimum value for PaymentInterval in seconds.
	LoanSetMinPaymentInterval = 60
)

// LoanSetFlags represents flags for LoanSet transactions.
const (
	// tfLoanOverpayment indicates that the loan supports over payments.
	tfLoanOverpayment uint32 = 0x00010000
)

// CounterpartySignature represents the signature of the counterparty over the transaction.
type CounterpartySignature struct {
	// The Public Key to be used to verify the validity of the signature.
	SigningPubKey string `json:",omitempty"`
	// The signature of over all signing fields.
	TxnSignature string `json:",omitempty"`
	// An array of transaction signatures from the Counterparty signers to indicate their approval of this transaction.
	Signers []types.Signer `json:",omitempty"`
}

// Flatten returns a map representation of the CounterpartySignature for JSON-RPC submission.
func (cs *CounterpartySignature) Flatten() map[string]interface{} {
	flattened := make(map[string]interface{})
	if cs.SigningPubKey != "" {
		flattened["SigningPubKey"] = cs.SigningPubKey
	}
	if cs.TxnSignature != "" {
		flattened["TxnSignature"] = cs.TxnSignature
	}
	if len(cs.Signers) > 0 {
		flattenedSigners := make([]map[string]interface{}, len(cs.Signers))
		for i, signer := range cs.Signers {
			flattenedSigners[i] = signer.Flatten()
		}
		flattened["Signers"] = flattenedSigners
	}
	return flattened
}

// LoanSet creates a new Loan object.
//
// ```json
//
//	{
//	  "TransactionType": "LoanSet",
//	  "Account": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
//	  "LoanBrokerID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "PrincipalRequested": "100000"
//	}
//
// ```
type LoanSet struct {
	BaseTx
	// The Loan Broker ID associated with the loan.
	LoanBrokerID string
	// The principal amount requested by the Borrower.
	PrincipalRequested types.XRPLNumber
	// The signature of the counterparty over the transaction.
	CounterpartySignature *CounterpartySignature `json:",omitempty"`
	// The address of the counterparty of the Loan.
	Counterparty *types.Address `json:",omitempty"`
	// Arbitrary metadata in hex format. The field is limited to 512 characters.
	Data *types.Data `json:",omitempty"`
	// A nominal funds amount paid to the LoanBroker.Owner when the Loan is created.
	LoanOriginationFee *types.XRPLNumber `json:",omitempty"`
	// A nominal amount paid to the LoanBroker.Owner with every Loan payment.
	LoanServiceFee *types.XRPLNumber `json:",omitempty"`
	// A nominal funds amount paid to the LoanBroker.Owner when a payment is late.
	LatePaymentFee *types.XRPLNumber `json:",omitempty"`
	// A nominal funds amount paid to the LoanBroker.Owner when an early full repayment is made.
	ClosePaymentFee *types.XRPLNumber `json:",omitempty"`
	// A fee charged on overpayments in 1/10th basis points. Valid values are between 0 and 100000 inclusive. (0 - 100%)
	OverpaymentFee *uint32 `json:",omitempty"`
	// Annualized interest rate of the Loan in 1/10th basis points. Valid values are between 0 and 100000 inclusive. (0 - 100%)
	InterestRate *types.InterestRate `json:",omitempty"`
	// A premium added to the interest rate for late payments in 1/10th basis points.
	// Valid values are between 0 and 100000 inclusive. (0 - 100%)
	LateInterestRate *types.InterestRate `json:",omitempty"`
	// A Fee Rate charged for repaying the Loan early in 1/10th basis points.
	// Valid values are between 0 and 100000 inclusive. (0 - 100%)
	CloseInterestRate *types.InterestRate `json:",omitempty"`
	// An interest rate charged on over payments in 1/10th basis points. Valid values are between 0 and 100000 inclusive. (0 - 100%)
	OverpaymentInterestRate *types.InterestRate `json:",omitempty"`
	// The total number of payments to be made against the Loan.
	PaymentTotal *types.PaymentTotal `json:",omitempty"`
	// Number of seconds between Loan payments.
	PaymentInterval *types.PaymentInterval `json:",omitempty"`
	// The number of seconds after the Loan's Payment Due Date can be Defaulted.
	GracePeriod *types.GracePeriod `json:",omitempty"`
}

// TxType returns the TxType for LoanSet transactions.
func (tx *LoanSet) TxType() TxType {
	return LoanSetTx
}

// SetLoanOverpaymentFlag sets the tfLoanOverpayment flag, indicating that the loan supports over payments.
func (tx *LoanSet) SetLoanOverpaymentFlag() {
	tx.Flags |= tfLoanOverpayment
}

// Flatten returns a map representation of the LoanSet transaction for JSON-RPC submission.
func (tx *LoanSet) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	flattened["LoanBrokerID"] = tx.LoanBrokerID
	flattened["PrincipalRequested"] = tx.PrincipalRequested.String()

	if tx.CounterpartySignature != nil {
		flattened["CounterpartySignature"] = tx.CounterpartySignature.Flatten()
	}

	if tx.Counterparty != nil {
		flattened["Counterparty"] = tx.Counterparty.String()
	}

	if tx.Data != nil && *tx.Data != "" {
		flattened["Data"] = string(*tx.Data)
	}

	if tx.LoanOriginationFee != nil && *tx.LoanOriginationFee != "" {
		flattened["LoanOriginationFee"] = tx.LoanOriginationFee.String()
	}

	if tx.LoanServiceFee != nil && *tx.LoanServiceFee != "" {
		flattened["LoanServiceFee"] = tx.LoanServiceFee.String()
	}

	if tx.LatePaymentFee != nil && *tx.LatePaymentFee != "" {
		flattened["LatePaymentFee"] = tx.LatePaymentFee.String()
	}

	if tx.ClosePaymentFee != nil && *tx.ClosePaymentFee != "" {
		flattened["ClosePaymentFee"] = tx.ClosePaymentFee.String()
	}

	if tx.OverpaymentFee != nil && *tx.OverpaymentFee != 0 {
		flattened["OverpaymentFee"] = *tx.OverpaymentFee
	}

	if tx.InterestRate != nil && *tx.InterestRate != 0 {
		flattened["InterestRate"] = uint32(*tx.InterestRate)
	}

	if tx.LateInterestRate != nil && *tx.LateInterestRate != 0 {
		flattened["LateInterestRate"] = uint32(*tx.LateInterestRate)
	}

	if tx.CloseInterestRate != nil && *tx.CloseInterestRate != 0 {
		flattened["CloseInterestRate"] = uint32(*tx.CloseInterestRate)
	}

	if tx.OverpaymentInterestRate != nil && *tx.OverpaymentInterestRate != 0 {
		flattened["OverpaymentInterestRate"] = uint32(*tx.OverpaymentInterestRate)
	}

	if tx.PaymentTotal != nil && *tx.PaymentTotal != 0 {
		flattened["PaymentTotal"] = uint32(*tx.PaymentTotal)
	}

	if tx.PaymentInterval != nil && *tx.PaymentInterval != 0 {
		flattened["PaymentInterval"] = uint32(*tx.PaymentInterval)
	}

	if tx.GracePeriod != nil && *tx.GracePeriod != 0 {
		flattened["GracePeriod"] = uint32(*tx.GracePeriod)
	}

	return flattened
}

// Validate checks LoanSet transaction fields and returns false with an error if invalid.
func (tx *LoanSet) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.LoanBrokerID == "" {
		return false, ErrLoanSetLoanBrokerIDRequired
	}

	if !IsLedgerEntryID(tx.LoanBrokerID) {
		return false, ErrLoanSetLoanBrokerIDInvalid
	}

	if tx.PrincipalRequested == "" {
		return false, ErrLoanSetPrincipalRequestedRequired
	}

	if !typecheck.IsXRPLNumber(tx.PrincipalRequested.String()) {
		return false, ErrLoanSetPrincipalRequestedInvalid
	}

	if tx.Data != nil && *tx.Data != "" {
		if !ValidateHexMetadata(tx.Data.Value(), LoanSetMaxDataLength) {
			return false, ErrLoanSetDataInvalid
		}
	}

	if tx.Counterparty != nil {
		if !addresscodec.IsValidAddress(tx.Counterparty.String()) {
			return false, ErrInvalidAccount
		}
	}

	if tx.OverpaymentFee != nil && *tx.OverpaymentFee > LoanSetMaxOverPaymentFeeRate {
		return false, ErrLoanSetOverpaymentFeeInvalid
	}

	if tx.InterestRate != nil && *tx.InterestRate > LoanSetMaxInterestRate {
		return false, ErrLoanSetInterestRateInvalid
	}

	if tx.LateInterestRate != nil && *tx.LateInterestRate > LoanSetMaxLateInterestRate {
		return false, ErrLoanSetLateInterestRateInvalid
	}

	if tx.CloseInterestRate != nil && *tx.CloseInterestRate > LoanSetMaxCloseInterestRate {
		return false, ErrLoanSetCloseInterestRateInvalid
	}

	if tx.OverpaymentInterestRate != nil && *tx.OverpaymentInterestRate > LoanSetMaxOverPaymentInterestRate {
		return false, ErrLoanSetOverpaymentInterestRateInvalid
	}

	if tx.PaymentInterval != nil && *tx.PaymentInterval != 0 && *tx.PaymentInterval < LoanSetMinPaymentInterval {
		return false, ErrLoanSetPaymentIntervalInvalid
	}

	if tx.PaymentInterval != nil && tx.GracePeriod != nil && *tx.PaymentInterval != 0 && *tx.GracePeriod != 0 && tx.GracePeriod.Value() > tx.PaymentInterval.Value() {
		return false, ErrLoanSetGracePeriodInvalid
	}

	// Validate optional XRPLNumber fields
	if tx.LoanOriginationFee != nil && *tx.LoanOriginationFee != "" && !typecheck.IsXRPLNumber(tx.LoanOriginationFee.String()) {
		return false, ErrLoanSetLoanOriginationFeeInvalid
	}

	if tx.LoanServiceFee != nil && *tx.LoanServiceFee != "" && !typecheck.IsXRPLNumber(tx.LoanServiceFee.String()) {
		return false, ErrLoanSetLoanServiceFeeInvalid
	}

	if tx.LatePaymentFee != nil && *tx.LatePaymentFee != "" && !typecheck.IsXRPLNumber(tx.LatePaymentFee.String()) {
		return false, ErrLoanSetLatePaymentFeeInvalid
	}

	if tx.ClosePaymentFee != nil && *tx.ClosePaymentFee != "" && !typecheck.IsXRPLNumber(tx.ClosePaymentFee.String()) {
		return false, ErrLoanSetClosePaymentFeeInvalid
	}

	return true, nil
}
