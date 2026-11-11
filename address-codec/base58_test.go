package addresscodec

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeBase58(t *testing.T) {
	tt := []struct {
		name           string
		input          []byte
		expectedOutput string
	}{
		{
			name:           "pass - successful encode with XRP alphabet - 1",
			input:          []byte("rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN"),
			expectedOutput: "s2Fku4vaPpFiqqXdAD3V5rYrSx5a9h9qvUJW3423akZSCeD",
		},
		{
			name:           "pass -successful encode with XRP alphabet - 2",
			input:          []byte("rJrpjzcxwQxokkqPxm62o5rtNfe2XimrTr"),
			expectedOutput: "s2i2Jk6bF44eDSXnnMjxeVhnYZ3qmbteqesuhS6Tz7CSd9j",
		},
		{
			name:           "pass - successful encode with XRP alphabet - 3",
			input:          []byte("rUxb5vn9fGYRV3KZcnu3JLM4q5DTnNSavf"),
			expectedOutput: "s2uiNSCBQnQfsVtnX49adC9QqtWNP8upC16t7GFLrmbR7tm",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedOutput, EncodeBase58(tc.input))
		})
	}
}

func TestDecodeBase58(t *testing.T) {
	tt := []struct {
		name           string
		input          string
		expectedOutput []byte
	}{
		{
			name:           "pass - successful decode with XRP alphabet - 1",
			input:          "s2Fku4vaPpFiqqXdAD3V5rYrSx5a9h9qvUJW3423akZSCeD",
			expectedOutput: []byte("rDTXLQ7ZKZVKz33zJbHjgVShjsBnqMBhmN"),
		},
		{
			name:           "pass - successful decode with XRP alphabet - 2",
			input:          "s2i2Jk6bF44eDSXnnMjxeVhnYZ3qmbteqesuhS6Tz7CSd9j",
			expectedOutput: []byte("rJrpjzcxwQxokkqPxm62o5rtNfe2XimrTr"),
		},
		{
			name:           "pass - successful decode with XRP alphabet - 3",
			input:          "s2uiNSCBQnQfsVtnX49adC9QqtWNP8upC16t7GFLrmbR7tm",
			expectedOutput: []byte("rUxb5vn9fGYRV3KZcnu3JLM4q5DTnNSavf"),
		},
		{
			name:           "invalid non-ASCII character",
			input:          "Ā",
			expectedOutput: []byte{},
		},
		{
			name:           "invalid UTF-8",
			input:          "\xff",
			expectedOutput: []byte{},
		},
		{
			name:           "multibyte character across chunk boundary",
			input:          "rrrrrrrrr" + "é",
			expectedOutput: []byte{},
		},
		{
			name:           "long generic value",
			input:          strings.Repeat("r", 129),
			expectedOutput: make([]byte, 129),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectedOutput, DecodeBase58(tc.input))
		})
	}
}
