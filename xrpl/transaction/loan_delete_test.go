package transaction

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoanDelete_TxType(t *testing.T) {
	tx := &LoanDelete{}
	assert.Equal(t, tx.TxType(), LoanDeleteTx)
}

func TestLoanDelete_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanDelete
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &LoanDelete{},
			expected: map[string]interface{}{
				"TransactionType": LoanDeleteTx.String(),
				"LoanID":          "",
			},
		},
		{
			name: "pass - complete",
			tx: &LoanDelete{
				BaseTx: BaseTx{
					Account:            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			expected: map[string]interface{}{
				"TransactionType":    LoanDeleteTx.String(),
				"Account":            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
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

func TestLoanDelete_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanDelete
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &LoanDelete{
				BaseTx: BaseTx{
					TransactionType: LoanDeleteTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - LoanID required",
			tx: &LoanDelete{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanDeleteTx,
				},
				LoanID: "",
			},
			expected: errors.New("loanDelete: LoanID is required"),
		},
		{
			name: "fail - LoanID invalid length",
			tx: &LoanDelete{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanDeleteTx,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
			},
			expected: errors.New("loanDelete: LoanID must be 64 characters hexadecimal string"),
		},
		{
			name: "fail - LoanID invalid hex",
			tx: &LoanDelete{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanDeleteTx,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430G",
			},
			expected: errors.New("loanDelete: LoanID must be 64 characters hexadecimal string"),
		},
		{
			name: "pass - complete",
			tx: &LoanDelete{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanDeleteTx,
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
