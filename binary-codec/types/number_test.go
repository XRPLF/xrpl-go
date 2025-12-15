package types

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/binary-codec/types/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNumber_FromJSON(t *testing.T) {
	tt := []struct {
		name         string
		input        any
		expectedErr  error
		validateFunc func(*testing.T, []byte) // Custom validation for successful cases
	}{
		{
			name:        "fail - value is not a string",
			input:       123,
			expectedErr: ErrInvalidNumber,
		},
		{
			name:        "fail - invalid format",
			input:       "invalid",
			expectedErr: ErrInvalidNumber,
		},
		{
			name:  "pass - empty string (treated as zero)",
			input: "",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
				// Check for zero special exponent (0x80000000 = -2147483648)
				require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 128, 0, 0, 0}, result)
			},
		},
		{
			name:  "pass - zero value",
			input: "0",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
				// Check for zero special exponent (0x80000000 = -2147483648)
				require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 128, 0, 0, 0}, result)
			},
		},
		{
			name:  "pass - positive integer",
			input: "1",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
				// Verify it's not zero
				require.NotEqual(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 128, 0, 0, 0}, result)
			},
		},
		{
			name:  "pass - negative integer",
			input: "-1",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
				// First byte should indicate negative (high bit set)
				require.True(t, result[0] >= 128, "Negative number should have high bit set in mantissa")
			},
		},
		{
			name:  "pass - decimal number",
			input: "123.456",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
			},
		},
		{
			name:  "pass - negative decimal",
			input: "-123.456",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
				require.True(t, result[0] >= 128, "Negative number should have high bit set")
			},
		},
		{
			name:  "pass - scientific notation positive exponent",
			input: "1.5e10",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
			},
		},
		{
			name:  "pass - scientific notation negative exponent",
			input: "1.5e-10",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
			},
		},
		{
			name:  "pass - very small number",
			input: "0.000000000000001",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
			},
		},
		{
			name:  "pass - number with plus sign",
			input: "+123.456",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
			},
		},
		{
			name:  "pass - number with capital E in exponent",
			input: "1.5E10",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
			},
		},
		{
			name:  "pass - decimal with trailing zero",
			input: "5.0",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
			},
		},
		{
			name:  "pass - decimal less than one",
			input: "0.5",
			validateFunc: func(t *testing.T, result []byte) {
				require.Len(t, result, 12, "Number should encode to 12 bytes")
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			n := &Number{}
			actual, err := n.FromJSON(tc.input)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				require.Nil(t, actual)
			} else {
				require.NoError(t, err)
				if tc.validateFunc != nil {
					tc.validateFunc(t, actual)
				}
			}
		})
	}
}

func TestNumber_ToJSON(t *testing.T) {
	defs := definitions.Get()

	tt := []struct {
		name         string
		input        []byte
		malleate     func(t *testing.T) interfaces.BinaryParser
		validateFunc func(*testing.T, any) // Custom validation
		expectedErr  error
	}{
		{
			name:  "fail - binary parser has no data",
			input: []byte{},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				parserMock := testutil.NewMockBinaryParser(gomock.NewController(t))
				parserMock.EXPECT().ReadBytes(gomock.Any()).Return([]byte{}, errors.New("binary parser has no data"))
				return parserMock
			},
			expectedErr: errors.New("binary parser has no data"),
		},
		{
			name:  "pass - zero value",
			input: []byte{0, 0, 0, 0, 0, 0, 0, 0, 128, 0, 0, 0}, // special exponent for zero
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0, 0, 0, 0, 0, 0, 0, 0, 128, 0, 0, 0}, defs)
			},
			validateFunc: func(t *testing.T, result any) {
				require.Equal(t, "0", result)
			},
			expectedErr: nil,
		},
		{
			name:  "pass - can decode positive number",
			input: []byte{0x00, 0x03, 0x8D, 0x7E, 0xA4, 0xC6, 0x80, 0x00, 0xFF, 0xFF, 0xFF, 0xF1},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0x00, 0x03, 0x8D, 0x7E, 0xA4, 0xC6, 0x80, 0x00, 0xFF, 0xFF, 0xFF, 0xF1}, defs)
			},
			validateFunc: func(t *testing.T, result any) {
				str, ok := result.(string)
				require.True(t, ok, "Result should be a string")
				require.NotEmpty(t, str, "Result should not be empty")
				require.NotEqual(t, "0", str, "Result should not be zero")
			},
			expectedErr: nil,
		},
		{
			name:  "pass - can decode negative number",
			input: []byte{0xFF, 0xFC, 0x72, 0x81, 0x5B, 0x39, 0x80, 0x00, 0xFF, 0xFF, 0xFF, 0xF1},
			malleate: func(t *testing.T) interfaces.BinaryParser {
				return serdes.NewBinaryParser([]byte{0xFF, 0xFC, 0x72, 0x81, 0x5B, 0x39, 0x80, 0x00, 0xFF, 0xFF, 0xFF, 0xF1}, defs)
			},
			validateFunc: func(t *testing.T, result any) {
				str, ok := result.(string)
				require.True(t, ok, "Result should be a string")
				require.NotEmpty(t, str, "Result should not be empty")
				require.True(t, str[0] == '-', "Result should be negative")
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			n := &Number{}
			parser := tc.malleate(t)
			actual, err := n.ToJSON(parser)
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				if tc.validateFunc != nil {
					tc.validateFunc(t, actual)
				}
			}
		})
	}
}

