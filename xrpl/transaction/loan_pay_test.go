package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestLoanPay_TxType(t *testing.T) {
	tx := &LoanPay{}
	assert.Equal(t, LoanPayTx, tx.TxType())
}

func TestLoanPay_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanPay
		expected map[string]any
	}{
		{
			name: "pass - empty",
			tx:   &LoanPay{},
			expected: map[string]any{
				"TransactionType": LoanPayTx.String(),
				"LoanID":          "",
			},
		},
		{
			name: "pass - complete",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: map[string]any{
				"TransactionType":    LoanPayTx.String(),
				"Account":            "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"LoanID":             "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Amount":             "10000",
			},
		},
		{
			name: "pass - with TfLoanPayOverpayment flag",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account: "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					Flags:   TfLoanPayOverpayment,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: map[string]any{
				"TransactionType": LoanPayTx.String(),
				"Account":         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				"Flags":           TfLoanPayOverpayment,
				"LoanID":          "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Amount":          "10000",
			},
		},
		{
			name: "pass - with TfLoanPayFullPayment flag",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account: "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					Flags:   TfLoanPayFullPayment,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: map[string]any{
				"TransactionType": LoanPayTx.String(),
				"Account":         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				"Flags":           TfLoanPayFullPayment,
				"LoanID":          "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Amount":          "10000",
			},
		},
		{
			name: "pass - with TfLoanPayLatePayment flag",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account: "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					Flags:   TfLoanPayLatePayment,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: map[string]any{
				"TransactionType": LoanPayTx.String(),
				"Account":         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
				"Flags":           TfLoanPayLatePayment,
				"LoanID":          "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Amount":          "10000",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestLoanPay_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanPay
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &LoanPay{
				BaseTx: BaseTx{
					TransactionType: LoanPayTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - LoanID required",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanPayTx,
				},
				LoanID: "",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: ErrLoanPayLoanIDRequired,
		},
		{
			name: "fail - Amount required",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanPayTx,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: nil,
			},
			expected: ErrLoanPayAmountRequired,
		},
		{
			name: "fail - LoanID invalid",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanPayTx,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: ErrLoanPayLoanIDInvalid,
		},
		{
			name: "pass - complete",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanPayTx,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: nil,
		},
		{
			name: "pass - with TfLoanPayOverpayment flag",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanPayTx,
					Flags:           TfLoanPayOverpayment,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: nil,
		},
		{
			name: "pass - with TfLoanPayFullPayment flag",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanPayTx,
					Flags:           TfLoanPayFullPayment,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: nil,
		},
		{
			name: "pass - with TfLoanPayLatePayment flag",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanPayTx,
					Flags:           TfLoanPayLatePayment,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: nil,
		},
		{
			name: "fail - mutually exclusive flags (Overpayment + FullPayment)",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanPayTx,
					Flags:           TfLoanPayOverpayment | TfLoanPayFullPayment,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: ErrLoanPayMutuallyExclusiveFlags,
		},
		{
			name: "fail - mutually exclusive flags (all three set)",
			tx: &LoanPay{
				BaseTx: BaseTx{
					Account:         "rHLLL3Z7uBLK49yZcMaj8FAP7DU12Nw5A5",
					TransactionType: LoanPayTx,
					Flags:           TfLoanPayOverpayment | TfLoanPayFullPayment | TfLoanPayLatePayment,
				},
				LoanID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: ErrLoanPayMutuallyExclusiveFlags,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			ok, err := testcase.tx.Validate()
			assert.Equal(t, ok, testcase.expected == nil)
			if testcase.expected != nil {
				assert.Contains(t, err.Error(), testcase.expected.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
