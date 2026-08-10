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
			name:        "fail - invalid quality - incomplete exponent",
			input:       "1e",
			expectedErr: bigdecimal.ErrInvalidZeroValue,
		},
		{
			name:        "fail - malformed zero - lone dot",
			input:       ".",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - malformed zero - double dot",
			input:       "..",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - malformed zero - zeros around double dot",
			input:       "0..0",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - malformed zero - zeros around multiple dots",
			input:       "000...000",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - exponent out of int range",
			input:       "1e99999999999999999999",
			expectedErr: bigdecimal.ErrInvalidZeroValue,
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
			name:        "fail - non-canonical decoded quality",
			input:       "0.72057594037927935",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - normalized exponent underflow",
			input:       "1e-82",
			expectedErr: ErrInvalidQuality,
		},
		{
			name:        "fail - normalized exponent overflow",
			input:       "1e96",
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
			name:     "pass - valid quality with decimal - trailing dot",
			input:    "0.",
			expected: "5500000000000000",
		},
		{
			name:     "pass - valid quality with decimal - leading dot",
			input:    ".0",
			expected: "5500000000000000",
		},
		{
			name:     "pass - valid zero quality - negative zero",
			input:    "-0",
			expected: "5500000000000000",
		},
		{
			name:     "pass - valid zero quality - with exponent",
			input:    "0e5",
			expected: "5500000000000000",
		},
		{
			name:     "pass - valid zero quality - signed decimal with exponent",
			input:    "-0.00e99",
			expected: "5500000000000000",
		},
		{
			name:     "pass - valid negative quality",
			input:    "-195796912.5171664",
			expected: "5D06F4C3362FE1D0",
		},
		{
			name:     "pass - quality below one - leading fractional zero",
			input:    "0.05",
			expected: "5311C37937E08000",
		},
		{
			name:     "pass - quality below one - one mantissa digit",
			input:    "0.5",
			expected: "5411C37937E08000",
		},
		{
			name:     "pass - quality below one - two mantissa digits",
			input:    "0.55",
			expected: "54138A388A43C000",
		},
		{
			name:     "pass - quality below one - three mantissa digits",
			input:    "0.123",
			expected: "54045EADB112E000",
		},
		{
			name:     "pass - valid quality - non decimal",
			input:    "195796912",
			expected: "5D06F4C335E0F800",
		},
		{
			name:     "pass - valid quality",
			input:    "195796912.5171664",
			expected: "5D06F4C3362FE1D0",
		},
		{
			name:     "pass - minimum canonical exponent",
			input:    "1e-81",
			expected: "04038D7EA4C68000",
		},
		{
			name:     "pass - maximum canonical exponent",
			input:    "9999999999999999e80",
			expected: "B42386F26FC0FFFF",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeQuality(tc.input)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrInvalidQuality)
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, encoded)
			}
		})
	}
}

func TestQualityCodec_RoundTrip(t *testing.T) {
	qualities := []string{
		"5500000000000000",
		"5311C37937E08000",
		"5411C37937E08000",
		"54138A388A43C000",
		"54045EADB112E000",
		"5D06F4C335E0F800",
		"5D06F4C3362FE1D0",
		"04038D7EA4C68000",
		"B42386F26FC0FFFF",
	}

	for _, quality := range qualities {
		t.Run(quality, func(t *testing.T) {
			decoded, err := DecodeQuality(quality)
			require.NoError(t, err)

			encoded, err := EncodeQuality(decoded)
			require.NoError(t, err)
			require.Equal(t, quality, encoded)
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
			name:     "pass - non-canonical quality uses full mantissa",
			input:    "53FFFFFFFFFFFFFF",
			expected: "0.72057594037927935",
		},
		{
			name:     "pass - quality below one - leading fractional zero",
			input:    "5311C37937E08000",
			expected: "0.05",
		},
		{
			name:     "pass - quality below one - one mantissa digit",
			input:    "5411C37937E08000",
			expected: "0.5",
		},
		{
			name:     "pass - quality below one - two mantissa digits",
			input:    "54138A388A43C000",
			expected: "0.55",
		},
		{
			name:     "pass - quality below one - three mantissa digits",
			input:    "54045EADB112E000",
			expected: "0.123",
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
