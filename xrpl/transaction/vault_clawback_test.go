package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestVaultClawback_TxType(t *testing.T) {
	tx := &VaultClawback{}
	assert.Equal(t, tx.TxType(), VaultClawbackTx)
}

func TestVaultClawback_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultClawback
		expected FlatTransaction
	}{
		{
			name: "pass - empty",
			tx:   &VaultClawback{},
			expected: FlatTransaction{
				"TransactionType": VaultClawbackTx.String(),
				"VaultID":         "",
				"Holder":          "",
			},
		},
		{
			name: "pass - complete",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Amount:  types.XRPCurrencyAmount(10000),
			},
			expected: FlatTransaction{
				"TransactionType":    VaultClawbackTx.String(),
				"Account":            "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"VaultID":            "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Holder":             "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
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

func TestVaultClawback_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *VaultClawback
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					TransactionType: VaultClawbackTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - VaultID required",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256(""),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
			},
			expected: ErrVaultClawbackVaultIDRequired,
		},
		{
			name: "fail - VaultID invalid",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("INVALIDID"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
			},
			expected: ErrVaultClawbackVaultIDInvalid,
		},
		{
			name: "fail - Holder required",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "",
			},
			expected: ErrVaultClawbackHolderRequired,
		},
		{
			name: "fail - Holder invalid",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "INVALID",
			},
			expected: ErrVaultClawbackHolderInvalid,
		},
		{
			name: "pass - complete without amount",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
			},
			expected: nil,
		},
		{
			name: "fail - Amount cannot be XRP",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Amount:  types.XRPCurrencyAmount(10000),
			},
			expected: ErrVaultClawbackAmountInvalidType,
		},
		{
			name: "pass - complete with MPT amount",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF",
					Value:         "1000",
				},
			},
			expected: nil,
		},
		{
			name: "fail - invalid MPT amount with non-hex issuance ID",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "not-hex",
					Value:         "1000",
				},
			},
			expected: ErrInvalidMPTIssuanceID,
		},
		{
			name: "fail - invalid MPT amount with missing value",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF",
				},
			},
			expected: ErrInvalidMPTValue,
		},
		{
			name: "fail - invalid MPT amount with negative value",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF",
					Value:         "-5",
				},
			},
			expected: ErrInvalidMPTValue,
		},
		{
			name: "fail - invalid IOU amount with missing issuer",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Amount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Value:    "1234",
				},
			},
			expected: ErrInvalidTokenFields,
		},
		{
			name: "pass - complete with IOU amount",
			tx: &VaultClawback{
				BaseTx: BaseTx{
					Account:         "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
					TransactionType: VaultClawbackTx,
				},
				VaultID: types.Hash256("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"),
				Holder:  "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
				Amount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Issuer:   "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb",
					Value:    "1234",
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
