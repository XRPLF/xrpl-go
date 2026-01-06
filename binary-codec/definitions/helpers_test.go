package definitions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTypeNameByFieldName(t *testing.T) {

	tt := []struct {
		description   string
		input         string
		expected      string
		expectedError error
	}{
		{
			description:   "test that `TransferRate` gives `UInt32`",
			input:         "TransferRate",
			expected:      "UInt32",
			expectedError: nil,
		},
		{
			description: "test that invalid value gives an error",
			input:       "yurt",
			expected:    "",
			expectedError: &NotFoundError{
				Instance: "FieldName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {

		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetTypeNameByFieldName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})

	}

}

func TestGetTypeCodeByTypeName(t *testing.T) {

	tt := []struct {
		description   string
		input         string
		expected      int32
		expectedError error
	}{
		{
			description:   "test that `Done` gives correct code",
			input:         "Done",
			expected:      -1,
			expectedError: nil,
		},
		{
			description:   "test that `Hash128` gives correct code",
			input:         "Hash128",
			expected:      4,
			expectedError: nil,
		},
		{
			description: "test that incorrect value gives an error",
			input:       "yurt",
			expected:    0,
			expectedError: &NotFoundError{
				Instance: "TypeName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {

		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetTypeCodeByTypeName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})

	}

}

