package keypairs

import (
	"testing"

	"github.com/Peersyst/xrpl-go/keypairs/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/stretchr/testify/require"
)

func TestGetCryptoImplementationFromKey(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected interfaces.KeypairCryptoAlg
	}{
		{
			name:     "fail - empty key",
			input:    "",
			expected: nil,
		},
		{
			name:     "fail - short key",
			input:    "E",
			expected: nil,
		},
		{
			name:     "fail - invalid key",
			input:    "invalid",
			expected: nil,
		},
		{
			name:     "pass - get ED25519 implementation",
			input:    "ED4924A9045FE5ED8B22BAA7B6229A72A287CCF3EA287AADD3A032A24C0F008FA6",
			expected: crypto.ED25519(),
		},
		{
			name:     "pass - get SECP256K1 implementation",
			input:    "0003540DE0F1438F58C4822F99795AD3D1F83C8D123C7767228E04185C542C41680D",
			expected: crypto.SECP256K1(),
		},
		{
			name:     "pass - get nil implementation",
			input:    "0103540DE0F1438F58C4822F99795AD3D1F83C8D123C7767228E04185C542C41680D",
			expected: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, getCryptoImplementationFromKey(tc.input))
		})
	}
}
