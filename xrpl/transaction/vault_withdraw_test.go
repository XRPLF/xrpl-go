package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultWithdraw_TxType(t *testing.T) {
	tx := &VaultWithdraw{}
	assert.Equal(t, VaultWithdrawTx, tx.TxType())
}

func TestVaultWithdraw_Flatten(t *testing.T) {
	dest := types.Address("rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85")
	destTag := uint32(42)

	testcases := []struct {
		name     string
		tx       *VaultWithdraw
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &VaultWithdraw{},
			expected: FlatTransaction{
				"TransactionType": VaultWithdrawTx.String(),
				"VaultID":         "",
			},
		},
		{
			name: "pass - complete",
			tx: &VaultWithdraw{
				BaseTx: BaseTx{
					Account:            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				VaultID:        types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Amount:         types.XRPCurrencyAmount(10000),
				Destination:    &dest,
				DestinationTag: &destTag,
			},
			expected: FlatTransaction{
				"TransactionType":    VaultWithdrawTx.String(),
				"Account":            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"VaultID":            "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Amount":             "10000",
				"Destination":        "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				"DestinationTag":     uint32(42),
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestVaultWithdraw_Validate(t *testing.T) {
	invalidDest := types.Address("INVALID")

	testcases := []struct {
		name     string
		tx       *VaultWithdraw
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &VaultWithdraw{
				BaseTx: BaseTx{
					TransactionType: VaultWithdrawTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - VaultID required",
			tx: &VaultWithdraw{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultWithdrawTx,
				},
				VaultID: types.Hash256(""),
				Amount:  types.XRPCurrencyAmount(10000),
			},
			expected: ErrVaultWithdrawVaultIDRequired,
		},
		{
			name: "fail - VaultID invalid",
			tx: &VaultWithdraw{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultWithdrawTx,
				},
				VaultID: types.Hash256("INVALIDID"),
				Amount:  types.XRPCurrencyAmount(10000),
			},
			expected: ErrVaultWithdrawVaultIDInvalid,
		},
		{
			name: "fail - Amount required",
			tx: &VaultWithdraw{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultWithdrawTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Amount:  nil,
			},
			expected: ErrVaultWithdrawAmountRequired,
		},
		{
			name: "fail - invalid Destination",
			tx: &VaultWithdraw{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultWithdrawTx,
				},
				VaultID:     types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: &invalidDest,
			},
			expected: ErrInvalidDestination,
		},
		{
			name: "pass - complete",
			tx: &VaultWithdraw{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultWithdrawTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Amount:  types.XRPCurrencyAmount(10000),
			},
			expected: nil,
		},
		{
			name: "pass - with valid Destination",
			tx: &VaultWithdraw{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultWithdrawTx,
				},
				VaultID:     types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: func() *types.Address { v := types.Address("rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb"); return &v }(),
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
