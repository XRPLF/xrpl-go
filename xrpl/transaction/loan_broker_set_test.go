package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoanBrokerSet_TxType(t *testing.T) {
	tx := &LoanBrokerSet{}
	assert.Equal(t, LoanBrokerSetTx, tx.TxType())
}

func TestLoanBrokerSet_Flatten(t *testing.T) {
	managementFeeRate := types.InterestRate(1000)
	coverRateMinimum := types.InterestRate(5000)

	testcases := []struct {
		name     string
		tx       *LoanBrokerSet
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &LoanBrokerSet{},
			expected: FlatTransaction{
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
			expected: FlatTransaction{
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
			assert.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestLoanBrokerSet_Validate(t *testing.T) {
	id := types.LoanBrokerID("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430")
	nonzero := types.InterestRate(1)
	zero := types.InterestRate(0)
	testcases := []struct {
		name     string
		tx       *LoanBrokerSet
		expected error
	}{
		{
			name: "fail - update ManagementFeeRate",
			tx: &LoanBrokerSet{
				BaseTx:  BaseTx{Account: testAddrAccount, TransactionType: LoanBrokerSetTx},
				VaultID: id.Value(), LoanBrokerID: &id, ManagementFeeRate: &nonzero,
			},
			expected: ErrTransactionInvalidField{Type: "LoanBrokerSet", Field: "ManagementFeeRate"},
		},
		{
			name: "fail - update cover rates",
			tx: &LoanBrokerSet{
				BaseTx:  BaseTx{Account: testAddrAccount, TransactionType: LoanBrokerSetTx},
				VaultID: id.Value(), LoanBrokerID: &id, CoverRateMinimum: &nonzero, CoverRateLiquidation: &nonzero,
			},
			expected: ErrTransactionInvalidField{Type: "LoanBrokerSet", Field: "CoverRateMinimum"},
		},
		{
			name: "fail - update CoverRateLiquidation alone",
			tx: &LoanBrokerSet{
				BaseTx:  BaseTx{Account: testAddrAccount, TransactionType: LoanBrokerSetTx},
				VaultID: id.Value(), LoanBrokerID: &id, CoverRateLiquidation: &nonzero,
			},
			expected: ErrTransactionInvalidField{Type: "LoanBrokerSet", Field: "CoverRateLiquidation"},
		},
		{
			name: "pass - update with omitted zero rates",
			tx: &LoanBrokerSet{
				BaseTx:  BaseTx{Account: testAddrAccount, TransactionType: LoanBrokerSetTx},
				VaultID: id.Value(), LoanBrokerID: &id, ManagementFeeRate: &zero, CoverRateMinimum: &zero, CoverRateLiquidation: &zero,
			},
		},
		{
			name: "pass - create with rates",
			tx: &LoanBrokerSet{
				BaseTx:  BaseTx{Account: testAddrAccount, TransactionType: LoanBrokerSetTx},
				VaultID: id.Value(), ManagementFeeRate: &nonzero, CoverRateMinimum: &nonzero, CoverRateLiquidation: &nonzero,
			},
		},
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
			name: "fail - zero VaultID",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID: strings.Repeat("0", 64),
			},
			expected: ErrLoanBrokerSetVaultIDInvalid,
		},
		{
			name: "fail - zero LoanBrokerID",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:      "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				LoanBrokerID: func() *types.LoanBrokerID { v := types.LoanBrokerID(strings.Repeat("0", 64)); return &v }(),
			},
			expected: ErrLoanBrokerSetLoanBrokerIDInvalid,
		},
		{
			name: "fail - DebtMaximum exponent too large to encode",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:     "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				DebtMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("1e40000"); return &v }(),
			},
			expected: ErrLoanBrokerSetDebtMaximumInvalid,
		},
		{
			name: "fail - DebtMaximum without integer digits",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:     "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				DebtMaximum: func() *types.XRPLNumber { v := types.XRPLNumber(".5"); return &v }(),
			},
			expected: ErrLoanBrokerSetDebtMaximumInvalid,
		},
		{
			name: "fail - DebtMaximum negative",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:     "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				DebtMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("-1"); return &v }(),
			},
			expected: ErrLoanBrokerSetDebtMaximumNegative,
		},
		{
			name: "fail - DebtMaximum above float64 range",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:     "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				DebtMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("1e400"); return &v }(),
			},
			expected: ErrLoanBrokerSetDebtMaximumInvalid,
		},
		{
			name: "fail - DebtMaximum above the MPT maximum",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:     "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				DebtMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("9223372036854775808"); return &v }(),
			},
			expected: ErrLoanBrokerSetDebtMaximumInvalid,
		},
		{
			name: "fail - DebtMaximum 1e19",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:     "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				DebtMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("1e19"); return &v }(),
			},
			expected: ErrLoanBrokerSetDebtMaximumInvalid,
		},
		{
			name: "pass - DebtMaximum at the MPT maximum",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:     "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				DebtMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("9223372036854775807"); return &v }(),
			},
		},
		{
			name: "pass - DebtMaximum that encodes to the MPT maximum",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:     "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				DebtMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("9223372036854775807.4"); return &v }(),
			},
		},
		{
			name: "pass - DebtMaximum zero",
			tx: &LoanBrokerSet{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerSetTx,
				},
				VaultID:     "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				DebtMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("0"); return &v }(),
			},
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
				require.Error(t, err)
				assert.Contains(t, err.Error(), testcase.expected.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
