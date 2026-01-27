package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoanBrokerDelete_TxType(t *testing.T) {
	tx := &LoanBrokerDelete{}
	assert.Equal(t, tx.TxType(), LoanBrokerDeleteTx)
}

func TestLoanBrokerDelete_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanBrokerDelete
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &LoanBrokerDelete{},
			expected: map[string]interface{}{
				"TransactionType": LoanBrokerDeleteTx.String(),
				"LoanBrokerID":    "",
			},
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerDelete{
				BaseTx: BaseTx{
					Account:            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			expected: map[string]interface{}{
				"TransactionType":    LoanBrokerDeleteTx.String(),
				"Account":            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"LoanBrokerID":       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.tx.Flatten(), testcase.expected)
		})
	}
}

func TestLoanBrokerDelete_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanBrokerDelete
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &LoanBrokerDelete{
				BaseTx: BaseTx{
					TransactionType: LoanBrokerDeleteTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - LoanBrokerID required",
			tx: &LoanBrokerDelete{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerDeleteTx,
				},
				LoanBrokerID: "",
			},
			expected: ErrLoanBrokerDeleteLoanBrokerIDRequired,
		},
		{
			name: "fail - LoanBrokerID invalid",
			tx: &LoanBrokerDelete{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerDeleteTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
			},
			expected: ErrLoanBrokerDeleteLoanBrokerIDInvalid,
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerDelete{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerDeleteTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
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
