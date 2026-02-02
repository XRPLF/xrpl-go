package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestLoanBrokerCoverWithdraw_TxType(t *testing.T) {
	tx := &LoanBrokerCoverWithdraw{}
	assert.Equal(t, tx.TxType(), LoanBrokerCoverWithdrawTx)
}

func TestLoanBrokerCoverWithdraw_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanBrokerCoverWithdraw
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &LoanBrokerCoverWithdraw{},
			expected: map[string]interface{}{
				"TransactionType": LoanBrokerCoverWithdrawTx.String(),
				"LoanBrokerID":    "",
			},
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerCoverWithdraw{
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
				"TransactionType":    LoanBrokerCoverWithdrawTx.String(),
				"Account":            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"LoanBrokerID":       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Amount":             "10000",
			},
		},
		{
			name: "pass - with Destination",
			tx: &LoanBrokerCoverWithdraw{
				BaseTx: BaseTx{
					Account:            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:       types.XRPCurrencyAmount(10000),
				Destination:  func() *types.Address { v := types.Address("rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5"); return &v }(),
			},
			expected: map[string]interface{}{
				"TransactionType":    LoanBrokerCoverWithdrawTx.String(),
				"Account":            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"LoanBrokerID":       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Amount":             "10000",
				"Destination":        "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.tx.Flatten(), testcase.expected)
		})
	}
}

func TestLoanBrokerCoverWithdraw_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanBrokerCoverWithdraw
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &LoanBrokerCoverWithdraw{
				BaseTx: BaseTx{
					TransactionType: LoanBrokerCoverWithdrawTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - LoanBrokerID required",
			tx: &LoanBrokerCoverWithdraw{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverWithdrawTx,
				},
				LoanBrokerID: "",
				Amount:       types.XRPCurrencyAmount(10000),
			},
			expected: ErrLoanBrokerCoverWithdrawLoanBrokerIDRequired,
		},
		{
			name: "fail - Amount required",
			tx: &LoanBrokerCoverWithdraw{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverWithdrawTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:       nil,
			},
			expected: ErrLoanBrokerCoverWithdrawAmountRequired,
		},
		{
			name: "fail - LoanBrokerID invalid",
			tx: &LoanBrokerCoverWithdraw{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverWithdrawTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
				Amount:       types.XRPCurrencyAmount(10000),
			},
			expected: ErrLoanBrokerCoverWithdrawLoanBrokerIDInvalid,
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerCoverWithdraw{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverWithdrawTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:       types.XRPCurrencyAmount(10000),
			},
			expected: nil,
		},
		{
			name: "pass - with Destination",
			tx: &LoanBrokerCoverWithdraw{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverWithdrawTx,
				},
				LoanBrokerID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:       types.XRPCurrencyAmount(10000),
				Destination:  func() *types.Address { v := types.Address("rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5"); return &v }(),
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
