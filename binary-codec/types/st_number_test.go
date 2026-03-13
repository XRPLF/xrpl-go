package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/stretchr/testify/require"
)

func TestNumber_FromJSON_ToJSON_Roundtrip(t *testing.T) {
	tt := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Positive normal value",
			input:    "99",
			expected: "99",
		},
		{
			name:     "Positive very large value",
			input:    "100000000000",
			expected: "1e11",
		},
		{
			name:     "Positive large value",
			input:    "10000000000",
			expected: "10000000000",
		},
		{
			name:     "Negative normal value",
			input:    "-123",
			expected: "-123",
		},
		{
			name:     "Negative very large value",
			input:    "-100000000000",
			expected: "-1e11",
		},
		{
			name:     "Negative large value",
			input:    "-10000000000",
			expected: "-10000000000",
		},
		{
			name:     "Positive very small value",
			input:    "0.00000000001",
			expected: "1e-11",
		},
		{
			name:     "Positive small value",
			input:    "0.0001",
			expected: "0.0001",
		},
		{
			name:     "Zero roundtrip",
			input:    "0",
			expected: "0",
		},
		{
			name:     "Roundtrip scientific positive",
			input:    "1.23e5",
			expected: "123000",
		},
		{
			name:     "Roundtrip scientific negative",
			input:    "-4.56e-7",
			expected: "-0.000000456",
		},
		{
			name:     "Negative medium value",
			input:    "-987654321",
			expected: "-987654321",
		},
		{
			name:     "Positive medium value",
			input:    "987654321",
			expected: "987654321",
		},
		{
			name:     "Decimal without exponent",
			input:    "0.5",
			expected: "0.5",
		},
		{
			name:     "Rounds up mantissa (last digit >= 5)",
			input:    "9223372036854775895",
			expected: "9223372036854775900",
		},
		{
			name:     "Rounds down mantissa (last digit < 5)",
			input:    "9323372036854775804",
			expected: "9323372036854775800",
		},
		{
			name:     "Small value with trailing zeros",
			input:    "0.002500",
			expected: "0.0025",
		},
		{
			name:     "Large value with trailing zeros",
			input:    "9900000000000000000000",
			expected: "99e20",
		},
		{
			name:     "Small value with leading zeros",
			input:    "0.0000000000000000000099",
			expected: "99e-22",
		},
		{
			name:     "Mantissa greater than MAX_MANTISSA",
			input:    "9999999999999999999999",
			expected: "1e22",
		},
		{
			name:     "Mantissa greater than MAX_INT64",
			input:    "92233720368547758079",
			expected: "922337203685477581e2",
		},
	}

	defs := definitions.Get()

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			n := &Number{}

			// FromJSON
			encoded, err := n.FromJSON(tc.input)
			require.NoError(t, err)
			require.Len(t, encoded, 12, "encoded bytes should be 12 bytes")

			// ToJSON
			parser := serdes.NewBinaryParser(encoded, defs)
			result, err := n.ToJSON(parser)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestNumber_FromJSON_Errors(t *testing.T) {
	tt := []struct {
		name  string
		input any
	}{
		{
			name:  "Exponent overflow",
			input: "1e40000",
		},
		{
			name:  "Underflow - value too small",
			input: "1e-40000",
		},
		{
			name:  "Invalid input - letters",
			input: "abc123",
		},
		{
			name:  "Invalid input - empty string",
			input: "",
		},
		{
			name:  "Invalid type - not a string",
			input: 12345,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			n := &Number{}
			result, err := n.FromJSON(tc.input)
			require.Error(t, err)
			require.Nil(t, result)
		})
	}
}

func TestNumber_HexRoundtrip(t *testing.T) {
	defs := definitions.Get()
	n := &Number{}

	encoded, err := n.FromJSON("42.7")
	require.NoError(t, err)

	// Convert to hex and back
	hexStr := hex.EncodeToString(encoded)
	decoded, err := hex.DecodeString(hexStr)
	require.NoError(t, err)

	// Parse the decoded bytes
	parser := serdes.NewBinaryParser(decoded, defs)
	result, err := n.ToJSON(parser)
	require.NoError(t, err)
	require.Equal(t, "42.7", result)
}

func TestNumber_ToJSON_Zero(t *testing.T) {
	defs := definitions.Get()
	n := &Number{}

	// Construct zero bytes manually: mantissa=0 (8 bytes), exponent=0x80000000 (4 bytes)
	buf := make([]byte, 12)
	writeInt64BE(buf, 0, 0)
	writeInt32BE(buf, defaultZeroExp, 8)

	parser := serdes.NewBinaryParser(buf, defs)
	result, err := n.ToJSON(parser)
	require.NoError(t, err)
	require.Equal(t, "0", result)
}

func TestNumber_ToJSON_ParserError(t *testing.T) {
	defs := definitions.Get()
	n := &Number{}

	// Parser with insufficient data (only 4 bytes, needs 12)
	parser := serdes.NewBinaryParser([]byte{0, 0, 0, 0}, defs)
	_, err := n.ToJSON(parser)
	require.Error(t, err)
}

func TestNumber_RoundTrip_Values(t *testing.T) {
	// Test a variety of values for encode -> decode consistency
	values := []string{
		"1",
		"-1",
		"0",
		"100",
		"-100",
		"0.001",
		"-0.001",
		"1000000000000000000",
		"-1000000000000000000",
		"1.5",
		"-1.5",
		"999999999999999999",
		"1e10",
		"1e-10",
	}

	defs := definitions.Get()

	for _, v := range values {
		t.Run(fmt.Sprintf("RoundTrip_%s", v), func(t *testing.T) {
			n := &Number{}

			encoded, err := n.FromJSON(v)
			require.NoError(t, err)
			require.Len(t, encoded, 12)

			parser := serdes.NewBinaryParser(encoded, defs)
			result, err := n.ToJSON(parser)
			require.NoError(t, err)
			require.NotEmpty(t, result)
		})
	}
}
