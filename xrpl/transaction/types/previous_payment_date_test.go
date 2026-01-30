package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPreviousPaymentDate_Value(t *testing.T) {
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
			name:  "typical timestamp",
			value: 1724871860,
			want:  1724871860,
		},
		{
			name:  "max uint32",
			value: 4294967295,
			want:  4294967295,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PreviousPaymentDate(tt.value)
			require.Equal(t, tt.want, (&result).Value())
		})
	}
}
