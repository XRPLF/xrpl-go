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

func TestMPTokenIssuance_SetLsmfMPTCanMutateCanLock(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsmfMPTCanMutateCanLock()
	require.Equal(t, LsmfMPTCanMutateCanLock, mpTokenIssuance.MutableFlags)
}

func TestMPTokenIssuance_SetLsmfMPTCanMutateRequireAuth(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsmfMPTCanMutateRequireAuth()
	require.Equal(t, LsmfMPTCanMutateRequireAuth, mpTokenIssuance.MutableFlags)
}

func TestMPTokenIssuance_SetLsmfMPTCanMutateCanEscrow(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsmfMPTCanMutateCanEscrow()
	require.Equal(t, LsmfMPTCanMutateCanEscrow, mpTokenIssuance.MutableFlags)
}

func TestMPTokenIssuance_SetLsmfMPTCanMutateCanTrade(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsmfMPTCanMutateCanTrade()
	require.Equal(t, LsmfMPTCanMutateCanTrade, mpTokenIssuance.MutableFlags)
}

func TestMPTokenIssuance_SetLsmfMPTCanMutateCanTransfer(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsmfMPTCanMutateCanTransfer()
	require.Equal(t, LsmfMPTCanMutateCanTransfer, mpTokenIssuance.MutableFlags)
}

func TestMPTokenIssuance_SetLsmfMPTCanMutateCanClawback(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsmfMPTCanMutateCanClawback()
	require.Equal(t, LsmfMPTCanMutateCanClawback, mpTokenIssuance.MutableFlags)
}

func TestMPTokenIssuance_SetLsmfMPTCanMutateMetadata(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsmfMPTCanMutateMetadata()
	require.Equal(t, LsmfMPTCanMutateMetadata, mpTokenIssuance.MutableFlags)
}

func TestMPTokenIssuance_SetLsmfMPTCanMutateTransferFee(t *testing.T) {
	mpTokenIssuance := &MPTokenIssuance{}
	mpTokenIssuance.SetLsmfMPTCanMutateTransferFee()
	require.Equal(t, LsmfMPTCanMutateTransferFee, mpTokenIssuance.MutableFlags)
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
				MutableFlags:      LsmfMPTCanMutateCanLock | LsmfMPTCanMutateMetadata,
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := testutil.SerializeAndDeserialize(t, test.mpTokenIssuance, test.expected); err != nil {
				t.Error(err)
			}
		})
	}
}
