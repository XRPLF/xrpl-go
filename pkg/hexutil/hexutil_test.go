package hexutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeToUpperHex(t *testing.T) {
	tt := []struct {
		name     string
		input    []byte
		expected string
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

func TestDecodeFixedHex(t *testing.T) {
	tt := []struct {
		name    string
		hex     string
		size    int
		wantErr bool
	}{
		{
			name:    "pass - valid 4 bytes",
			hex:     "deadbeef",
			size:    4,
			wantErr: false,
		},
		{
			name:    "pass - empty string with size 0",
			hex:     "",
			size:    0,
			wantErr: false,
		},
		{
			name:    "fail - wrong byte length",
			hex:     "0102",
			size:    32,
			wantErr: true,
		},
		{
			name:    "fail - invalid hex chars",
			hex:     "zzzz",
			size:    2,
			wantErr: true,
		},
		{
			name:    "fail - odd length hex",
			hex:     "012",
			size:    1,
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeFixedHex(tc.hex, tc.size)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, tc.size)
		})
	}
}
