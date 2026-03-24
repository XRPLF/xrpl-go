package serdes

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/stretchr/testify/require"
)

func TestFieldIDCodec_Encode(t *testing.T) {
	tt := []struct {
		description string
		input       string
		expected    []byte
		expectedErr error
	}{
		{
			description: "Type Code and Field Code < 16",
			input:       "Sequence",
			expected:    []byte{36},
			expectedErr: nil,
		},
		{
			description: "Additional Type Code and Field Code < 16",
			input:       "Flags",
			expected:    []byte{34},
			expectedErr: nil,
		},
		{
			description: "Additional Type Code and Field Code < 16",
			input:       "DestinationTag",
			expected:    []byte{46},
			expectedErr: nil,
		},
		{
			description: "Type Code >= 16 and Field Code < 16",
			input:       "Paths",
			expected:    []byte{1, 18},
			expectedErr: nil,
		},
		{
			description: "Additional Type Code >= 16 and Field Code < 16",
			input:       "CloseResolution",
			expected:    []byte{1, 16},
			expectedErr: nil,
		},
		{
			description: "Type Code < 16 and Field Code >= 16",
			input:       "SetFlag",
			expected:    []byte{32, 33},
			expectedErr: nil,
		},
		{
			description: "Additional Type Code < 16 and Field Code >= 16",
			input:       "Nickname",
			expected:    []byte{80, 18},
			expectedErr: nil,
		},
		{
			description: "Type Code and Field Code >= 16",
			input:       "TickSize",
			expected:    []byte{0, 16, 16},
			expectedErr: nil,
		},
		{
			description: "Additional Type Code and Field Code >= 16",
			input:       "UNLModifyDisabling",
			expected:    []byte{0, 16, 17},
			expectedErr: nil,
		},
		{
			description: "Non existent field name",
			input:       "yurt",
			expected:    nil,
			expectedErr: &definitions.NotFoundError{Instance: "FieldName", Input: "yurt"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			got, err := NewFieldIDCodec(definitions.Get()).Encode(tc.input)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestFieldIDCodec_Decode(t *testing.T) {
	tt := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:        "pass -Decode Sequence fieldId (Type Code and Field Code < 16)",
			input:       "24",
			expected:    "Sequence",
			expectedErr: nil,
		},
		{
			name:        "pass - Decode DestinationTag fieldId (Type Code and Field Code < 16)",
			input:       "2e",
			expected:    "DestinationTag",
			expectedErr: nil,
		},
		{
			name:        "pass - Decode Paths fieldId (Type Code >= 16 and Field Code < 16)",
			input:       "0112",
			expected:    "Paths",
			expectedErr: nil,
		},
		{
			name:        "pass - Decode CloseResolution fieldId (Type Code >= 16 and Field Code < 16)",
			input:       "0110",
			expected:    "CloseResolution",
			expectedErr: nil,
		},
		{
			name:        "pass - Decode SetFlag fieldId (Type Code < 16 and Field Code >= 16)",
			input:       "2021",
			expected:    "SetFlag",
			expectedErr: nil,
		},
		{
			name:        "pass - Decode Nickname fieldId (Type Code < 16 and Field Code >= 16)",
			input:       "5012",
			expected:    "Nickname",
			expectedErr: nil,
		},
		{
			name:        "pass - Decode TickSize fieldId (Type Code and Field Code >= 16)",
			input:       "001010",
			expected:    "TickSize",
			expectedErr: nil,
		},
		{
			name:        "pass - Decode UNLModifyDisabling fieldId (Type Code and Field Code >= 16)",
			input:       "001011",
			expected:    "UNLModifyDisabling",
			expectedErr: nil,
		},
		{
			name:        "fail - Non existent field name",
			input:       "ff",
			expected:    "",
			expectedErr: &definitions.NotFoundErrorFieldHeader{Instance: "FieldHeader"},
		},
		{
			name:        "fail - Invalid hex string",
			input:       "zz",
			expected:    "",
			expectedErr: errors.New("encoding/hex: invalid byte: U+007A"),
		},
		{
			name:        "fail - Invalid field ID length",
			input:       "ffffffff",
			expected:    "",
			expectedErr: ErrInvalidFieldIDLength,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewFieldIDCodec(definitions.Get()).Decode(tc.input)

			if tc.expectedErr != nil {
				require.Error(t, err, tc.expectedErr.Error())
				require.Empty(t, actual)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}