func TestGetTypeCodeByFieldName(t *testing.T) {
	tt := []struct {
		description   string
		input         string
		expected      int32
		expectedError error
	}{
		{
			description:   "test that `TransferRate` gives 2",
			input:         "TransferRate",
			expected:      2,
			expectedError: nil,
		},
		{
			description:   "test that `OwnerNode` gives 3",
			input:         "OwnerNode",
			expected:      3,
			expectedError: nil,
		},
		{
			description: "test that non-existent value gives error",
			input:       "yurt",
			expected:    0,
			expectedError: &NotFoundError{
				Instance: "FieldName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {

		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetTypeCodeByFieldName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetFieldCodeByFieldName(t *testing.T) {

	tt := []struct {
		description   string
		input         string
		expected      int32
		expectedError error
	}{
		{
			description:   "correct FieldCode",
			input:         "TransferRate",
			expected:      11,
			expectedError: nil,
		},
		{
			description: "Invalid FieldName",
			input:       "yurt",
			expected:    0,
			expectedError: &NotFoundError{
				Instance: "FieldName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {

		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetFieldCodeByFieldName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetFieldHeaderByFieldName(t *testing.T) {
	tt := []struct {
		description   string
		input         string
		expected      *FieldHeader
		expectedError error
	}{
		{
			description: "correct FieldHeader",
			input:       "TransferRate",
			expected: &FieldHeader{
				TypeCode:  2,
				FieldCode: 11,
			},
			expectedError: nil,
		},
		{
			description: "Invalid FieldName",
			input:       "yurt",
			expected:    nil,
			expectedError: &NotFoundError{
				Instance: "FieldName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {

		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetFieldHeaderByFieldName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetFieldNameByFieldHeader(t *testing.T) {
	tt := []struct {
		description   string
		input         FieldHeader
		expected      string
		expectedError error
	}{
		{
			description: "correct fieldName",
			input: FieldHeader{
				TypeCode:  1,
				FieldCode: 1,
			},
			expected:      "LedgerEntryType",
			expectedError: nil,
		},
		{
			description: "correct fieldName 2",
			input: FieldHeader{
				TypeCode:  5,
				FieldCode: 21,
			},
			expected:      "Digest",
			expectedError: nil,
		},
		{
			description: "invalid FieldHeader",
			input: FieldHeader{
				TypeCode:  0000000000000111,
				FieldCode: 000000000000111,
			},
			expected: "",
			expectedError: &NotFoundErrorFieldHeader{
				Instance: "FieldHeader",
				Input: FieldHeader{
					TypeCode:  0000000000000111,
					FieldCode: 000000000000111,
				},
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetFieldNameByFieldHeader(test.input)
			if test.expectedError != nil {
				require.Error(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetFieldInfoByFieldName(t *testing.T) {
	tt := []struct {
		description   string
		input         string
		expected      *FieldInfo
		expectedError error
	}{
		{
			description: "correct FieldInfo",
			input:       "TransferRate",
			expected: &FieldInfo{
				Nth:            11,
				IsVLEncoded:    false,
				IsSerialized:   true,
				IsSigningField: true,
				Type:           "UInt32",
			},
			expectedError: nil,
		},
		{
			description: "invalid FieldInfo",
			input:       "yurt",
			expected:    nil,
			expectedError: &NotFoundError{
				Instance: "FieldName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetFieldInfoByFieldName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}

		})
	}
}

func TestGetFieldInstanceByFieldName(t *testing.T) {
	tt := []struct {
		description   string
		input         string
		expected      *FieldInstance
		expectedError error
	}{
		{
			description: "correct FieldInstance",
			input:       "TransferRate",
			expected: &FieldInstance{
				FieldName: "TransferRate",
				FieldInfo: &FieldInfo{
					Nth:            11,
					IsVLEncoded:    false,
					IsSerialized:   true,
					IsSigningField: true,
					Type:           "UInt32",
				},
				FieldHeader: &FieldHeader{
					TypeCode:  2,
					FieldCode: 11,
				},
				Ordinal: 131083,
			},
		},
		{
			description: "invalid FieldName",
			input:       "yurt",
			expected:    nil,
			expectedError: &NotFoundError{
				Instance: "FieldName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {

		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetFieldInstanceByFieldName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetTransactionTypeCodeByTransactionTypeName(t *testing.T) {
	tt := []struct {
		description   string
		input         string
		expected      int32
		expectedError error
	}{
		{
			description:   "correct TransactionTypeCode",
			input:         "EscrowCreate",
			expected:      1,
			expectedError: nil,
		},
		{
			description: "invalid TransactionTypeName",
			input:       "yurt",
			expected:    0,
			expectedError: &NotFoundError{
				Instance: "TransactionTypeName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetTransactionTypeCodeByTransactionTypeName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetTransactionTypeNameByTransactionTypeCode(t *testing.T) {
	tt := []struct {
		description   string
		input         int32
		expected      string
		expectedError error
	}{
		{
			description:   "correct TypeName",
			input:         1,
			expected:      "EscrowCreate",
			expectedError: nil,
		},
		{
			description: "invalid TransactionTypeCode",
			input:       999999999,
			expected:    "",
			expectedError: &NotFoundErrorInt{
				Instance: "TransactionTypeCode",
				Input:    999999999,
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetTransactionTypeNameByTransactionTypeCode(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetTransactionResultNameByTransactionResultTypeCode(t *testing.T) {
	tt := []struct {
		description   string
		input         int32
		expected      string
		expectedError error
	}{
		{
			description:   "correct TransactionResultName",
			input:         100,
			expected:      "tecCLAIM",
			expectedError: nil,
		},
		{
			description: "invalid txResultTypeCode",
			input:       999999999,
			expected:    "",
			expectedError: &NotFoundErrorInt{
				Instance: "TransactionResultTypeCode",
				Input:    999999999,
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetTransactionResultNameByTransactionResultTypeCode(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetTransactionResultTypeCodeByTransactionResultName(t *testing.T) {
	tt := []struct {
		description   string
		input         string
		expected      int32
		expectedError error
	}{
		{
			description:   "correct TransactionResultTypeCode",
			input:         "tecCLAIM",
			expected:      100,
			expectedError: nil,
		},
		{
			description: "invalid TransactionResultName",
			input:       "yurt",
			expected:    0,
			expectedError: &NotFoundError{
				Instance: "TransactionResultName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetTransactionResultTypeCodeByTransactionResultName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetLedgerEntryTypeCodeByLedgerEntryTypeName(t *testing.T) {
	tt := []struct {
		description   string
		input         string
		expected      int32
		expectedError error
	}{
		{
			description:   "correct LedgerEntryTypeCode",
			input:         "AccountRoot",
			expected:      97,
			expectedError: nil,
		},
		{
			description: "invalid LedgerEntryTypeName",
			input:       "yurt",
			expected:    0,
			expectedError: &NotFoundError{
				Instance: "LedgerEntryTypeName",
				Input:    "yurt",
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetLedgerEntryTypeCodeByLedgerEntryTypeName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}

}

func TestGetLedgerEntryTypeNameByLedgerEntryTypeCode(t *testing.T) {
	tt := []struct {
		description   string
		input         int32
		expected      string
		expectedError error
	}{
		{
			description:   "correct LedgerEntryTypeName",
			input:         97,
			expected:      "AccountRoot",
			expectedError: nil,
		},
		{
			description: "invalid LedgerEntryTypeCode",
			input:       999999999,
			expected:    "",
			expectedError: &NotFoundErrorInt{
				Instance: "LedgerEntryTypeCode",
				Input:    999999999,
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetLedgerEntryTypeNameByLedgerEntryTypeCode(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetDelegatablePermissionValueByName(t *testing.T) {
	tt := []struct {
		description   string
		input         string
		expected      int32
		expectedError error
	}{
		{
			description:   "pass - valid granular permission",
			input:         "TrustlineAuthorize",
			expected:      65537,
			expectedError: nil,
		},
		{
			description:   "pass - valid transaction type permission",
			input:         "Payment",
			expected:      1,
			expectedError: nil,
		},
		{
			description: "fail - invalid permission name",
			input:       "InvalidPermission",
			expected:    0,
			expectedError: &NotFoundError{
				Instance: "DelegatablePermissionName",
				Input:    "InvalidPermission",
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetDelegatablePermissionValueByName(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestGetDelegatablePermissionNameByValue(t *testing.T) {
	tt := []struct {
		description   string
		input         int32
		expected      string
		expectedError error
	}{
		{
			description:   "valid granular permission value",
			input:         65537,
			expected:      "TrustlineAuthorize",
			expectedError: nil,
		},
		{
			description:   "valid transaction type permission value",
			input:         1,
			expected:      "Payment",
			expectedError: nil,
		},
		{
			description: "invalid permission value",
			input:       99999,
			expected:    "",
			expectedError: &NotFoundErrorInt{
				Instance: "DelegatablePermissionValue",
				Input:    99999,
			},
		},
	}

	for _, test := range tt {
		t.Run(test.description, func(t *testing.T) {
			got, err := definitions.GetDelegatablePermissionNameByValue(test.input)
			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
				require.Zero(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}
