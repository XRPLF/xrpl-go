package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHolder(t *testing.T) {
	tests := []struct {
		name  string
		value Address
	}{
		{"classic address", "rNZ9m6AP9K7z3EVg6GhPMx36V4QmZKeWds"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Holder(tt.value)
			require.NotNil(t, got)
			require.Equal(t, tt.value, *got)
			require.NotSame(t, got, Holder(tt.value))
		})
	}
}
