package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVaultDelete_TxType(t *testing.T) {
	tx := &VaultDelete{}
	assert.Equal(t, tx.TxType(), VaultDeleteTx)
}

func TestVaultDelete_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultDelete
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &VaultDelete{},
			expected: FlatTransaction{
				"TransactionType": VaultDeleteTx.String(),
				"VaultID":         "",
			},
		},
		{
			name: "pass - complete",
			tx: &VaultDelete{
				BaseTx: BaseTx{
					Account:            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
			expected: FlatTransaction{
				"TransactionType":    VaultDeleteTx.String(),
				"Account":            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"VaultID":            "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.tx.Flatten(), testcase.expected)
		})
	}
}

func TestVaultDelete_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultDelete
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &VaultDelete{
				BaseTx: BaseTx{
					TransactionType: VaultDeleteTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - VaultID required",
			tx: &VaultDelete{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultDeleteTx,
				},
				VaultID: "",
			},
			expected: ErrVaultDeleteVaultIDRequired,
		},
		{
			name: "fail - VaultID invalid",
			tx: &VaultDelete{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultDeleteTx,
				},
				VaultID: "INVALIDID",
			},
			expected: ErrVaultDeleteVaultIDInvalid,
		},
		{
			name: "pass - complete",
			tx: &VaultDelete{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultDeleteTx,
				},
				VaultID: "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
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
