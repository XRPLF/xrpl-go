package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentInterval_Value(t *testing.T) {
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
			name:  "minimum valid value",
			value: 60,
			want:  60,
		},
		{
			name:  "typical value",
			value: 2592000,
			want:  2592000,
		},
		{
			name:  "max uint32",
			value: 4294967295,
			want:  4294967295,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PaymentInterval(tt.value)
			require.Equal(t, tt.want, (&result).Value())
		})
	}
}
