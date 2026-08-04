package types

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMPTAmount(t *testing.T) {
	amount := MPTAmount(10000)

	require.Equal(t, uint64(10000), amount.Uint64())
	require.Equal(t, "10000", amount.String())
	require.Equal(t, "10000", amount.Flatten())
	require.False(t, amount.IsZero())
	require.True(t, amount.IsValid())
	require.True(t, MPTAmount(0).IsZero())
	require.False(t, MPTAmount(math.MaxInt64+1).IsValid())
}

func TestMPTAmountMarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		amount      MPTAmount
		expected    string
		expectedErr error
	}{
		{
			name:     "maximum",
			amount:   MaxMPTAmount,
			expected: `"9223372036854775807"`,
		},
		{
			name:        "out of range",
			amount:      MPTAmount(math.MaxInt64 + 1),
			expectedErr: ErrInvalidMPTAmount,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.amount)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			require.JSONEq(t, test.expected, string(encoded))
		})
	}
}

func TestMPTAmountUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    MPTAmount
		expectedErr error
	}{
		{name: "zero", input: `"0"`, expected: MPTAmount(0)},
		{name: "one", input: `"1"`, expected: MPTAmount(1)},
		{name: "amount", input: `"10000"`, expected: MPTAmount(10000)},
		{name: "maximum", input: `"9223372036854775807"`, expected: MaxMPTAmount},
		{name: "unquoted number", input: `1`, expectedErr: ErrInvalidMPTAmount},
		{name: "empty string", input: `""`, expectedErr: ErrInvalidMPTAmount},
		{name: "negative", input: `"-1"`, expectedErr: ErrInvalidMPTAmount},
		{name: "explicit plus", input: `"+1"`, expectedErr: ErrInvalidMPTAmount},
		{name: "decimal", input: `"1.0"`, expectedErr: ErrInvalidMPTAmount},
		{name: "hexadecimal", input: `"0x10"`, expectedErr: ErrInvalidMPTAmount},
		{name: "out of range", input: `"9223372036854775808"`, expectedErr: ErrInvalidMPTAmount},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var amount MPTAmount
			err := json.Unmarshal([]byte(test.input), &amount)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.expected, amount)
		})
	}
}
