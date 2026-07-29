package definitions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfidentialMPTWireFieldDefinitions(t *testing.T) {
	tests := []struct {
		name    string
		info    FieldInfo
		header  FieldHeader
		ordinal int32
	}{
		{
			name: "BlindingFactor",
			info: FieldInfo{
				Nth:            40,
				IsVLEncoded:    false,
				IsSerialized:   true,
				IsSigningField: true,
				Type:           "Hash256",
			},
			header:  FieldHeader{TypeCode: 5, FieldCode: 40},
			ordinal: 327720,
		},
		{
			name: "AmountCommitment",
			info: FieldInfo{
				Nth:            45,
				IsVLEncoded:    true,
				IsSerialized:   true,
				IsSigningField: true,
				Type:           "Blob",
			},
			header:  FieldHeader{TypeCode: 7, FieldCode: 45},
			ordinal: 458797,
		},
		{
			name: "BalanceCommitment",
			info: FieldInfo{
				Nth:            46,
				IsVLEncoded:    true,
				IsSerialized:   true,
				IsSigningField: true,
				Type:           "Blob",
			},
			header:  FieldHeader{TypeCode: 7, FieldCode: 46},
			ordinal: 458798,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			field, err := Get().GetFieldInstanceByFieldName(tc.name)
			require.NoError(t, err)
			require.Equal(t, tc.info, *field.FieldInfo)
			require.Equal(t, tc.header, *field.FieldHeader)
			require.Equal(t, tc.ordinal, field.Ordinal)
		})
	}
}

func TestConfidentialMPTTransactionResultCodes(t *testing.T) {
	tests := []struct {
		name string
		code int32
	}{
		{name: "tecBAD_PROOF", code: 199},
		{name: "temBAD_CIPHERTEXT", code: -248},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, err := Get().GetTransactionResultTypeCodeByTransactionResultName(tc.name)
			require.NoError(t, err)
			require.Equal(t, tc.code, code)
		})
	}
}

func TestConfidentialMPTTransactionTypeCodes(t *testing.T) {
	tests := []struct {
		name string
		code int32
	}{
		{name: "ConfidentialMPTConvert", code: 85},
		{name: "ConfidentialMPTMergeInbox", code: 86},
		{name: "ConfidentialMPTConvertBack", code: 87},
		{name: "ConfidentialMPTSend", code: 88},
		{name: "ConfidentialMPTClawback", code: 89},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, err := Get().GetTransactionTypeCodeByTransactionTypeName(tc.name)
			require.NoError(t, err)
			require.Equal(t, tc.code, code)
		})
	}
}
