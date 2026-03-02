package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestVaultDeposit_TxType(t *testing.T) {
	tx := &VaultDeposit{}
	assert.Equal(t, tx.TxType(), VaultDepositTx)
}

func TestVaultDeposit_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultDeposit
		expected map[string]interface{}
	}{
		{
			name: "pass - empty",
			tx:   &VaultDeposit{},
			expected: map[string]interface{}{
				"TransactionType": VaultDepositTx.String(),
				"VaultID":         "",
			},
		},
		{
			name: "pass - complete",
			tx: &VaultDeposit{
				BaseTx: BaseTx{
					Account:            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:  types.XRPCurrencyAmount(10000),
			},
			expected: map[string]interface{}{
				"TransactionType":    VaultDepositTx.String(),
				"Account":            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"VaultID":            "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
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

func TestVaultDeposit_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultDeposit
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &VaultDeposit{
				BaseTx: BaseTx{
					TransactionType: VaultDepositTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - VaultID required",
			tx: &VaultDeposit{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultDepositTx,
				},
				VaultID: "",
				Amount:  types.XRPCurrencyAmount(10000),
			},
			expected: ErrVaultDepositVaultIDRequired,
		},
		{
			name: "fail - VaultID invalid",
			tx: &VaultDeposit{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultDepositTx,
				},
				VaultID: "INVALIDID",
				Amount:  types.XRPCurrencyAmount(10000),
			},
			expected: ErrVaultDepositVaultIDInvalid,
		},
		{
			name: "fail - Amount required",
			tx: &VaultDeposit{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultDepositTx,
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:  nil,
			},
			expected: ErrVaultDepositAmountRequired,
		},
		{
			name: "pass - complete",
			tx: &VaultDeposit{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultDepositTx,
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				Amount:  types.XRPCurrencyAmount(10000),
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
