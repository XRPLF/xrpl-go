package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOwnerCount_Value(t *testing.T) {
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
			name:  "single value",
			value: 1,
			want:  1,
		},
		{
			name:  "typical value",
			value: 100,
			want:  100,
		},
		{
			name:  "max uint32",
			value: 4294967295,
			want:  4294967295,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OwnerCount(tt.value)
			require.Equal(t, tt.want, (&result).Value())
		})
	}
}
