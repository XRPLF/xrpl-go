package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoverRate_Value(t *testing.T) {
	tests := []struct {
		name  string
		value uint32
		want  uint32
	}{
		{
			name:  "zero value",
			value: 0,
			want:  0,
		},
		{
			name:  "minimum value",
			value: 1,
			want:  1,
		},
		{
			name:  "typical value",
			value: 50000,
			want:  50000,
		},
		{
			name:  "maximum valid value",
			value: 100000,
			want:  100000,
		},
		{
			name:  "max uint32",
			value: 4294967295,
			want:  4294967295,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CoverRate(tt.value)
			require.Equal(t, tt.want, (&result).Value())
		})
	}
}
