package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestMPTokenIssuanceSet_TxType(t *testing.T) {
	tx := &MPTokenIssuanceSet{}
	require.Equal(t, MPTokenIssuanceSetTx, tx.TxType())
}

func TestMPTokenIssuanceSet_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *MPTokenIssuanceSet
		expected FlatTransaction
	}{
		{
			name: "pass - with holder",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Flags:   1,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				Holder:            types.Holder("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Flags":             uint32(1),
				"MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"Holder":            "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			},
		},
		{
			name: "pass - without holder",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Flags:   1,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Flags":             uint32(1),
				"MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
			},
		},
		{
			name: "pass - with MPTokenMetadata",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTokenMetadata:   types.MPTokenMetadata("464f4f"),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"MPTokenMetadata":   "464f4f",
			},
		},
		{
			name: "pass - with TransferFee",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				TransferFee:       types.TransferFee(314),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"TransferFee":       314,
			},
		},
		{
			name: "pass - with MutableFlags",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(TmfMPTSetCanLock),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"MutableFlags":      1,
			},
		},
		{
			name: "pass - with all DynamicMPT fields",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTokenMetadata:   types.MPTokenMetadata("464f4f"),
				TransferFee:       types.TransferFee(314),
				MutableFlags:      types.MutableFlags(TmfMPTSetCanLock),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"MPTokenMetadata":   "464f4f",
				"TransferFee":       314,
				"MutableFlags":      1,
			},
		},
		{
			name: "pass - with DomainID",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				DomainID:          types.DomainID("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"DomainID":          "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattened := tt.tx.Flatten()
			require.Equal(t, tt.expected, flattened)
		})
	}
}

