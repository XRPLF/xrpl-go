package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestLoanBrokerSet_TxType(t *testing.T) {
	tx := &LoanBrokerSet{}
	assert.Equal(t, tx.TxType(), LoanBrokerSetTx)
}

func TestLoanBrokerSet_Flatten(t *testing.T) {
	managementFeeRate := types.InterestRate(1000)
	coverRateMinimum := types.InterestRate(5000)

	testcases := []struct {
		name     string
		tx       *LoanBrokerSet
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &LoanBrokerSet{},
			expected: map[string]interface{}{
				"TransactionType": LoanBrokerSetTx.String(),
				"VaultID":         "",
			},
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				VaultID:           "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				ManagementFeeRate: &managementFeeRate,
				CoverRateMinimum:  &coverRateMinimum,
			},
			expected: map[string]interface{}{
				"TransactionType":    LoanBrokerSetTx.String(),
				"Account":            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"VaultID":            "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"ManagementFeeRate":  uint32(1000),
				"CoverRateMinimum":   uint32(5000),
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.tx.Flatten(), testcase.expected)
		})
	}
}

func TestLoanBrokerSet_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanBrokerSet
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					TransactionType: LoanBrokerSetTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - VaultID required",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID: "",
			},
			expected: ErrLoanBrokerSetVaultIDRequired,
		},
		{
			name: "fail - VaultID invalid",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43",
			},
			expected: ErrLoanBrokerSetVaultIDInvalid,
		},
		{
			name: "fail - Data too long",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Data:    func() *types.Data { v := types.Data("A" + strings.Repeat("B", 512)); return &v }(),
			},
			expected: ErrLoanBrokerSetDataInvalid,
		},
		{
			name: "fail - ManagementFeeRate too high",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:           "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				ManagementFeeRate: func() *types.InterestRate { v := types.InterestRate(10001); return &v }(),
			},
			expected: ErrLoanBrokerSetManagementFeeRateInvalid,
		},
		{
			name: "fail - CoverRateMinimum too high",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:          "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				CoverRateMinimum: func() *types.InterestRate { v := types.InterestRate(100001); return &v }(),
			},
			expected: ErrLoanBrokerSetCoverRateMinimumInvalid,
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:              "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				ManagementFeeRate:    func() *types.InterestRate { v := types.InterestRate(1000); return &v }(),
				CoverRateMinimum:     func() *types.InterestRate { v := types.InterestRate(5000); return &v }(),
				CoverRateLiquidation: func() *types.InterestRate { v := types.InterestRate(1); return &v }(),
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
