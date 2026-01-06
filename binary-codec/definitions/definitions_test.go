package definitions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefinitions(t *testing.T) {
	loadDefinitions()
	require.Equal(t, int32(-1), definitions.Types["Done"])
	require.Equal(t, int32(4), definitions.Types["Hash128"])
	require.Equal(t, int32(97), definitions.LedgerEntryTypes["AccountRoot"])
	require.Equal(t, int32(-399), definitions.TransactionResults["telLOCAL_ERROR"])
	require.Equal(t, int32(1), definitions.TransactionTypes["EscrowCreate"])
	require.Equal(t, &FieldInfo{Nth: 0, IsVLEncoded: false, IsSerialized: false, IsSigningField: false, Type: "Unknown"}, definitions.Fields["Generic"].FieldInfo)
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
