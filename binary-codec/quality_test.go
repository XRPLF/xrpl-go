package binarycodec

import (
	"testing"

	bigdecimal "github.com/Peersyst/xrpl-go/pkg/big-decimal"
	"github.com/stretchr/testify/require"
)

func TestQualityCodec_Encode(t *testing.T) {
	testcases := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:        "fail - invalid quality - empty string",
			input:       "",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - invalid character",
			input:       "invalid",
			expectedErr: bigdecimal.ErrInvalidCharacter{Allowed: bigdecimal.AllowedCharacters},
		},
		{
			name:        "fail - invalid quality - overflow",
			input:       "195796912.51716641",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - overflow",
			input:       "1195796912.5171664",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:     "pass - valid zero quality",
			input:    "0",
			expected: "5500000000000000",
		},
		{
			name:     "pass - valid quality with decimal",
			input:    "0.0",
			expected: "5500000000000000",
		},
		{
			name:     "pass - valid quality with decimal - leading dot",
			input:    "0.",
			expected: "5500000000000000",
		},
		{
			name:     "pass - valid quality with decimal - trailing dot",
			input:    ".0",
			expected: "5500000000000000",
		},
		{
			name:     "pass - valid negative quality",
			input:    "-195796912.5171664",
			expected: "5D06F4C3362FE1D0",
		},
		{
			name:     "pass - valid quality - non decimal",
			input:    "195796912",
			expected: "640000000BAB9FB0",
		},
		{
			name:     "pass - valid quality",
			input:    "195796912.5171664",
			expected: "5D06F4C3362FE1D0",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeQuality(tc.input)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, encoded)
			}
		})
	}
}

func TestQualityCodec_Decode(t *testing.T) {
	testcases := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:        "fail - invalid quality - empty string",
			input:       "",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - malformed hex",
			input:       "GG00000000000000",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - odd-length hex",
			input:       "550000000000000",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - one decoded byte",
			input:       "00",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - two decoded bytes",
			input:       "0000",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - three decoded bytes",
			input:       "000000",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - four decoded bytes",
			input:       "00000000",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - five decoded bytes",
			input:       "0000000000",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - six decoded bytes",
			input:       "000000000000",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - invalid quality - seven decoded bytes",
			input:       "00000000000000",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:     "pass - exact eight decoded bytes - zero quality",
			input:    "5500000000000000",
			expected: "0",
		},
		{
			name:     "pass - longer input uses final eight decoded bytes",
			input:    "FF5500000000000000",
			expected: "0",
		},
		{
			name:     "pass - valid quality",
			input:    "5D06F4C3362FE1D0",
			expected: "195796912.5171664",
		},
		{
			name:     "pass - valid quality - non decimal",
			input:    "640000000BAB9FB0",
			expected: "195796912",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, err := DecodeQuality(tc.input)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, decoded)
			}
		})
	}
}
