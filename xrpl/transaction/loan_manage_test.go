package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoanManage_TxType(t *testing.T) {
	tx := &LoanManage{}
	assert.Equal(t, tx.TxType(), LoanManageTx)
}

func TestLoanManage_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanManage
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &LoanManage{},
			expected: map[string]interface{}{
				"TransactionType": LoanManageTx.String(),
				"LoanID":          "",
			},
		},
		{
			name: "pass - complete",
			tx: &LoanManage{
				BaseTx: BaseTx{
					Account:            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
					Flags:              uint32(TfLoanDefault),
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			expected: map[string]interface{}{
				"TransactionType":    LoanManageTx.String(),
				"Account":            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"Flags":              uint32(TfLoanDefault),
				"LoanID":             "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.tx.Flatten(), testcase.expected)
		})
	}
}

func TestLoanManage_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanManage
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &LoanManage{
				BaseTx: BaseTx{
					TransactionType: LoanManageTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - LoanID required",
			tx: &LoanManage{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanManageTx,
				},
				LoanID: "",
			},
			expected: ErrLoanManageLoanIDRequired,
		},
		{
			name: "fail - LoanID invalid",
			tx: &LoanManage{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanManageTx,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
			},
			expected: ErrLoanManageLoanIDInvalid,
		},
		{
			name: "fail - TfLoanImpair and TfLoanUnimpair both set",
			tx: &LoanManage{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanManageTx,
					Flags:           uint32(TfLoanImpair | TfLoanUnimpair),
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			expected: ErrLoanManageFlagsConflict,
		},
		{
			name: "pass - complete",
			tx: &LoanManage{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanManageTx,
					Flags:           uint32(TfLoanDefault),
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			expected: nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ok, err := testcase.tx.Validate()
			assert.Equal(t, ok, testcase.expected == nil)
			if testcase.expected != nil {
				assert.Contains(t, err.Error(), testcase.expected.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestLoanManage_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*LoanManage)
		expected uint32
	}{
		{
			name: "pass - SetLoanDefaultFlag",
			setter: func(lm *LoanManage) {
				lm.SetLoanDefaultFlag()
			},
			expected: TfLoanDefault,
		},
		{
			name: "pass - SetLoanImpairFlag",
			setter: func(lm *LoanManage) {
				lm.SetLoanImpairFlag()
			},
			expected: TfLoanImpair,
		},
		{
			name: "pass - SetLoanUnimpairFlag",
			setter: func(lm *LoanManage) {
				lm.SetLoanUnimpairFlag()
			},
			expected: TfLoanUnimpair,
		},
		{
			name: "pass - SetLoanDefaultFlag and SetLoanImpairFlag",
			setter: func(lm *LoanManage) {
				lm.SetLoanDefaultFlag()
				lm.SetLoanImpairFlag()
			},
			expected: TfLoanDefault | TfLoanImpair,
		},
		{
			name: "pass - SetLoanDefaultFlag and SetLoanUnimpairFlag",
			setter: func(lm *LoanManage) {
				lm.SetLoanDefaultFlag()
				lm.SetLoanUnimpairFlag()
			},
			expected: TfLoanDefault | TfLoanUnimpair,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := &LoanManage{}
			tt.setter(lm)
			if lm.Flags != tt.expected {
				t.Errorf("Expected LoanManage Flags to be %d, got %d", tt.expected, lm.Flags)
			}
		})
	}
}
