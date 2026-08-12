package transaction

import (
	"encoding/json"
	"strings"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestMPTokenIssuanceCreate_TxType(t *testing.T) {
	tx := &MPTokenIssuanceCreate{}
	require.Equal(t, MPTokenIssuanceCreateTx, tx.TxType())
}

func TestMPTokenIssuanceCreate_Flatten(t *testing.T) {
	amount := types.MPTAmount(10000)

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
			name: "pass - MPTokenIssuanceCreate with ImmutableFlags",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account: "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				},
				ImmutableFlags: types.ImmutableFlags(TifMPTCanLock | TifMPTMetadata),
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "MPTokenIssuanceCreate",
				"ImmutableFlags": 65538
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
	amount := types.MPTAmount(10000)
	zeroAmount := types.MPTAmount(0)
	maxAmount := types.MaxMPTAmount
	tooLargeAmount := types.MPTAmount(uint64(types.MaxMPTAmount) + 1)
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
			name: "pass - maximum MaximumAmount",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MaximumAmount: &maxAmount,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - zero MaximumAmount",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MaximumAmount: &zeroAmount,
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMPTIssuanceCreateMaximumAmountInvalid,
		},
		{
			name: "fail - MaximumAmount above protocol maximum",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				MaximumAmount: &tooLargeAmount,
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMPTIssuanceCreateMaximumAmountInvalid,
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
			name: "pass - valid with ImmutableFlags",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				ImmutableFlags: types.ImmutableFlags(TifMPTCanLock | TifMPTMetadata),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - ImmutableFlags cannot be zero",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				ImmutableFlags: types.ImmutableFlags(0),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMPTIssuanceCreateImmutableFlagsZero,
		},
		{
			name: "fail - ImmutableFlags contains unsupported bits",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
				},
				ImmutableFlags: types.ImmutableFlags(0x00000001),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMPTIssuanceCreateInvalidImmutableFlags,
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
		{
			name: "fail - non-zero TransferFee with confidential balances",
			tx: &MPTokenIssuanceCreate{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: MPTokenIssuanceCreateTx,
					Flags:           TfMPTCanTransfer | TfMPTCanHoldConfidentialBalance,
				},
				TransferFee: types.TransferFee(1),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMPTIssuanceCreateTransferFeeWithConfidentialBalance,
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

func TestMPTokenIssuanceCreate_MaximumAmountJSON(t *testing.T) {
	const input = `{
		"TransactionType":"MPTokenIssuanceCreate",
		"Account":"rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
		"MaximumAmount":"9223372036854775807"
	}`

	var tx MPTokenIssuanceCreate
	err := json.Unmarshal([]byte(input), &tx)
	require.NoError(t, err)
	require.NotNil(t, tx.MaximumAmount)
	require.Equal(t, types.MaxMPTAmount, *tx.MaximumAmount)

	encoded, err := json.Marshal(tx)
	require.NoError(t, err)
	flattened := make(map[string]any)
	require.NoError(t, json.Unmarshal(encoded, &flattened))
	require.Equal(t, "9223372036854775807", flattened["MaximumAmount"])

	for _, invalid := range []string{`10000`, `"0x10"`, `"9223372036854775808"`} {
		t.Run("reject "+invalid, func(t *testing.T) {
			payload := `{"MaximumAmount":` + invalid + `}`
			var invalidTx MPTokenIssuanceCreate
			err := json.Unmarshal([]byte(payload), &invalidTx)
			require.ErrorIs(t, err, types.ErrInvalidMPTAmount)
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
		{
			name:     "MPTCanHoldConfidentialBalance",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanHoldConfidentialBalanceFlag,
			flagMask: TfMPTCanHoldConfidentialBalance,
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

	expectedFlags := TfMPTCanLock | TfMPTRequireAuth | TfMPTCanEscrow | TfMPTCanTrade |
		TfMPTCanTransfer | TfMPTCanClawback | TfMPTCanHoldConfidentialBalance
	require.Equal(t, uint32(expectedFlags), tx.Flags)
}

func TestMPTokenIssuanceCreate_ImmutableFlags(t *testing.T) {
	require.Equal(t, uint32(0x00000002), TifMPTCanLock)
	require.Equal(t, uint32(0x00000004), TifMPTRequireAuth)
	require.Equal(t, uint32(0x00000008), TifMPTCanEscrow)
	require.Equal(t, uint32(0x00000010), TifMPTCanTrade)
	require.Equal(t, uint32(0x00000020), TifMPTCanTransfer)
	require.Equal(t, uint32(0x00000040), TifMPTCanClawback)
	require.Equal(t, uint32(0x00000080), TifMPTCanHoldConfidentialBalance)
	require.Equal(t, uint32(0x00010000), TifMPTMetadata)
	require.Equal(t, uint32(0x00020000), TifMPTTransferFee)

	tests := []struct {
		name     string
		setFlag  func(*MPTokenIssuanceCreate)
		flagMask uint32
	}{
		{
			name:     "MPTCanLock",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanLockImmutableFlag,
			flagMask: TifMPTCanLock,
		},
		{
			name:     "MPTRequireAuth",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTRequireAuthImmutableFlag,
			flagMask: TifMPTRequireAuth,
		},
		{
			name:     "MPTCanEscrow",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanEscrowImmutableFlag,
			flagMask: TifMPTCanEscrow,
		},
		{
			name:     "MPTCanTrade",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanTradeImmutableFlag,
			flagMask: TifMPTCanTrade,
		},
		{
			name:     "MPTCanTransfer",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanTransferImmutableFlag,
			flagMask: TifMPTCanTransfer,
		},
		{
			name:     "MPTCanClawback",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanClawbackImmutableFlag,
			flagMask: TifMPTCanClawback,
		},
		{
			name:     "MPTCanHoldConfidentialBalance",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTCanHoldConfidentialBalanceImmutableFlag,
			flagMask: TifMPTCanHoldConfidentialBalance,
		},
		{
			name:     "MPTMetadata",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTMetadataImmutableFlag,
			flagMask: TifMPTMetadata,
		},
		{
			name:     "MPTTransferFee",
			setFlag:  (*MPTokenIssuanceCreate).SetMPTTransferFeeImmutableFlag,
			flagMask: TifMPTTransferFee,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MPTokenIssuanceCreate{}
			tt.setFlag(tx)
			require.NotNil(t, tx.ImmutableFlags)
			require.Equal(t, tt.flagMask, *tx.ImmutableFlags)
		})
	}

	// Test all immutable flags together.
	tx := &MPTokenIssuanceCreate{}
	for _, tt := range tests {
		tt.setFlag(tx)
	}

	expectedImmutableFlags := TifMPTCanLock | TifMPTRequireAuth | TifMPTCanEscrow |
		TifMPTCanTrade | TifMPTCanTransfer | TifMPTCanClawback |
		TifMPTCanHoldConfidentialBalance | TifMPTMetadata | TifMPTTransferFee
	require.Equal(t, uint32(expectedImmutableFlags), *tx.ImmutableFlags)
}

func TestMPTokenIssuanceCreate_SigningPayloadUsesDecimalMaximumAmount(t *testing.T) {
	maximumAmount := types.MPTAmount(10000)
	tx := MPTokenIssuanceCreate{
		BaseTx: BaseTx{
			Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
			TransactionType: MPTokenIssuanceCreateTx,
			Fee:             types.XRPCurrencyAmount(12),
			Sequence:        1,
		},
		MaximumAmount: &maximumAmount,
	}

	valid, err := tx.Validate()
	require.NoError(t, err)
	require.True(t, valid)

	signingPayload, err := binarycodec.EncodeForSigning(tx.Flatten())
	require.NoError(t, err)
	require.Contains(t, signingPayload, "30180000000000002710")
}
