package transaction

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathStep_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		step     PathStep
		expected map[string]any
	}{
		{
			name:     "pass - empty",
			step:     PathStep{},
			expected: map[string]any{},
		},
		{
			name:     "pass - account",
			step:     PathStep{Account: "r3dFAtNXwRFCyBGz5BcWhMj9a4cm7qkzzn"},
			expected: map[string]any{"account": "r3dFAtNXwRFCyBGz5BcWhMj9a4cm7qkzzn"},
		},
		{
			name:     "pass - currency",
			step:     PathStep{Currency: "XRP"},
			expected: map[string]any{"currency": "XRP"},
		},
		{
			name: "pass - currency and issuer",
			step: PathStep{Currency: "USD", Issuer: "r3dFAtNXwRFCyBGz5BcWhMj9a4cm7qkzzn"},
			expected: map[string]any{
				"currency": "USD",
				"issuer":   "r3dFAtNXwRFCyBGz5BcWhMj9a4cm7qkzzn",
			},
		},
		{
			name: "pass - X-addresses flatten to classic addresses",
			step: PathStep{
				Account:  "X7tR8jTwG7vJXotpo2oV43qtSxQXTSqrh7MjigzRZyqVvVr",
				Currency: "USD",
				Issuer:   "X7tR8jTwG7vJXotpo2oV43qtSxQXTSqrh7MjigzRZyqVvVr",
			},
			expected: map[string]any{
				"account":  "r3dFAtNXwRFCyBGz5BcWhMj9a4cm7qkzzn",
				"currency": "USD",
				"issuer":   "r3dFAtNXwRFCyBGz5BcWhMj9a4cm7qkzzn",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.step.Flatten())
		})
	}
}
