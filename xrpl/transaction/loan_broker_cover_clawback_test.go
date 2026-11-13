package transaction

import (
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoanBrokerCoverClawback_TxType(t *testing.T) {
	tx := &LoanBrokerCoverClawback{}
	assert.Equal(t, LoanBrokerCoverClawbackTx, tx.TxType())
}

func TestLoanBrokerCoverClawback_Flatten(t *testing.T) {
	id := types.LoanBrokerID("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430")

	emptyID := types.LoanBrokerID("")
	testcases := []struct {
		name     string
		tx       *LoanBrokerCoverClawback
		expected FlatTransaction
	}{
		{
			name:     "pass - empty LoanBrokerID",
			tx:       &LoanBrokerCoverClawback{LoanBrokerID: &emptyID},
			expected: FlatTransaction{"TransactionType": LoanBrokerCoverClawbackTx.String()},
		},
		{
			name: "pass - empty",
			tx:   &LoanBrokerCoverClawback{},
			expected: FlatTransaction{
				"TransactionType": LoanBrokerCoverClawbackTx.String(),
			},
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					Fee:                1000000,
					Sequence:           1,
					LastLedgerSequence: 3000000,
				},
				LoanBrokerID: &id,
				Amount:       types.XRPCurrencyAmount(10000),
			},
			expected: FlatTransaction{
				"TransactionType":    LoanBrokerCoverClawbackTx.String(),
				"Account":            "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
				"Fee":                "1000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(3000000),
				"LoanBrokerID":       "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
				"Amount":             "10000",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestLoanBrokerCoverClawback_EmptyIDEncodes(t *testing.T) {
	emptyID := types.LoanBrokerID("")
	tx := LoanBrokerCoverClawback{
		BaseTx:       BaseTx{Account: testAddrAccount, TransactionType: LoanBrokerCoverClawbackTx},
		LoanBrokerID: &emptyID,
		Amount:       types.IssuedCurrencyAmount{Currency: "USD", Issuer: testAddrDelegate, Value: "1"},
	}
	ok, err := tx.Validate()
	require.NoError(t, err)
	require.True(t, ok)
	require.NotContains(t, tx.Flatten(), "LoanBrokerID")
	_, err = binarycodec.Encode(tx.Flatten())
	require.NoError(t, err)
}

func TestLoanBrokerCoverClawback_Validate(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *LoanBrokerCoverClawback
		expected error
	}{
		{
			name: "fail - base tx invalid",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					TransactionType: LoanBrokerCoverClawbackTx,
				},
			},
			expected: ErrInvalidAccount,
		},
		{
			name: "fail - LoanBrokerID invalid",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: func() *types.LoanBrokerID {
					v := types.LoanBrokerID("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F43")
					return &v
				}(),
				Amount: types.XRPCurrencyAmount(10000),
			},
			expected: ErrLoanBrokerCoverClawbackLoanBrokerIDInvalid,
		},
		{
			name: "fail - zero LoanBrokerID",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: func() *types.LoanBrokerID {
					v := types.LoanBrokerID("0000000000000000000000000000000000000000000000000000000000000000")
					return &v
				}(),
			},
			expected: ErrLoanBrokerCoverClawbackLoanBrokerIDInvalid,
		},
		{
			name: "fail - Amount issuer is ACCOUNT_ZERO",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				Amount: types.IssuedCurrencyAmount{
					Issuer:   types.Address("rrrrrrrrrrrrrrrrrrrrrhoLvTp"),
					Currency: "FOO",
					Value:    "1",
				},
			},
			expected: ErrInvalidIssuer,
		},
		{
			name: "fail - Amount value is NaN",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				Amount: types.IssuedCurrencyAmount{
					Issuer:   types.Address("rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds"),
					Currency: "FOO",
					Value:    "NaN",
				},
			},
			expected: ErrInvalidTokenValue,
		},
		{
			name: "fail - both LoanBrokerID and Amount missing",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: nil,
				Amount:       nil,
			},
			expected: ErrLoanBrokerCoverClawbackLoanBrokerIDOrAmountRequired,
		},
		{
			name: "fail - LoanBrokerID empty string and Amount missing",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: func() *types.LoanBrokerID { v := types.LoanBrokerID(""); return &v }(),
				Amount:       nil,
			},
			expected: ErrLoanBrokerCoverClawbackLoanBrokerIDOrAmountRequired,
		},
		{
			name: "pass - complete",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: func() *types.LoanBrokerID {
					v := types.LoanBrokerID("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430")
					return &v
				}(),
				Amount: types.IssuedCurrencyAmount{
					Issuer:   types.Address("rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds"),
					Currency: "FOO",
					Value:    "0",
				},
			},
			expected: nil,
		},
		{
			name: "fail - Amount issued by Account without LoanBrokerID",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				Amount: types.IssuedCurrencyAmount{
					Issuer:   types.Address("rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds"),
					Currency: "FOO",
					Value:    "0",
				},
			},
			expected: ErrMissingField{Field: "LoanBrokerID"},
		},
		{
			name: "fail - Amount issued by Account X-address without LoanBrokerID",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				Amount: types.IssuedCurrencyAmount{
					Issuer:   types.Address("XVD5X2NbpcWzGE51Pu7f1z7bBZRYF3VQNLaVSNJkEca7jLR"),
					Currency: "FOO",
					Value:    "1",
				},
			},
			expected: ErrMissingField{Field: "LoanBrokerID"},
		},
		{
			name: "fail - MPT Amount without LoanBrokerID",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: func() *types.LoanBrokerID { v := types.LoanBrokerID(""); return &v }(),
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "00002403C84A0A28E0190E208E982C352BBD5006600555CF",
					Value:         "1",
				},
			},
			expected: ErrMissingField{Field: "LoanBrokerID"},
		},
		{
			name: "pass - MPT Amount with LoanBrokerID",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: func() *types.LoanBrokerID {
					v := types.LoanBrokerID("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430")
					return &v
				}(),
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "00002403C84A0A28E0190E208E982C352BBD5006600555CF",
					Value:         "1",
				},
			},
			expected: nil,
		},
		{
			name: "pass - MPT pointer with LoanBrokerID",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{Account: "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds", TransactionType: LoanBrokerCoverClawbackTx},
				LoanBrokerID: func() *types.LoanBrokerID {
					v := types.LoanBrokerID("B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430")
					return &v
				}(),
				Amount: &types.MPTCurrencyAmount{MPTIssuanceID: "00002403C84A0A28E0190E208E982C352BBD5006600555CF", Value: "1"},
			},
		},
		{
			name: "fail - MPT pointer without LoanBrokerID",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{Account: "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds", TransactionType: LoanBrokerCoverClawbackTx},
				Amount: &types.MPTCurrencyAmount{MPTIssuanceID: "00002403C84A0A28E0190E208E982C352BBD5006600555CF", Value: "1"},
			},
			expected: ErrMissingField{Field: "LoanBrokerID"},
		},
		{
			name: "pass - with Amount only from another issuer",
			tx: &LoanBrokerCoverClawback{
				BaseTx: BaseTx{
					Account:         "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds",
					TransactionType: LoanBrokerCoverClawbackTx,
				},
				LoanBrokerID: nil,
				Amount: types.IssuedCurrencyAmount{
					Issuer:   types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"),
					Currency: "FOO",
					Value:    "0",
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
				require.Error(t, err)
				assert.Contains(t, err.Error(), testcase.expected.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