func TestMPTokenIssuanceSet_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *MPTokenIssuanceSet
		wantOk  bool
		wantErr error
	}{
		{
			name: "pass - valid transaction with holder",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "",
			},
			wantOk:  false,
			wantErr: ErrInvalidMPTokenIssuanceID,
		},
		{
			name: "fail - no operation specified (no-op)",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetEmpty,
		},
		{
			name: "fail - invalid holder address",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				Holder:            types.Holder("invalid"),
			},
			wantOk:  false,
			wantErr: ErrInvalidAccount,
		},
		{
			name: "fail - holder same as account",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				Holder:            types.Holder("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
			},
			wantOk:  false,
			wantErr: ErrHolderAccountConflict,
		},
		{
			name: "fail - conflicting lock/unlock flags",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock | TfMPTUnlock,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
			},
			wantOk:  false,
			wantErr: ErrMPTokenIssuanceSetFlags,
		},
		{
			name: "fail - holder mutually exclusive with MutableFlags",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
				MutableFlags:      types.MutableFlags(TmfMPTSetCanLock),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetHolderMutuallyExclusive,
		},
		{
			name: "fail - holder mutually exclusive with MPTokenMetadata",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
				MPTokenMetadata:   types.MPTokenMetadata("464f4f"),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetHolderMutuallyExclusive,
		},
		{
			name: "fail - flags mutually exclusive with DynamicMPT fields",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(TmfMPTSetCanLock),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetFlagsMutuallyExclusive,
		},
		{
			name: "fail - MutableFlags cannot be zero",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(0),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetMutableFlagsZero,
		},
		{
			name: "fail - MutableFlags set/clear conflict",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(TmfMPTSetCanLock | TmfMPTClearCanLock),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetMutableFlagsConflict,
		},
		{
			name: "fail - TransferFee exceeds maximum",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				TransferFee:       types.TransferFee(50001),
			},
			wantOk:  false,
			wantErr: ErrInvalidTransferFee,
		},
		{
			name: "fail - invalid hex MPTokenMetadata",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTokenMetadata:   types.MPTokenMetadata("not-hex!"),
			},
			wantOk:  false,
			wantErr: ErrInvalidMPTokenMetadata,
		},
		{
			name: "fail - MPTokenMetadata exceeds 1024 bytes",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTokenMetadata:   types.MPTokenMetadata(strings.Repeat("AB", 1025)),
			},
			wantOk:  false,
			wantErr: ErrInvalidMPTokenMetadata,
		},
		{
			name: "pass - MPTokenMetadata exactly 1024 bytes",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTokenMetadata:   types.MPTokenMetadata(strings.Repeat("AB", 1024)),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "pass - empty MPTokenMetadata removes field",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTokenMetadata:   types.MPTokenMetadata(""),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "pass - valid DynamicMPT usage with all fields",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(TmfMPTSetCanLock | TmfMPTClearCanEscrow),
				TransferFee:       types.TransferFee(500),
				MPTokenMetadata:   types.MPTokenMetadata("464f4f"),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "pass - valid DomainID",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				DomainID:          types.DomainID("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "pass - empty DomainID removes domain",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				DomainID:          types.DomainID(""),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "fail - DomainID invalid hex",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				DomainID:          types.DomainID("not-valid"),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetDomainIDInvalid,
		},
		{
			name: "fail - DomainID mutually exclusive with Holder",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
				DomainID:          types.DomainID("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetHolderMutuallyExclusive,
		},
		{
			name: "fail - non-zero TransferFee with tmfMPTClearCanTransfer",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				TransferFee:       types.TransferFee(200),
				MutableFlags:      types.MutableFlags(TmfMPTClearCanTransfer),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetTransferFeeWithClearCanTransfer,
		},
		{
			name: "pass - zero TransferFee with tmfMPTClearCanTransfer",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				TransferFee:       types.TransferFee(0),
				MutableFlags:      types.MutableFlags(TmfMPTClearCanTransfer),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "pass - zero TransferFee alone is valid DynamicMPT operation",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				TransferFee:       types.TransferFee(0),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "fail - MutableFlags with Flags returns mutual exclusivity error",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(0),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetFlagsMutuallyExclusive,
		},
		{
			name: "fail - MutableFlags set/clear conflict RequireAuth",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(TmfMPTSetRequireAuth | TmfMPTClearRequireAuth),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetMutableFlagsConflict,
		},
		{
			name: "fail - MutableFlags set/clear conflict CanEscrow",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(TmfMPTSetCanEscrow | TmfMPTClearCanEscrow),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetMutableFlagsConflict,
		},
		{
			name: "fail - MutableFlags set/clear conflict CanTrade",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(TmfMPTSetCanTrade | TmfMPTClearCanTrade),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetMutableFlagsConflict,
		},
		{
			name: "fail - MutableFlags set/clear conflict CanTransfer",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(TmfMPTSetCanTransfer | TmfMPTClearCanTransfer),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetMutableFlagsConflict,
		},
		{
			name: "fail - MutableFlags set/clear conflict CanClawback",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MutableFlags:      types.MutableFlags(TmfMPTSetCanClawback | TmfMPTClearCanClawback),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetMutableFlagsConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := tt.tx.Validate()
			require.Equal(t, tt.wantOk, ok)
			require.Equal(t, tt.wantErr, err)
		})
	}
}

