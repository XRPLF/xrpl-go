package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssetScale(t *testing.T) {
	tests := []struct {
		name  string
		value uint8
	}{
		{"zero", 0},
		{"typical scale", 6},
		{"maximum", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AssetScale(tt.value)
			require.NotNil(t, got)
			require.Equal(t, tt.value, *got)
			require.NotSame(t, got, AssetScale(tt.value))
		})
	}
}
