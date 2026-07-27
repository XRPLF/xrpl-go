package types

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPermissionValue_FromJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    []byte
		expectedErr error
	}{
		{
			name:     "pass - uint32 value",
			input:    uint32(0x01020304),
			expected: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name:     "pass - json.Number value",
			input:    json.Number("1"),
			expected: []byte{0x00, 0x00, 0x00, 0x01},
		},
		{
			name:     "pass - permission name string",
			input:    "TrustlineAuthorize",
			expected: []byte{0x00, 0x01, 0x00, 0x01},
		},
		{
			name:        "fail - invalid json number",
			input:       json.Number("abc"),
			expectedErr: ErrPermissionValueOutOfRange,
		},
		{
			name:        "fail - negative json number",
			input:       json.Number("-1"),
			expectedErr: ErrPermissionValueOutOfRange,
		},
		{
			name:        "fail - json number out of range",
			input:       json.Number("4294967296"),
			expectedErr: ErrPermissionValueOutOfRange,
		},
		{
			name:        "fail - numeric value out of range",
			input:       uint64(math.MaxUint32) + 1,
			expectedErr: ErrPermissionValueOutOfRange,
		},
		{
			name:        "fail - unsupported type",
			input:       true,
			expectedErr: ErrUnsupportedPermissionType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := (&PermissionValue{}).FromJSON(tt.input)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				require.Nil(t, actual)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}
