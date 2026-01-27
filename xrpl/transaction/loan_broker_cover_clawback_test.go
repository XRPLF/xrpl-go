package transaction

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestLoanBrokerCoverClawback_TxType(t *testing.T) {
	tx := &LoanBrokerCoverClawback{}
	assert.Equal(t, tx.TxType(), LoanBrokerCoverClawbackTx)
}

func TestLoanBrokerCoverClawback_Flatten(t *testing.T) {
	id := types.LoanBrokerID("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430")

	testcases := []struct {
		name     string
		tx       *LoanBrokerCoverClawback
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &LoanBrokerCoverClawback{},
			expected: map[string]interface{}{
				"TransactionType": LoanBrokerCoverClawbackTx.String(),
			},
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				LoanBrokerID: &id,
				Amount:       types.XRPCurrencyAmount(10000),
			},
			expected: map[string]interface{}{
				"TransactionType":    LoanBrokerCoverClawbackTx.String(),
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

func TestLoanBrokerCoverClawback_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanBrokerCoverClawback
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					TransactionType: LoanBrokerCoverClawbackTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - LoanBrokerID invalid",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: func() *types.LoanBrokerID {
					v := types.LoanBrokerID("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43")
					return &v
				}(),
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: errors.New("loanBrokerCoverClawback: LoanBrokerID must be 64 characters hexadecimal string"),
		},
		{
			name: "fail - both LoanBrokerID and Amount missing",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: nil,
				Amount:       nil,
			},
			expected: errors.New("loanBrokerCoverClawback: Either LoanBrokerID or Amount is required"),
		},
		{
			name: "fail - LoanBrokerID empty string and Amount missing",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: func() *types.LoanBrokerID { v := types.LoanBrokerID(""); return &v }(),
				Amount:       nil,
			},
			expected: errors.New("loanBrokerCoverClawback: Either LoanBrokerID or Amount is required"),
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: func() *types.LoanBrokerID {
					v := types.LoanBrokerID("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430")
					return &v
				}(),
				Amount: types.IssuedCurrencyAmount{
					Issuer:   types.Address("rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds"),
					Currency: "FOO",
					Value:    "0",
				},
			},
			expected: nil,
		},
		{
			name: "pass - with Amount only",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: nil,
				Amount: types.IssuedCurrencyAmount{
					Issuer:   types.Address("rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds"),
					Currency: "FOO",
					Value:    "0",
				},
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
