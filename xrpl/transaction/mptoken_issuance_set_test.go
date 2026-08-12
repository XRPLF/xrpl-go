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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				Holder:            types.Holder("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Flags":             uint32(1),
				"MPTokenIssuanceID": "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Flags":             uint32(1),
				"MPTokenIssuanceID": "000004C463C52827307480341125DA0577DEFC38405B0E3E",
			},
		},
		{
			name: "pass - with MPTokenMetadata",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				MPTokenMetadata:   types.MPTokenMetadata("464f4f"),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				"MPTokenMetadata":   "464f4f",
			},
		},
		{
			name: "pass - with TransferFee",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				TransferFee:       types.TransferFee(314),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				"TransferFee":       uint16(314),
			},
		},
		{
			name: "pass - with ImmutableFlags",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				ImmutableFlags:    types.ImmutableFlags(TifMPTCanLock),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				"ImmutableFlags":    uint32(2),
			},
		},
		{
			name: "pass - with all DynamicMPT fields",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				MPTokenMetadata:   types.MPTokenMetadata("464f4f"),
				TransferFee:       types.TransferFee(314),
				ImmutableFlags:    types.ImmutableFlags(TifMPTCanLock),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				"MPTokenMetadata":   "464f4f",
				"TransferFee":       uint16(314),
				"ImmutableFlags":    uint32(2),
			},
		},
		{
			name: "pass - with DomainID",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				DomainID:          types.DomainID("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
			},
			expected: FlatTransaction{
				"TransactionType":   "MPTokenIssuanceSet",
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"MPTokenIssuanceID": "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
			wantErr: ErrInvalidMPTokenIssuanceIDSet,
		},
		{
			name: "fail - non-hex MPTokenIssuanceID",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "not-a-hex-value!",
			},
			wantOk:  false,
			wantErr: ErrInvalidMPTokenIssuanceIDSet,
		},
		{
			name: "fail - short hexadecimal MPTokenIssuanceID",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "000004C4",
			},
			wantOk:  false,
			wantErr: ErrInvalidMPTokenIssuanceIDSet,
		},
		{
			name: "fail - long hexadecimal MPTokenIssuanceID",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E00",
			},
			wantOk:  false,
			wantErr: ErrInvalidMPTokenIssuanceIDSet,
		},
		{
			name: "fail - no operation specified (no-op)",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetEmpty,
		},
		{
			name: "fail - holder without lock or unlock",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
			},
			wantOk:  false,
			wantErr: ErrMPTokenIssuanceSetFlags,
		},
		{
			name: "pass - universal Flags with mutation",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           types.TfUniversal,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				MPTokenMetadata:   types.MPTokenMetadata("464f4f"),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "fail - Flags contains unsupported bits",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           0x00000200,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetInvalidFlags,
		},
		{
			name: "fail - holder mutually exclusive with ImmutableFlags",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
				ImmutableFlags:    types.ImmutableFlags(TifMPTCanLock),
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
				MPTokenMetadata:   types.MPTokenMetadata("464f4f"),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetHolderMutuallyExclusive,
		},
		{
			name: "fail - lock mutually exclusive with DynamicMPT fields",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				ImmutableFlags:    types.ImmutableFlags(TifMPTCanLock),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetFlagsMutuallyExclusive,
		},
		{
			name: "fail - ImmutableFlags cannot be zero",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				ImmutableFlags:    types.ImmutableFlags(0),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetImmutableFlagsZero,
		},
		{
			name: "fail - ImmutableFlags contains unsupported bits",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				ImmutableFlags:    types.ImmutableFlags(0x00000001),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetInvalidImmutableFlags,
		},
		{
			name: "fail - TransferFee exceeds maximum",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				MPTokenMetadata:   types.MPTokenMetadata(""),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "pass - capability flags combine with DynamicMPT fields",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTSetCanLock | TfMPTSetCanTransfer,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				ImmutableFlags:    types.ImmutableFlags(TifMPTCanLock | TifMPTCanEscrow),
				TransferFee:       types.TransferFee(500),
				MPTokenMetadata:   types.MPTokenMetadata("464f4f"),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "fail - holder mutually exclusive with capability flags",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTSetCanLock,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetHolderMutuallyExclusive,
		},
		{
			name: "fail - lock mutually exclusive with capability flags",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock | TfMPTSetCanLock,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetFlagsMutuallyExclusive,
		},
		{
			name: "fail - non-zero TransferFee with confidential balances",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTSetCanHoldConfidentialBalance,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				TransferFee:       types.TransferFee(1),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetTransferFeeWithConfidentialBalance,
		},
		{
			name: "pass - valid DomainID",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
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
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				Holder:            types.Holder("rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2"),
				DomainID:          types.DomainID("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetHolderMutuallyExclusive,
		},
		{
			name: "pass - zero TransferFee alone is valid DynamicMPT operation",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				TransferFee:       types.TransferFee(0),
			},
			wantOk:  true,
			wantErr: nil,
		},
		{
			name: "fail - ImmutableFlags with lock returns mutual exclusivity error",
			tx: &MPTokenIssuanceSet{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceSetTx,
					Flags:           TfMPTLock,
				},
				MPTokenIssuanceID: "000004C463C52827307480341125DA0577DEFC38405B0E3E",
				ImmutableFlags:    types.ImmutableFlags(0),
			},
			wantOk:  false,
			wantErr: ErrMPTIssuanceSetFlagsMutuallyExclusive,
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

