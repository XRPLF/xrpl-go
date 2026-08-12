package definitions

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ugorji/go/codec"
)

//go:embed definitions.json
var definitionsDocBytes []byte

type formatField struct {
	Name        string `json:"name"`
	Optionality int    `json:"optionality"`
}

type formatDefinitionsDoc struct {
	LedgerEntryFormats map[string][]formatField `json:"LEDGER_ENTRY_FORMATS"`
	TransactionFormats map[string][]formatField `json:"TRANSACTION_FORMATS"`
	Hash               string                   `json:"hash"`
}

func TestLoadDefinitions(t *testing.T) {
	loadDefinitions()
	require.Equal(t, int32(-1), definitions.Types["Done"])
	require.Equal(t, int32(4), definitions.Types["Hash128"])
	require.Equal(t, int32(22), definitions.Types["Hash384"])
	require.Equal(t, int32(23), definitions.Types["Hash512"])
	require.Equal(t, int32(97), definitions.LedgerEntryTypes["AccountRoot"])
	require.Equal(t, int32(-399), definitions.TransactionResults["telLOCAL_ERROR"])
	require.Equal(t, int32(-249), definitions.TransactionResults["temBAD_MPT"])
	require.Equal(t, int32(-84), definitions.TransactionResults["terLOCKED"])
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
			name:    "ImmutableFlags",
			info:    &FieldInfo{Nth: 53, IsVLEncoded: false, IsSerialized: true, IsSigningField: true, Type: "UInt32"},
			header:  &FieldHeader{TypeCode: 2, FieldCode: 53},
			ordinal: 131125,
		},
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
		{
			name:    "BlindingFactor",
			info:    &FieldInfo{Nth: 40, IsVLEncoded: false, IsSerialized: true, IsSigningField: true, Type: "Hash256"},
			header:  &FieldHeader{TypeCode: 5, FieldCode: 40},
			ordinal: 327720,
		},
		{
			name:    "AmountCommitment",
			info:    &FieldInfo{Nth: 45, IsVLEncoded: true, IsSerialized: true, IsSigningField: true, Type: "Blob"},
			header:  &FieldHeader{TypeCode: 7, FieldCode: 45},
			ordinal: 458797,
		},
		{
			name:    "BalanceCommitment",
			info:    &FieldInfo{Nth: 46, IsVLEncoded: true, IsSerialized: true, IsSigningField: true, Type: "Blob"},
			header:  &FieldHeader{TypeCode: 7, FieldCode: 46},
			ordinal: 458798,
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

func TestConfidentialMPTFormats(t *testing.T) {
	var jsonHandle codec.JsonHandle
	jsonHandle.MapKeyAsString = true

	decoder := codec.NewDecoderBytes(definitionsDocBytes, &jsonHandle)
	var document formatDefinitionsDoc
	require.NoError(t, decoder.Decode(&document))
	require.Empty(t, document.Hash, "the embedded definitions contain an additional protocol overlay, so a source snapshot hash would be invalid")

	tests := []struct {
		name     string
		actual   []formatField
		expected []formatField
	}{
		{
			name:   "MPToken ledger entry",
			actual: document.LedgerEntryFormats["MPToken"],
			expected: []formatField{
				{Name: "Account", Optionality: 0},
				{Name: "MPTokenIssuanceID", Optionality: 0},
				{Name: "MPTAmount", Optionality: 2},
				{Name: "LockedAmount", Optionality: 1},
				{Name: "ConfidentialBalanceInbox", Optionality: 1},
				{Name: "ConfidentialBalanceSpending", Optionality: 1},
				{Name: "ConfidentialBalanceVersion", Optionality: 2},
				{Name: "IssuerEncryptedBalance", Optionality: 1},
				{Name: "AuditorEncryptedBalance", Optionality: 1},
				{Name: "HolderEncryptionKey", Optionality: 1},
				{Name: "OwnerNode", Optionality: 0},
				{Name: "PreviousTxnID", Optionality: 0},
				{Name: "PreviousTxnLgrSeq", Optionality: 0},
			},
		},
		{
			name:   "ConfidentialMPTConvert transaction",
			actual: document.TransactionFormats["ConfidentialMPTConvert"],
			expected: []formatField{
				{Name: "MPTokenIssuanceID", Optionality: 0},
				{Name: "MPTAmount", Optionality: 0},
				{Name: "HolderEncryptionKey", Optionality: 1},
				{Name: "HolderEncryptedAmount", Optionality: 0},
				{Name: "IssuerEncryptedAmount", Optionality: 0},
				{Name: "AuditorEncryptedAmount", Optionality: 1},
				{Name: "BlindingFactor", Optionality: 0},
				{Name: "ZKProof", Optionality: 1},
			},
		},
		{
			name:   "ConfidentialMPTMergeInbox transaction",
			actual: document.TransactionFormats["ConfidentialMPTMergeInbox"],
			expected: []formatField{
				{Name: "MPTokenIssuanceID", Optionality: 0},
			},
		},
		{
			name:   "ConfidentialMPTConvertBack transaction",
			actual: document.TransactionFormats["ConfidentialMPTConvertBack"],
			expected: []formatField{
				{Name: "MPTokenIssuanceID", Optionality: 0},
				{Name: "MPTAmount", Optionality: 0},
				{Name: "HolderEncryptedAmount", Optionality: 0},
				{Name: "IssuerEncryptedAmount", Optionality: 0},
				{Name: "AuditorEncryptedAmount", Optionality: 1},
				{Name: "BlindingFactor", Optionality: 0},
				{Name: "ZKProof", Optionality: 0},
				{Name: "BalanceCommitment", Optionality: 0},
			},
		},
		{
			name:   "ConfidentialMPTSend transaction",
			actual: document.TransactionFormats["ConfidentialMPTSend"],
			expected: []formatField{
				{Name: "MPTokenIssuanceID", Optionality: 0},
				{Name: "Destination", Optionality: 0},
				{Name: "DestinationTag", Optionality: 1},
				{Name: "SenderEncryptedAmount", Optionality: 0},
				{Name: "DestinationEncryptedAmount", Optionality: 0},
				{Name: "IssuerEncryptedAmount", Optionality: 0},
				{Name: "AuditorEncryptedAmount", Optionality: 1},
				{Name: "ZKProof", Optionality: 0},
				{Name: "AmountCommitment", Optionality: 0},
				{Name: "BalanceCommitment", Optionality: 0},
				{Name: "CredentialIDs", Optionality: 1},
			},
		},
		{
			name:   "ConfidentialMPTClawback transaction",
			actual: document.TransactionFormats["ConfidentialMPTClawback"],
			expected: []formatField{
				{Name: "MPTokenIssuanceID", Optionality: 0},
				{Name: "Holder", Optionality: 0},
				{Name: "MPTAmount", Optionality: 0},
				{Name: "ZKProof", Optionality: 0},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.actual)
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
