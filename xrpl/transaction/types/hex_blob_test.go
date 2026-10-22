package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHexBlob(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"hex data", "DEADBEEF"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HexBlob(tt.value)
			require.NotNil(t, got)
			require.Equal(t, tt.value, *got)
			require.NotSame(t, got, HexBlob(tt.value))
		})
	}
}
