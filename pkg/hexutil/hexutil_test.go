package hexutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeToUpperHex(t *testing.T) {
	tt := []struct {
		name  string
		input []byte
		expected    string
	}{
		{
			name:     "pass - nil input returns empty string",
			input:    nil,
			expected: "",
		},
		{
			name:     "pass - empty input returns empty string",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "pass - single byte encodes correctly",
			input:    []byte{0xFF},
			expected: "FF",
		},
		{
			name:     "pass - multiple bytes encode correctly",
			input:    []byte{0xDE, 0xAD, 0xBE, 0xEF},
			expected: "DEADBEEF",
		},
		{
			name:     "pass - output is uppercase",
			input:    []byte{0xab, 0xcd, 0xef},
			expected: "ABCDEF",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := EncodeToUpperHex(tc.input)
			require.Equal(t, tc.expected, got)
		})
	}
}
