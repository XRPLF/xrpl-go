package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDomainID(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"domain ID", "6FA75B2EB30C4D0E2B1C1A3D7B5F4C9E8A0D2E1F3B4C5D6E7F8091A2B3C4D5E6"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DomainID(tt.value)
			require.NotNil(t, got)
			require.Equal(t, tt.value, *got)
			require.NotSame(t, got, DomainID(tt.value))
		})
	}
}
