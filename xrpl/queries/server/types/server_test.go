package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStateUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string
	}{
		{
			name:     "validator_list_expires as string",
			json:     `{"validator_list_expires": "2025-Jan-01 00:00:00 UTC"}`,
			expected: "2025-Jan-01 00:00:00 UTC",
		},
		{
			name:     "validator_list_expires as number zero",
			json:     `{"validator_list_expires": 0}`,
			expected: "0",
		},
		{
			name:     "validator_list_expires as nonzero number",
			json:     `{"validator_list_expires": 740944800}`,
			expected: "740944800",
		},
		{
			name:     "validator_list_expires absent",
			json:     `{}`,
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var s State
			err := json.Unmarshal([]byte(tc.json), &s)
			require.NoError(t, err)
			require.Equal(t, tc.expected, s.ValidatorListExpires)
		})
	}
}
