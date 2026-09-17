package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCredentialIDs_IsValid(t *testing.T) {
	const credentialIDBytes = 32
	const maxCredentialIDs = 8
	id := strings.Repeat("AB", credentialIDBytes)
	maximum := make(CredentialIDs, maxCredentialIDs)
	for i := range maximum {
		maximum[i] = fmt.Sprintf("%0*X", 2*credentialIDBytes, i+1)
	}
	tests := []struct {
		name          string
		credentialIDs CredentialIDs
		expected      bool
	}{
		{"fail - nil", nil, false},
		{"fail - empty", CredentialIDs{}, false},
		{"pass - one", CredentialIDs{id}, true},
		{"pass - lowercase", CredentialIDs{strings.ToLower(id)}, true},
		{"pass - eight", maximum, true},
		{"fail - nine", append(append(CredentialIDs{}, maximum...), id), false},
		{"fail - duplicate", CredentialIDs{id, id}, false},
		{"fail - duplicate case", CredentialIDs{id, strings.ToLower(id)}, false},
		{"fail - empty id", CredentialIDs{""}, false},
		{"fail - zero id", CredentialIDs{strings.Repeat("00", credentialIDBytes)}, false},
		{"fail - short", CredentialIDs{"AB"}, false},
		{"fail - odd", CredentialIDs{id[:len(id)-1]}, false},
		{"fail - long", CredentialIDs{id + "00"}, false},
		{"fail - nonhex", CredentialIDs{strings.Repeat("ZZ", credentialIDBytes)}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.credentialIDs.IsValid()
			require.Equal(t, test.expected, result)
		})
	}
}

func TestCredentialIDs_Flatten(t *testing.T) {
	tests := []struct {
		name          string
		credentialIDs CredentialIDs
		expected      []string
	}{
		{
			name:          "pass - empty",
			credentialIDs: CredentialIDs{},
			expected:      []string{},
		},
		{
			name:          "pass - valid",
			credentialIDs: CredentialIDs{"0000000000000000000000000000000000000000000000000000000000000000"},
			expected:      []string{"0000000000000000000000000000000000000000000000000000000000000000"},
		},
		{
			name:          "pass - valid with two ids",
			credentialIDs: CredentialIDs{"0000000000000000000000000000000000000000000000000000000000000000", "6D795F63726564656E7469616C"},
			expected:      []string{"0000000000000000000000000000000000000000000000000000000000000000", "6D795F63726564656E7469616C"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.credentialIDs.Flatten()
			require.Equal(t, test.expected, result)
		})
	}
}
