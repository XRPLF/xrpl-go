package transaction

// LoanManageFlags represents flags for LoanManage transactions.
const (
	// tfLoanDefault indicates that the Loan should be defaulted.
	tfLoanDefault uint32 = 0x00010000
	// tfLoanImpair indicates that the Loan should be impaired.
	tfLoanImpair uint32 = 0x00020000
	// tfLoanUnimpair indicates that the Loan should be un-impaired.
	tfLoanUnimpair uint32 = 0x00040000
)

// LoanManage modifies an existing Loan object.
//
// ```json
//
//	{
//	  "TransactionType": "LoanManage",
//	  "Account": "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
//	  "LoanID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Flags": 65536
//	}
//
// ```
type LoanManage struct {
	BaseTx
	// The ID of the Loan object to be updated.
	LoanID string
}

// TxType returns the TxType for LoanManage transactions.
func (tx *LoanManage) TxType() TxType {
	return LoanManageTx
}

// SetLoanDefaultFlag sets the tfLoanDefault flag, indicating that the Loan should be defaulted.
func (tx *LoanManage) SetLoanDefaultFlag() {
	tx.Flags |= tfLoanDefault
}

// SetLoanImpairFlag sets the tfLoanImpair flag, indicating that the Loan should be impaired.
func (tx *LoanManage) SetLoanImpairFlag() {
	tx.Flags |= tfLoanImpair
}

// SetLoanUnimpairFlag sets the tfLoanUnimpair flag, indicating that the Loan should be un-impaired.
func (tx *LoanManage) SetLoanUnimpairFlag() {
	tx.Flags |= tfLoanUnimpair
}

// Flatten returns a map representation of the LoanManage transaction for JSON-RPC submission.
func (tx *LoanManage) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Account != "" {
		flattened["Account"] = tx.Account.String()
	}

	flattened["LoanID"] = tx.LoanID

	return flattened
}

// Validate checks LoanManage transaction fields and returns false with an error if invalid.
func (tx *LoanManage) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.LoanID == "" {
		return false, ErrLoanManageLoanIDRequired
	}

	if !IsLedgerEntryID(tx.LoanID) {
		return false, ErrLoanManageLoanIDInvalid
	}

	// Check that tfLoanImpair and tfLoanUnimpair are not both set
	if (tx.Flags&tfLoanImpair) != 0 && (tx.Flags&tfLoanUnimpair) != 0 {
		return false, ErrLoanManageFlagsConflict
	}

	return true, nil
}
