package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestVaultSet_TxType(t *testing.T) {
	tx := &VaultSet{}
	assert.Equal(t, VaultSetTx, tx.TxType())
}

func TestVaultSet_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultSet
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &VaultSet{},
			expected: FlatTransaction{
				"TransactionType": VaultSetTx.String(),
				"VaultID":         "",
			},
		},
		{
			name: "pass - complete",
			tx: &VaultSet{
				BaseTx: BaseTx{
					Account:            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
			},
			expected: FlatTransaction{
				"TransactionType":    VaultSetTx.String(),
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

func TestVaultSet_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultSet
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &VaultSet{
				BaseTx: BaseTx{
					TransactionType: VaultSetTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - VaultID required",
			tx: &VaultSet{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultSetTx,
				},
				VaultID: types.Hash256(""),
			},
			expected: ErrVaultSetVaultIDRequired,
		},
		{
			name: "fail - VaultID invalid",
			tx: &VaultSet{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultSetTx,
				},
				VaultID: types.Hash256("INVALIDID"),
			},
			expected: ErrVaultSetVaultIDInvalid,
		},
		{
			name: "fail - Data not hex",
			tx: &VaultSet{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultSetTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Data:    func() *types.Data { v := types.Data("zznothex"); return &v }(),
			},
			expected: ErrVaultSetDataInvalid,
		},
		{
			name: "fail - Data too large (> 256 bytes)",
			tx: &VaultSet{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultSetTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Data:    func() *types.Data { v := types.Data(strings.Repeat("AB", 257)); return &v }(),
			},
			expected: ErrVaultSetDataInvalid,
		},
		{
			name: "fail - AssetsMaximum invalid",
			tx: &VaultSet{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultSetTx,
				},
				VaultID:       types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				AssetsMaximum: func() *types.XRPLNumber { v := types.XRPLNumber("notanumber"); return &v }(),
			},
			expected: ErrVaultSetAssetsMaximumInvalid,
		},
		{
			name: "pass - complete",
			tx: &VaultSet{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultSetTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
			},
			expected: nil,
		},
		{
			name: "fail - DomainID invalid (too short)",
			tx: &VaultSet{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultSetTx,
				},
				VaultID:  types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				DomainID: func() *string { v := "TOOSHORT"; return &v }(),
			},
			expected: ErrVaultSetDomainIDInvalid,
		},
		{
			name: "pass - with valid DomainID",
			tx: &VaultSet{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultSetTx,
				},
				VaultID:  types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				DomainID: func() *string { v := "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"; return &v }(),
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
				assert.NoError(t, err)
			}
		})
	}
}