func TestMPTokenIssuanceSet_MutableFlags(t *testing.T) {
	tests := []struct {
		name     string
		setFlag  func(*MPTokenIssuanceSet)
		flagMask uint32
	}{
		{
			name:     "MPTSetCanLock",
			setFlag:  (*MPTokenIssuanceSet).SetMPTSetCanLockMutableFlag,
			flagMask: TmfMPTSetCanLock,
		},
		{
			name:     "MPTClearCanLock",
			setFlag:  (*MPTokenIssuanceSet).SetMPTClearCanLockMutableFlag,
			flagMask: TmfMPTClearCanLock,
		},
		{
			name:     "MPTSetRequireAuth",
			setFlag:  (*MPTokenIssuanceSet).SetMPTSetRequireAuthMutableFlag,
			flagMask: TmfMPTSetRequireAuth,
		},
		{
			name:     "MPTClearRequireAuth",
			setFlag:  (*MPTokenIssuanceSet).SetMPTClearRequireAuthMutableFlag,
			flagMask: TmfMPTClearRequireAuth,
		},
		{
			name:     "MPTSetCanEscrow",
			setFlag:  (*MPTokenIssuanceSet).SetMPTSetCanEscrowMutableFlag,
			flagMask: TmfMPTSetCanEscrow,
		},
		{
			name:     "MPTClearCanEscrow",
			setFlag:  (*MPTokenIssuanceSet).SetMPTClearCanEscrowMutableFlag,
			flagMask: TmfMPTClearCanEscrow,
		},
		{
			name:     "MPTSetCanTrade",
			setFlag:  (*MPTokenIssuanceSet).SetMPTSetCanTradeMutableFlag,
			flagMask: TmfMPTSetCanTrade,
		},
		{
			name:     "MPTClearCanTrade",
			setFlag:  (*MPTokenIssuanceSet).SetMPTClearCanTradeMutableFlag,
			flagMask: TmfMPTClearCanTrade,
		},
		{
			name:     "MPTSetCanTransfer",
			setFlag:  (*MPTokenIssuanceSet).SetMPTSetCanTransferMutableFlag,
			flagMask: TmfMPTSetCanTransfer,
		},
		{
			name:     "MPTClearCanTransfer",
			setFlag:  (*MPTokenIssuanceSet).SetMPTClearCanTransferMutableFlag,
			flagMask: TmfMPTClearCanTransfer,
		},
		{
			name:     "MPTSetCanClawback",
			setFlag:  (*MPTokenIssuanceSet).SetMPTSetCanClawbackMutableFlag,
			flagMask: TmfMPTSetCanClawback,
		},
		{
			name:     "MPTClearCanClawback",
			setFlag:  (*MPTokenIssuanceSet).SetMPTClearCanClawbackMutableFlag,
			flagMask: TmfMPTClearCanClawback,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MPTokenIssuanceSet{}
			tt.setFlag(tx)
			require.NotNil(t, tx.MutableFlags)
			require.Equal(t, tt.flagMask, *tx.MutableFlags)
		})
	}

	// Test all mutable flags together
	tx := &MPTokenIssuanceSet{}
	for _, tt := range tests {
		tt.setFlag(tx)
	}

	expectedMutableFlags := TmfMPTSetCanLock | TmfMPTClearCanLock |
		TmfMPTSetRequireAuth | TmfMPTClearRequireAuth |
		TmfMPTSetCanEscrow | TmfMPTClearCanEscrow |
		TmfMPTSetCanTrade | TmfMPTClearCanTrade |
		TmfMPTSetCanTransfer | TmfMPTClearCanTransfer |
		TmfMPTSetCanClawback | TmfMPTClearCanClawback
	require.Equal(t, expectedMutableFlags, *tx.MutableFlags)
}

func TestMPTokenIssuanceSet_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setFlags func(*MPTokenIssuanceSet)
		want     uint32
	}{
		{
			name: "pass - set MPTLock flag",
			setFlags: func(tx *MPTokenIssuanceSet) {
				tx.SetMPTLockFlag()
			},
			want: TfMPTLock,
		},
		{
			name: "pass - set MPTUnlock flag",
			setFlags: func(tx *MPTokenIssuanceSet) {
				tx.SetMPTUnlockFlag()
			},
			want: TfMPTUnlock,
		},
		{
			name: "pass - set both flags",
			setFlags: func(tx *MPTokenIssuanceSet) {
				tx.SetMPTLockFlag()
				tx.SetMPTUnlockFlag()
			},
			want: TfMPTLock | TfMPTUnlock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MPTokenIssuanceSet{}
			tt.setFlags(tx)
			require.Equal(t, tt.want, tx.Flags)
		})
	}
}
