package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestLoanBrokerCoverDeposit_TxType(t *testing.T) {
	tx := &LoanBrokerCoverDeposit{}
	assert.Equal(t, tx.TxType(), LoanBrokerCoverDepositTx)
}

func TestLoanBrokerCoverDeposit_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanBrokerCoverDeposit
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &LoanBrokerCoverDeposit{},
			expected: map[string]interface{}{
				"TransactionType": LoanBrokerCoverDepositTx.String(),
				"LoanBrokerID":    "",
			},
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerCoverDeposit{
				BaseTx: BaseTx{
					Account:            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:       types.XRPCurrencyAmount(10000),
			},
			expected: map[string]interface{}{
				"TransactionType":    LoanBrokerCoverDepositTx.String(),
				"Account":            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"LoanBrokerID":       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Amount":             "10000",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.tx.Flatten(), testcase.expected)
		})
	}
}

func TestLoanBrokerCoverDeposit_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanBrokerCoverDeposit
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &LoanBrokerCoverDeposit{
				BaseTx: BaseTx{
					TransactionType: LoanBrokerCoverDepositTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - LoanBrokerID required",
			tx: &LoanBrokerCoverDeposit{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverDepositTx,
				},
				LoanBrokerID: "",
				Amount:       types.XRPCurrencyAmount(10000),
			},
			expected: ErrLoanBrokerCoverDepositLoanBrokerIDRequired,
		},
		{
			name: "fail - Amount required",
			tx: &LoanBrokerCoverDeposit{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverDepositTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:       nil,
			},
			expected: ErrLoanBrokerCoverDepositAmountRequired,
		},
		{
			name: "fail - LoanBrokerID invalid",
			tx: &LoanBrokerCoverDeposit{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverDepositTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
				Amount:       types.XRPCurrencyAmount(10000),
			},
			expected: ErrLoanBrokerCoverDepositLoanBrokerIDInvalid,
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerCoverDeposit{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverDepositTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:       types.XRPCurrencyAmount(10000),
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