func TestNumber_RoundTrip(t *testing.T) {
	tt := []struct {
		name  string
		input string
	}{
		{
			name:  "zero",
			input: "0",
		},
		{
			name:  "positive integer",
			input: "1",
		},
		{
			name:  "negative integer",
			input: "-1",
		},
		{
			name:  "positive decimal",
			input: "123.456",
		},
		{
			name:  "negative decimal",
			input: "-123.456",
		},
		{
			name:  "scientific notation",
			input: "1.5e10",
		},
		{
			name:  "small number",
			input: "0.000000000000001",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			n := &Number{}

			// Encode
			encoded, err := n.FromJSON(tc.input)
			require.NoError(t, err)
			require.NotNil(t, encoded)
			require.Equal(t, 12, len(encoded), "Number should encode to 12 bytes")

			// Decode
			defs := definitions.Get()
			parser := serdes.NewBinaryParser(encoded, defs)
			decoded, err := n.ToJSON(parser)
			require.NoError(t, err)
			require.NotNil(t, decoded)

			// The decoded value might not exactly match the input due to normalization,
			// but re-encoding should produce the same bytes
			reencoded, err := n.FromJSON(decoded)
			require.NoError(t, err)
			require.Equal(t, encoded, reencoded, "Round-trip encoding should be stable")
		})
	}
}

func TestParseNumberString(t *testing.T) {
	tt := []struct {
		name             string
		input            string
		expectedMantissa int64
		expectedExponent int32
		expectedErr      error
	}{
		{
			name:             "pass - zero",
			input:            "0",
			expectedMantissa: 0,
			expectedExponent: numberZeroExponent,
			expectedErr:      nil,
		},
		{
			name:             "pass - simple integer",
			input:            "123",
			expectedMantissa: 123,
			expectedExponent: 0,
			expectedErr:      nil,
		},
		{
			name:             "pass - negative integer",
			input:            "-456",
			expectedMantissa: -456,
			expectedExponent: 0,
			expectedErr:      nil,
		},
		{
			name:             "pass - decimal number",
			input:            "123.456",
			expectedMantissa: 123456,
			expectedExponent: -3,
			expectedErr:      nil,
		},
		{
			name:             "pass - number with exponent",
			input:            "1.5e10",
			expectedMantissa: 15,
			expectedExponent: 9,
			expectedErr:      nil,
		},
		{
			name:             "pass - negative exponent",
			input:            "1.5e-5",
			expectedMantissa: 15,
			expectedExponent: -6,
			expectedErr:      nil,
		},
		{
			name:             "pass - with plus sign",
			input:            "+123",
			expectedMantissa: 123,
			expectedExponent: 0,
			expectedErr:      nil,
		},
		{
			name:             "pass - leading zeros in decimal",
			input:            "0.00123",
			expectedMantissa: 123,
			expectedExponent: -5,
			expectedErr:      nil,
		},
		{
			name:             "fail - invalid characters",
			input:            "12a34",
			expectedMantissa: 0,
			expectedExponent: 0,
			expectedErr:      ErrInvalidNumber,
		},
		{
			name:             "fail - multiple decimal points",
			input:            "12.34.56",
			expectedMantissa: 0,
			expectedExponent: 0,
			expectedErr:      ErrInvalidNumber,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mantissa, exponent, err := parseNumberString(tc.input)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedMantissa, mantissa)
				require.Equal(t, tc.expectedExponent, exponent)
			}
		})
	}
}

