package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImmutableFlags(t *testing.T) {
	tests := []struct {
		name  string
		value uint32
	}{
		{"zero", 0},
		{"one flag", 0x00000001},
		{"all bits", 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ImmutableFlags(tt.value)
			require.NotNil(t, got)
			require.Equal(t, tt.value, *got)
			require.NotSame(t, got, ImmutableFlags(tt.value))
		})
	}
}
