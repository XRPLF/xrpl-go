package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestData_Value(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "empty string",
			value: "",
			want:  "",
		},
		{
			name:  "typical hex string",
			value: "A1B2C3D4E5F6",
			want:  "A1B2C3D4E5F6",
		},
		{
			name:  "max length string",
			value: "A" + string(make([]byte, 511)),
			want:  "A" + string(make([]byte, 511)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Data(tt.value)
			require.Equal(t, tt.want, (&result).Value())
		})
	}
}