func TestNormalizeNumber(t *testing.T) {
	tt := []struct {
		name             string
		inputMantissa    int64
		inputExponent    int32
		expectedMantissa int64
		expectedExponent int32
		expectedErr      error
	}{
		{
			name:             "pass - zero",
			inputMantissa:    0,
			inputExponent:    0,
			expectedMantissa: 0,
			expectedExponent: numberZeroExponent,
			expectedErr:      nil,
		},
		{
			name:             "pass - already normalized",
			inputMantissa:    1000000000000000, // exactly at min mantissa
			inputExponent:    0,
			expectedMantissa: 1000000000000000,
			expectedExponent: 0,
			expectedErr:      nil,
		},
		{
			name:             "pass - needs scaling up",
			inputMantissa:    1,
			inputExponent:    0,
			expectedMantissa: 1000000000000000,
			expectedExponent: -15,
			expectedErr:      nil,
		},
		{
			name:             "pass - needs scaling down",
			inputMantissa:    10000000000000000, // above max mantissa
			inputExponent:    0,
			expectedMantissa: 1000000000000000,
			expectedExponent: 1,
			expectedErr:      nil,
		},
		{
			name:             "pass - negative mantissa",
			inputMantissa:    -1,
			inputExponent:    0,
			expectedMantissa: -1000000000000000,
			expectedExponent: -15,
			expectedErr:      nil,
		},
		{
			name:             "pass - large positive",
			inputMantissa:    9999999999999999, // at max mantissa
			inputExponent:    0,
			expectedMantissa: 9999999999999999,
			expectedExponent: 0,
			expectedErr:      nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mantissa, exponent, err := normalizeNumber(tc.inputMantissa, tc.inputExponent)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedMantissa, mantissa)
				require.Equal(t, tc.expectedExponent, exponent)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tt := []struct {
		name         string
		mantissa     int64
		exponent     int32
		validateFunc func(*testing.T, string)
	}{
		{
			name:     "zero value",
			mantissa: 0,
			exponent: numberZeroExponent,
			validateFunc: func(t *testing.T, result string) {
				require.Equal(t, "0", result)
			},
		},
		{
			name:     "normalized mantissa in decimal range",
			mantissa: 1000000000000000,
			exponent: -15,
			validateFunc: func(t *testing.T, result string) {
				require.NotEmpty(t, result)
				// Should produce a valid number string
				require.Contains(t, "0123456789.-e", string(result[0]))
			},
		},
		{
			name:     "large exponent uses scientific notation",
			mantissa: 1000000000000000,
			exponent: 10,
			validateFunc: func(t *testing.T, result string) {
				require.NotEmpty(t, result)
				require.Contains(t, result, "e")
			},
		},
		{
			name:     "negative mantissa",
			mantissa: -1000000000000000,
			exponent: -15,
			validateFunc: func(t *testing.T, result string) {
				require.NotEmpty(t, result)
				require.True(t, result[0] == '-', "Should start with minus sign")
			},
		},
		{
			name:     "scientific notation format",
			mantissa: 5000000000000000,
			exponent: 0,
			validateFunc: func(t *testing.T, result string) {
				require.NotEmpty(t, result)
				require.Contains(t, result, "e")
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := formatNumber(tc.mantissa, tc.exponent)
			tc.validateFunc(t, result)
		})
	}
}
