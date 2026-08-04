package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestMPTokenIssuance_EntryType(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	require.Equal(t, MPTokenIssuanceEntry, mpTokenIssuance.EntryType())
}

func TestMPTokenIssuance_SetLsfMPTLocked(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsfMPTLocked()
	require.Equal(t, LsfMPTLocked, mpTokenIssuance.Flags)
}

func TestMPTokenIssuance_SetLsfMPTCanLock(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsfMPTCanLock()
	require.Equal(t, LsfMPTCanLock, mpTokenIssuance.Flags)
}

func TestMPTokenIssuance_SetLsfMPTRequireAuth(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsfMPTRequireAuth()
	require.Equal(t, LsfMPTRequireAuth, mpTokenIssuance.Flags)
}

func TestMPTokenIssuance_SetLsfMPTCanEscrow(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsfMPTCanEscrow()
	require.Equal(t, LsfMPTCanEscrow, mpTokenIssuance.Flags)
}

func TestMPTokenIssuance_SetLsfMPTCanTrade(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsfMPTCanTrade()
	require.Equal(t, LsfMPTCanTrade, mpTokenIssuance.Flags)
}

func TestMPTokenIssuance_SetLsfMPTCanTransfer(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsfMPTCanTransfer()
	require.Equal(t, LsfMPTCanTransfer, mpTokenIssuance.Flags)
}

func TestMPTokenIssuance_SetLsfMPTCanClawback(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsfMPTCanClawback()
	require.Equal(t, LsfMPTCanClawback, mpTokenIssuance.Flags)
}

