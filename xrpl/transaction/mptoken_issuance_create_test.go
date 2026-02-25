package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestMPTokenIssuanceCreate_TxType(t *testing.T) {
	tx := &MPTokenIssuanceCreate{}
	require.Equal(t, MPTokenIssuanceCreateTx, tx.TxType())
}

func TestMPTokenIssuanceCreate_Flatten(t *testing.T) {
	amount := types.XRPCurrencyAmount(10000)

	tests := []struct {
		name     string
		tx       *MPTokenIssuanceCreate
		expected string
	}{
		{
			name: "pass - BaseTx only MPTokenIssuanceCreate",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "MPTokenIssuanceCreate"
			}`,
		},
		{
			name: "pass - MPTokenIssuanceCreate with all fields",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				AssetScale:      types.AssetScale(2),
				TransferFee:     types.TransferFee(314),
				MaximumAmount:   &amount,
				MPTokenMetadata: types.MPTokenMetadata("FOO"),
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "MPTokenIssuanceCreate",
				"AssetScale": 2,
				"TransferFee": 314,
				"MaximumAmount": "10000",
				"MPTokenMetadata": "FOO"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := testutil.CompareFlattenAndExpected(tt.tx.Flatten(), []byte(tt.expected)); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestMPTokenIssuanceCreate_Validate(t *testing.T) {
	amount := types.XRPCurrencyAmount(10000)
	tests := []struct {
		name       string
		tx         *MPTokenIssuanceCreate
		wantValid  bool
		wantErr    bool
		errMessage error
	}{
		{
			name: "pass - valid with all fields",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
					Flags:           TfMPTCanTransfer,
				},
				AssetScale:      types.AssetScale(2),
				TransferFee:     types.TransferFee(314),
				MaximumAmount:   &amount,
				MPTokenMetadata: types.MPTokenMetadata("464f4f"),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - valid with minimal fields",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MPTokenMetadata: types.MPTokenMetadata("464f4f"),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid account",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "invalid",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MPTokenMetadata: types.MPTokenMetadata("464f4f"),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidAccount,
		},
		{
			name: "fail - invalid flags",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MPTokenMetadata: types.MPTokenMetadata("464f4f"),
				TransferFee:     types.TransferFee(314),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrTransferFeeRequiresCanTransfer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.errMessage, err)
				require.False(t, valid)
			} else {
				require.NoError(t, err)
				require.True(t, valid)
			}
		})
	}
}

func TestMPTokenIssuanceCreate_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setFlag  func(*MPTokenIssuanceCreate)
		flagMask uint32
	}{
		{
			name:     "MPTCanLock",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanLockFlag,
			flagMask: TfMPTCanLock,
		},
		{
			name:     "MPTRequireAuth",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTRequireAuthFlag,
			flagMask: TfMPTRequireAuth,
		},
		{
			name:     "MPTCanEscrow",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanEscrowFlag,
			flagMask: TfMPTCanEscrow,
		},
		{
			name:     "MPTCanTrade",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanTradeFlag,
			flagMask: TfMPTCanTrade,
		},
		{
			name:     "MPTCanTransfer",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanTransferFlag,
			flagMask: TfMPTCanTransfer,
		},
		{
			name:     "MPTCanClawback",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanClawbackFlag,
			flagMask: TfMPTCanClawback,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MPTokenMetadata: types.MPTokenMetadata("464f4f"),
			}

			tt.setFlag(tx)
			require.Equal(t, uint32(tt.flagMask), tx.Flags&tt.flagMask)
		})
	}

	// Test all flags together
	tx := &MPTokenIssuanceCreate{
		BaseTx: BaseTx{
			Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
			TransactionType: MPTokenIssuanceCreateTx,
		},
		MPTokenMetadata: types.MPTokenMetadata("464f4f"),
	}

	for _, tt := range tests {
		tt.setFlag(tx)
	}

	expectedFlags := TfMPTCanLock | TfMPTRequireAuth | TfMPTCanEscrow | TfMPTCanTrade | TfMPTCanTransfer | TfMPTCanClawback
	require.Equal(t, uint32(expectedFlags), tx.Flags)
}
