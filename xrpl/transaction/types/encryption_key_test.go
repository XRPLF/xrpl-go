package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptionKey(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"public key", "0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncryptionKey(tt.value)
			require.NotNil(t, got)
			require.Equal(t, tt.value, *got)
			require.NotSame(t, got, EncryptionKey(tt.value))
		})
	}
}