func TestMPTokenIssuanceSerialization(t *testing.T) {
	tests := []struct {
		name            string
		mpTokenIssuance *MPTokenIssuance
		expected        string
	}{
		{
			name: "pass - valid MPToken with LsfMPTLocked",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTLocked,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 1,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1
}`,
		},
		{
			name: "pass - valid MPToken with LsfMPTCanLock",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTCanLock,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 2,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1
}`,
		},
		{
			name: "pass - valid MPToken with LsfMPTRequireAuth",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTRequireAuth,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 4,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1
}`,
		},
		{
			name: "pass - valid MPToken with LsfMPTCanEscrow",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTCanEscrow,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 8,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1
}`,
		},
		{
			name: "pass - valid MPToken with LsfMPTCanTrade",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTCanTrade,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 16,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1
}`,
		},
		{
			name: "pass - valid MPToken with LsfMPTCanTransfer",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTCanTransfer,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 32,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1
}`,
		},
		{
			name: "pass - valid MPToken with LsfMPTCanClawback",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTCanClawback,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
			},

			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 64,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1
}`,
		},
		{
			name: "pass - valid MPToken with DomainID",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTRequireAuth,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
				DomainID:          "B738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 4,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1,
	"DomainID": "B738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"
}`,
		},
		{
			name: "pass - valid MPToken with MutableFlags",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTCanLock | LsfMPTCanTransfer,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
				MutableFlags:      LsmfMPTCanEnableCanLock | LsmfMPTCanMutateMetadata,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 34,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1,
	"MutableFlags": 65538
}`,
		},
		{
			name: "pass - valid MPTokenIssuance with LsfMPTCanHoldConfidentialBalance",
			mpTokenIssuance: &MPTokenIssuance{
				Index:             types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:   MPTokenIssuanceEntry,
				Flags:             LsfMPTCanHoldConfidentialBalance,
				Issuer:            types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:        2,
				MaximumAmount:     1000,
				OutstandingAmount: 100,
				TransferFee:       100,
				MPTokenMetadata:   "7B227469636B6572",
				OwnerNode:         1,
				PreviousTxnID:     types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq: 234644,
				Sequence:          1,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 128,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1
}`,
		},
		{
			name: "pass - valid MPTokenIssuance with confidential transfer fields",
			mpTokenIssuance: &MPTokenIssuance{
				Index:                         types.Hash256("A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9"),
				LedgerEntryType:               MPTokenIssuanceEntry,
				Flags:                         LsfMPTCanHoldConfidentialBalance | LsfMPTCanTransfer,
				Issuer:                        types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				AssetScale:                    2,
				MaximumAmount:                 1000,
				OutstandingAmount:             100,
				TransferFee:                   100,
				MPTokenMetadata:               "7B227469636B6572",
				OwnerNode:                     1,
				PreviousTxnID:                 types.Hash256("8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB"),
				PreviousTxnLgrSeq:             234644,
				Sequence:                      1,
				IssuerEncryptionKey:           "AABBCCDD",
				AuditorEncryptionKey:          "EEFF0011",
				ConfidentialOutstandingAmount: 500,
			},
			expected: `{
	"index": "A738A1E6E8505E1FC77BBB9FEF84FF9A9C609F2739E0F9573CDD6367100A0AA9",
	"LedgerEntryType": "MPTokenIssuance",
	"Flags": 160,
	"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"AssetScale": 2,
	"MaximumAmount": 1000,
	"OutstandingAmount": 100,
	"TransferFee": 100,
	"MPTokenMetadata": "7B227469636B6572",
	"OwnerNode": 1,
	"PreviousTxnID": "8089451B193AAD110ACED3D62BE79BB523658545E6EE8B7BB0BE573FED9BCBFB",
	"PreviousTxnLgrSeq": 234644,
	"Sequence": 1,
	"IssuerEncryptionKey": "AABBCCDD",
	"AuditorEncryptionKey": "EEFF0011",
	"ConfidentialOutstandingAmount": 500
}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, test.mpTokenIssuance, test.expected); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestMPTokenIssuanceFlagValuesAndSettersPreserveBits(t *testing.T) {
	require.Equal(t, uint32(0x02), LsfMPTCanLock)
	require.Equal(t, uint32(0x04), LsfMPTRequireAuth)
	require.Equal(t, uint32(0x08), LsfMPTCanEscrow)
	require.Equal(t, uint32(0x10), LsfMPTCanTrade)
	require.Equal(t, uint32(0x20), LsfMPTCanTransfer)
	require.Equal(t, uint32(0x40), LsfMPTCanClawback)
	require.Equal(t, uint32(0x80), LsfMPTCanHoldConfidentialBalance)

	tests := []struct {
		name    string
		set     func(*MPTokenIssuance)
		literal uint32
	}{
		{"CanLock", (*MPTokenIssuance).SetLsfMPTCanLock, 0x02},
		{"RequireAuth", (*MPTokenIssuance).SetLsfMPTRequireAuth, 0x04},
		{"CanEscrow", (*MPTokenIssuance).SetLsfMPTCanEscrow, 0x08},
		{"CanTrade", (*MPTokenIssuance).SetLsfMPTCanTrade, 0x10},
		{"CanTransfer", (*MPTokenIssuance).SetLsfMPTCanTransfer, 0x20},
		{"CanClawback", (*MPTokenIssuance).SetLsfMPTCanClawback, 0x40},
		{"CanHoldConfidentialBalance", (*MPTokenIssuance).SetLsfMPTCanHoldConfidentialBalance, 0x80},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issuance := &MPTokenIssuance{Flags: 0x100}
			test.set(issuance)
			require.Equal(t, uint32(0x100)|test.literal, issuance.Flags)
		})
	}
}

func TestMPTokenIssuanceMutableFlagValuesAndSettersPreserveBits(t *testing.T) {
	require.Equal(t, uint32(0x02), LsmfMPTCanEnableCanLock)
	require.Equal(t, uint32(0x04), LsmfMPTCanEnableRequireAuth)
	require.Equal(t, uint32(0x08), LsmfMPTCanEnableCanEscrow)
	require.Equal(t, uint32(0x10), LsmfMPTCanEnableCanTrade)
	require.Equal(t, uint32(0x20), LsmfMPTCanEnableCanTransfer)
	require.Equal(t, uint32(0x40), LsmfMPTCanEnableCanClawback)
	require.Equal(t, uint32(0x80), LsmfMPTCannotEnableCanHoldConfidentialBalance)
	require.Equal(t, uint32(0x10000), LsmfMPTCanMutateMetadata)
	require.Equal(t, uint32(0x20000), LsmfMPTCanMutateTransferFee)
	require.Equal(t, uint32(0x000300FE), MPTokenIssuanceMutableFlagsMask)

	tests := []struct {
		name    string
		set     func(*MPTokenIssuance)
		literal uint32
	}{
		{"CanEnableCanLock", (*MPTokenIssuance).SetLsmfMPTCanEnableCanLock, 0x02},
		{"CanEnableRequireAuth", (*MPTokenIssuance).SetLsmfMPTCanEnableRequireAuth, 0x04},
		{"CanEnableCanEscrow", (*MPTokenIssuance).SetLsmfMPTCanEnableCanEscrow, 0x08},
		{"CanEnableCanTrade", (*MPTokenIssuance).SetLsmfMPTCanEnableCanTrade, 0x10},
		{"CanEnableCanTransfer", (*MPTokenIssuance).SetLsmfMPTCanEnableCanTransfer, 0x20},
		{"CanEnableCanClawback", (*MPTokenIssuance).SetLsmfMPTCanEnableCanClawback, 0x40},
		{"CannotEnableCanHoldConfidentialBalance", (*MPTokenIssuance).SetLsmfMPTCannotEnableCanHoldConfidentialBalance, 0x80},
		{"CanMutateMetadata", (*MPTokenIssuance).SetLsmfMPTCanMutateMetadata, 0x10000},
		{"CanMutateTransferFee", (*MPTokenIssuance).SetLsmfMPTCanMutateTransferFee, 0x20000},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issuance := &MPTokenIssuance{MutableFlags: 0x80000000}
			test.set(issuance)
			require.Equal(t, uint32(0x80000000)|test.literal, issuance.MutableFlags)
		})
	}
}
