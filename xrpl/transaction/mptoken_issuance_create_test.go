package transaction

import (
	"strings"
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
		{
			name: "pass - MPTokenIssuanceCreate with MutableFlags",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account: "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				},
				MutableFlags: types.MutableFlags(TmfMPTCanEnableCanLock | TmfMPTCanMutateMetadata),
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "MPTokenIssuanceCreate",
				"MutableFlags": 65538
			}`,
		},
		{
			name: "pass - MPTokenIssuanceCreate with DomainID",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account: "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					Flags:   TfMPTRequireAuth,
				},
				DomainID: types.DomainID("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "MPTokenIssuanceCreate",
				"Flags": 4,
				"DomainID": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"
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
		{
			name: "pass - TransferFee zero without TfMPTCanTransfer flag",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				TransferFee: types.TransferFee(0),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - MPTokenMetadata not valid hex",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MPTokenMetadata: types.MPTokenMetadata("not-hex!"),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidMPTokenMetadata,
		},
		{
			name: "fail - MPTokenMetadata exceeds 1024 bytes",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MPTokenMetadata: types.MPTokenMetadata(strings.Repeat("AB", 1025)),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidMPTokenMetadata,
		},
		{
			name: "pass - valid with MutableFlags",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MutableFlags: types.MutableFlags(TmfMPTCanEnableCanLock | TmfMPTCanMutateMetadata),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - MutableFlags cannot be zero",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MutableFlags: types.MutableFlags(0),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMPTIssuanceCreateInvalidMutableFlags,
		},
		{
			name: "pass - valid with DomainID and TfMPTRequireAuth",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
					Flags:           TfMPTRequireAuth,
				},
				DomainID: types.DomainID("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - DomainID without TfMPTRequireAuth",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				DomainID: types.DomainID("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMPTIssuanceCreateDomainIDRequiresRequireAuth,
		},
		{
			name: "fail - DomainID invalid hex",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
					Flags:           TfMPTRequireAuth,
				},
				DomainID: types.DomainID("not-valid"),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMPTIssuanceCreateDomainIDInvalid,
		},
		{
			name: "fail - TransferFee exceeds MaxTransferFee",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
					Flags:           TfMPTCanTransfer,
				},
				TransferFee: types.TransferFee(50001),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidTransferFee,
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

func TestMPTokenIssuanceCreateFlagValuesAndSettersPreserveBits(t *testing.T) {
	require.Equal(t, uint32(0x02), TfMPTCanLock)
	require.Equal(t, uint32(0x04), TfMPTRequireAuth)
	require.Equal(t, uint32(0x08), TfMPTCanEscrow)
	require.Equal(t, uint32(0x10), TfMPTCanTrade)
	require.Equal(t, uint32(0x20), TfMPTCanTransfer)
	require.Equal(t, uint32(0x40), TfMPTCanClawback)
	require.Equal(t, uint32(0x80), TfMPTCanHoldConfidentialBalance)

	tests := []struct {
		name    string
		setFlag func(*MPTokenIssuanceCreate)
		literal uint32
	}{
		{"MPTCanLock", (*MPTokenIssuanceCreate).SetMPTCanLockFlag, 0x02},
		{"MPTRequireAuth", (*MPTokenIssuanceCreate).SetMPTRequireAuthFlag, 0x04},
		{"MPTCanEscrow", (*MPTokenIssuanceCreate).SetMPTCanEscrowFlag, 0x08},
		{"MPTCanTrade", (*MPTokenIssuanceCreate).SetMPTCanTradeFlag, 0x10},
		{"MPTCanTransfer", (*MPTokenIssuanceCreate).SetMPTCanTransferFlag, 0x20},
		{"MPTCanClawback", (*MPTokenIssuanceCreate).SetMPTCanClawbackFlag, 0x40},
		{"MPTCanHoldConfidentialBalance", (*MPTokenIssuanceCreate).SetMPTCanHoldConfidentialBalanceFlag, 0x80},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &MPTokenIssuanceCreate{BaseTx: BaseTx{Flags: 0x100}}
			test.setFlag(tx)
			require.Equal(t, uint32(0x100)|test.literal, tx.Flags)
		})
	}
}

func TestMPTokenIssuanceCreateMutableFlagSettersPreserveBits(t *testing.T) {
	tests := []struct {
		name    string
		setFlag func(*MPTokenIssuanceCreate)
		literal uint32
	}{
		{"MPTCanEnableCanLock", (*MPTokenIssuanceCreate).SetMPTCanEnableCanLockFlag, 0x02},
		{"MPTCanEnableRequireAuth", (*MPTokenIssuanceCreate).SetMPTCanEnableRequireAuthFlag, 0x04},
		{"MPTCanEnableCanEscrow", (*MPTokenIssuanceCreate).SetMPTCanEnableCanEscrowFlag, 0x08},
		{"MPTCanEnableCanTrade", (*MPTokenIssuanceCreate).SetMPTCanEnableCanTradeFlag, 0x10},
		{"MPTCanEnableCanTransfer", (*MPTokenIssuanceCreate).SetMPTCanEnableCanTransferFlag, 0x20},
		{"MPTCanEnableCanClawback", (*MPTokenIssuanceCreate).SetMPTCanEnableCanClawbackFlag, 0x40},
		{"MPTCanMutateMetadata", (*MPTokenIssuanceCreate).SetMPTCanMutateMetadataFlag, 0x10000},
		{"MPTCanMutateTransferFee", (*MPTokenIssuanceCreate).SetMPTCanMutateTransferFeeFlag, 0x20000},
		{"MPTCannotEnableCanHoldConfidentialBalance", (*MPTokenIssuanceCreate).SetMPTCannotEnableCanHoldConfidentialBalanceFlag, 0x80},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &MPTokenIssuanceCreate{MutableFlags: types.MutableFlags(0x80000000)}
			test.setFlag(tx)
			require.Equal(t, uint32(0x80000000)|test.literal, *tx.MutableFlags)
		})
	}
}

func TestMPTokenIssuanceCreateMutableFlagValuesAndMask(t *testing.T) {
	require.Equal(t, uint32(0x02), TmfMPTCanEnableCanLock)
	require.Equal(t, uint32(0x04), TmfMPTCanEnableRequireAuth)
	require.Equal(t, uint32(0x08), TmfMPTCanEnableCanEscrow)
	require.Equal(t, uint32(0x10), TmfMPTCanEnableCanTrade)
	require.Equal(t, uint32(0x20), TmfMPTCanEnableCanTransfer)
	require.Equal(t, uint32(0x40), TmfMPTCanEnableCanClawback)
	require.Equal(t, uint32(0x80), TmfMPTCannotEnableCanHoldConfidentialBalance)
	require.Equal(t, uint32(0x10000), TmfMPTCanMutateMetadata)
	require.Equal(t, uint32(0x20000), TmfMPTCanMutateTransferFee)
	require.Equal(t, uint32(0x000300FE), MPTokenIssuanceCreateMutableFlagsMask)
}

func TestMPTokenIssuanceCreate_ValidateMutableFlagsMask(t *testing.T) {
	tests := []struct {
		name  string
		flags uint32
		ok    bool
	}{
		{"one valid bit", TmfMPTCanEnableCanLock, true},
		{"all valid bits", MPTokenIssuanceCreateMutableFlagsMask, true},
		{"present zero", 0, false},
		{"reserved bit 0x01", 0x01, false},
		{"removed bit 0x40000", 0x40000, false},
		{"unknown high bit", 0x80000000, false},
		{"valid and unknown", TmfMPTCanMutateMetadata | 0x100, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &MPTokenIssuanceCreate{
				BaseTx:       BaseTx{Account: "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2", TransactionType: MPTokenIssuanceCreateTx},
				MutableFlags: types.MutableFlags(test.flags),
			}
			ok, err := tx.Validate()
			require.Equal(t, test.ok, ok)
			if test.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrMPTIssuanceCreateInvalidMutableFlags)
			}
		})
	}
}

func TestMPTokenIssuanceCreate_TransferFeeConfidentialBalance(t *testing.T) {
	tests := []struct {
		name  string
		fee   *uint16
		flags uint32
		ok    bool
	}{
		{"non-zero fee with confidential", types.TransferFee(1), TfMPTCanTransfer | TfMPTCanHoldConfidentialBalance, false},
		{"zero fee with confidential", types.TransferFee(0), TfMPTCanHoldConfidentialBalance, true},
		{"absent fee with confidential", nil, TfMPTCanHoldConfidentialBalance, true},
		{"non-zero fee transferable", types.TransferFee(1), TfMPTCanTransfer, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &MPTokenIssuanceCreate{
				BaseTx:      BaseTx{Account: "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2", TransactionType: MPTokenIssuanceCreateTx, Flags: test.flags},
				TransferFee: test.fee,
			}
			ok, err := tx.Validate()
			require.Equal(t, test.ok, ok)
			if test.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrMPTIssuanceCreateTransferFeeWithConfidentialBalance)
			}
		})
	}
}

func TestMPTokenIssuanceCreate_SettersPreserveExistingBits(t *testing.T) {
	tx := &MPTokenIssuanceCreate{
		BaseTx:       BaseTx{Flags: 0x100},
		MutableFlags: types.MutableFlags(0x80000000),
	}
	tx.SetMPTCanHoldConfidentialBalanceFlag()
	tx.SetMPTCanEnableCanLockFlag()
	require.Equal(t, uint32(0x180), tx.Flags)
	require.Equal(t, uint32(0x80000002), *tx.MutableFlags)
}
