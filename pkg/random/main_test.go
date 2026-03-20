package random

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomizer_GenerateBytes(t *testing.T) {
	testcases := []struct {
		name        string
		input       int
		expected    []byte
		expectedErr error
	}{
		{
			name:     "pass - 0 bytes",
			input:    0,
			expected: []byte{},
		},
		{
			name:     "pass - 10 bytes",
			input:    10,
			expected: make([]byte, 10),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			randomizer := NewRandomizer()
			bytes, err := randomizer.GenerateBytes(tc.input)
			fmt.Println(string(bytes))
			if tc.expectedErr != nil {
				require.Equal(t, tc.expectedErr, err)
			}
			if len(bytes) != tc.input {
				t.Errorf("expected %d bytes, got %d", tc.input, len(bytes))
			}
		})
	}
}