func TestMPTokenIssuanceSet_ImmutableFlags(t *testing.T) {
	tests := []struct {
		name    string
		setFlag func(*MPTokenIssuanceSet)
		want    uint32
	}{
		{
			name:    "MPTCanLock",
			setFlag: (*MPTokenIssuanceSet).SetMPTCanLockImmutableFlag,
			want:    TifMPTCanLock,
		},
		{
			name:    "MPTRequireAuth",
			setFlag: (*MPTokenIssuanceSet).SetMPTRequireAuthImmutableFlag,
			want:    TifMPTRequireAuth,
		},
		{
			name:    "MPTCanEscrow",
			setFlag: (*MPTokenIssuanceSet).SetMPTCanEscrowImmutableFlag,
			want:    TifMPTCanEscrow,
		},
		{
			name:    "MPTCanTrade",
			setFlag: (*MPTokenIssuanceSet).SetMPTCanTradeImmutableFlag,
			want:    TifMPTCanTrade,
		},
		{
			name:    "MPTCanTransfer",
			setFlag: (*MPTokenIssuanceSet).SetMPTCanTransferImmutableFlag,
			want:    TifMPTCanTransfer,
		},
		{
			name:    "MPTCanClawback",
			setFlag: (*MPTokenIssuanceSet).SetMPTCanClawbackImmutableFlag,
			want:    TifMPTCanClawback,
		},
		{
			name:    "MPTCanHoldConfidentialBalance",
			setFlag: (*MPTokenIssuanceSet).SetMPTCanHoldConfidentialBalanceImmutableFlag,
			want:    TifMPTCanHoldConfidentialBalance,
		},
		{
			name:    "MPTMetadata",
			setFlag: (*MPTokenIssuanceSet).SetMPTMetadataImmutableFlag,
			want:    TifMPTMetadata,
		},
		{
			name:    "MPTTransferFee",
			setFlag: (*MPTokenIssuanceSet).SetMPTTransferFeeImmutableFlag,
			want:    TifMPTTransferFee,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MPTokenIssuanceSet{}
			tt.setFlag(tx)
			require.NotNil(t, tx.ImmutableFlags)
			require.Equal(t, tt.want, *tx.ImmutableFlags)
		})
	}

	// Test all immutable flags together.
	tx := &MPTokenIssuanceSet{}
	for _, tt := range tests {
		tt.setFlag(tx)
	}

	expectedImmutableFlags := TifMPTCanLock | TifMPTRequireAuth | TifMPTCanEscrow |
		TifMPTCanTrade | TifMPTCanTransfer | TifMPTCanClawback |
		TifMPTCanHoldConfidentialBalance | TifMPTMetadata | TifMPTTransferFee
	require.Equal(t, uint32(expectedImmutableFlags), *tx.ImmutableFlags)
}

func TestMPTokenIssuanceSet_Flags(t *testing.T) {
	require.Equal(t, uint32(0x00000001), TfMPTLock)
	require.Equal(t, uint32(0x00000002), TfMPTUnlock)
	require.Equal(t, uint32(0x00000004), TfMPTSetCanLock)
	require.Equal(t, uint32(0x00000008), TfMPTSetRequireAuth)
	require.Equal(t, uint32(0x00000010), TfMPTSetCanEscrow)
	require.Equal(t, uint32(0x00000020), TfMPTSetCanTrade)
	require.Equal(t, uint32(0x00000040), TfMPTSetCanTransfer)
	require.Equal(t, uint32(0x00000080), TfMPTSetCanClawback)
	require.Equal(t, uint32(0x00000100), TfMPTSetCanHoldConfidentialBalance)

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
			name: "pass - set capability flags",
			setFlags: func(tx *MPTokenIssuanceSet) {
				tx.SetMPTCanLockFlag()
				tx.SetMPTRequireAuthFlag()
				tx.SetMPTCanEscrowFlag()
				tx.SetMPTCanTradeFlag()
				tx.SetMPTCanTransferFlag()
				tx.SetMPTCanClawbackFlag()
				tx.SetMPTCanHoldConfidentialBalanceFlag()
			},
			want: mpTokenIssuanceSetEnableFlagMask,
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
