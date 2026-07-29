package definitions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefinitions(t *testing.T) {
	loadDefinitions()
	require.Equal(t, int32(-1), definitions.Types["Done"])
	require.Equal(t, int32(4), definitions.Types["Hash128"])
	require.Equal(t, int32(22), definitions.Types["Hash384"])
	require.Equal(t, int32(23), definitions.Types["Hash512"])
	require.Equal(t, int32(22), definitions.Types["UInt384"])
	require.Equal(t, int32(23), definitions.Types["UInt512"])
	require.Equal(t, int32(97), definitions.LedgerEntryTypes["AccountRoot"])
	require.Equal(t, int32(-399), definitions.TransactionResults["telLOCAL_ERROR"])
	require.Equal(t, int32(-249), definitions.TransactionResults["temBAD_MPT"])
	require.Equal(t, int32(-84), definitions.TransactionResults["terLOCKED"])
	require.Equal(t, int32(153), definitions.TransactionResults["tecHOOK_REJECTED"])
	require.Equal(t, int32(198), definitions.TransactionResults["tecNO_DELEGATE_PERMISSION"])
	require.Equal(t, int32(1), definitions.TransactionTypes["EscrowCreate"])
	require.Equal(t, &FieldInfo{Nth: 0, IsVLEncoded: false, IsSerialized: true, IsSigningField: true, Type: "Unknown"}, definitions.Fields["Generic"].FieldInfo)
	require.Equal(t, &FieldInfo{Nth: 28, IsVLEncoded: false, IsSerialized: true, IsSigningField: true, Type: "Hash256"}, definitions.Fields["NFTokenBuyOffer"].FieldInfo)
	require.Equal(t, &FieldInfo{Nth: 16, IsVLEncoded: false, IsSerialized: true, IsSigningField: true, Type: "UInt8"}, definitions.Fields["TickSize"].FieldInfo)
	require.Equal(t, &FieldHeader{TypeCode: 2, FieldCode: 4}, definitions.Fields["Sequence"].FieldHeader)
	require.Equal(t, &FieldHeader{TypeCode: 18, FieldCode: 1}, definitions.Fields["Paths"].FieldHeader)
	require.Equal(t, &FieldHeader{TypeCode: 2, FieldCode: 33}, definitions.Fields["SetFlag"].FieldHeader)
	require.Equal(t, &FieldHeader{TypeCode: 16, FieldCode: 16}, definitions.Fields["TickSize"].FieldHeader)
	require.Equal(t, "UInt32", definitions.Fields["TransferRate"].Type)
	require.Equal(t, "Sequence", definitions.FieldIDNameMap[FieldHeader{TypeCode: 2, FieldCode: 4}])
	require.Equal(t, "OfferSequence", definitions.FieldIDNameMap[FieldHeader{TypeCode: 2, FieldCode: 25}])
	require.Equal(t, "NFTokenSellOffer", definitions.FieldIDNameMap[FieldHeader{TypeCode: 5, FieldCode: 29}])
	require.Equal(t, int32(131076), definitions.Fields["Sequence"].Ordinal)
	require.Equal(t, int32(131097), definitions.Fields["OfferSequence"].Ordinal)
	require.Equal(t, int32(65537), definitions.GranularPermissions["TrustlineAuthorize"])
	require.Equal(t, int32(1), definitions.DelegatablePermissions["Payment"])

	mptFields := []struct {
		name    string
		info    *FieldInfo
		header  *FieldHeader
		ordinal int32
	}{
		{
			name:    "ReferenceHolding",
			info:    &FieldInfo{Nth: 39, IsVLEncoded: false, IsSerialized: true, IsSigningField: true, Type: "Hash256"},
			header:  &FieldHeader{TypeCode: 5, FieldCode: 39},
			ordinal: 327719,
		},
		{
			name:    "TakerPaysMPT",
			info:    &FieldInfo{Nth: 3, IsVLEncoded: false, IsSerialized: true, IsSigningField: true, Type: "Hash192"},
			header:  &FieldHeader{TypeCode: 21, FieldCode: 3},
			ordinal: 1376259,
		},
		{
			name:    "TakerGetsMPT",
			info:    &FieldInfo{Nth: 4, IsVLEncoded: false, IsSerialized: true, IsSigningField: true, Type: "Hash192"},
			header:  &FieldHeader{TypeCode: 21, FieldCode: 4},
			ordinal: 1376260,
		},
	}
	for _, field := range mptFields {
		t.Run(field.name, func(t *testing.T) {
			require.Equal(t, field.info, definitions.Fields[field.name].FieldInfo)
			require.Equal(t, field.header, definitions.Fields[field.name].FieldHeader)
			require.Equal(t, field.ordinal, definitions.Fields[field.name].Ordinal)
		})
	}
}

// Helper functions to create and test ordinals.
// func CreateOrdinal(fh FieldHeader) int32 {
// 	return fh.TypeCode<<16 | fh.FieldCode
// }

// func TestCreateOrdinal(t *testing.T) {
// 	tt := []struct {
// 		description string
// 		input       FieldHeader
// 	}{
// 		{
// 			description: "test ordinal creation",
// 			input:       FieldHeader{TypeCode: 2, FieldCode: 25},
// 		},
// 	}

// 	for _, tc := range tt {
// 		t.Run(tc.description, func(t *testing.T) {
// 			fmt.Println("Ordinal:", CreateOrdinal(tc.input))
// 		})
// 	}
// }

// nolint
func BenchmarkLoadDefinitions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		loadDefinitions()
	}
}

func TestGet(t *testing.T) {
	loadDefinitions()
	require.Equal(t, definitions, Get())
}
