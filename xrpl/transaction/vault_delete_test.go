package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultDelete_TxType(t *testing.T) {
	tx := &VaultDelete{}
	assert.Equal(t, VaultDeleteTx, tx.TxType())
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
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
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
			assert.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestVaultDelete_Validate(t *testing.T) {
	const maxMetadataBytes = 256
	testcases := []struct {
		name     string
		tx       *VaultDelete
		expected error
	}{
		{
			name: "valid metadata",
			tx: &VaultDelete{
				BaseTx:   BaseTx{Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es", TransactionType: VaultDeleteTx},
				VaultID:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				MemoData: func() *string { value := "aBcD"; return &value }(),
			},
			expected: nil,
		},
		{
			name: "empty metadata",
			tx: &VaultDelete{
				BaseTx:   BaseTx{Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es", TransactionType: VaultDeleteTx},
				VaultID:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				MemoData: func() *string { value := ""; return &value }(),
			},
			expected: ErrVaultDeleteMemoDataInvalid,
		},
		{
			name: "odd metadata",
			tx: &VaultDelete{
				BaseTx:   BaseTx{Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es", TransactionType: VaultDeleteTx},
				VaultID:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				MemoData: func() *string { value := "ABC"; return &value }(),
			},
			expected: ErrVaultDeleteMemoDataInvalid,
		},
		{
			name: "nonhex metadata",
			tx: &VaultDelete{
				BaseTx:   BaseTx{Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es", TransactionType: VaultDeleteTx},
				VaultID:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				MemoData: func() *string { value := "XX"; return &value }(),
			},
			expected: ErrVaultDeleteMemoDataInvalid,
		},
		{
			name: "maximum metadata",
			tx: &VaultDelete{
				BaseTx:   BaseTx{Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es", TransactionType: VaultDeleteTx},
				VaultID:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				MemoData: func() *string { value := strings.Repeat("AB", maxMetadataBytes); return &value }(),
			},
			expected: nil,
		},
		{
			name: "oversized metadata",
			tx: &VaultDelete{
				BaseTx:   BaseTx{Account: "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es", TransactionType: VaultDeleteTx},
				VaultID:  "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				MemoData: func() *string { value := strings.Repeat("AB", maxMetadataBytes+1); return &value }(),
			},
			expected: ErrVaultDeleteMemoDataInvalid,
		},
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
				VaultID: types.Hash256(""),
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
				VaultID: types.Hash256("INVALIDID"),
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
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
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
				require.NoError(t, err)
			}
		})
	}
}
