package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthorizeCredentials_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		credential AuthorizeCredentials
		expected   bool
	}{
		{
			name: "pass - valid authorize credentials",
			credential: AuthorizeCredentials{
				Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: "6D795F63726564656E7469616C",
			},
			expected: true,
		},
		{
			name: "fail - invalid credential type",
			credential: AuthorizeCredentials{
				Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				CredentialType: "invalid",
			},
			expected: false,
		},
		{
			name: "fail - invalid issuer",
			credential: AuthorizeCredentials{
				Issuer:         "invalid",
				CredentialType: "6D795F63726564656E7469616C",
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.credential.IsValid()
			require.Equal(t, test.expected, result)
		})
	}
}

func TestAuthorizeCredentials_Flatten(t *testing.T) {
	tests := []struct {
		name       string
		credential AuthorizeCredentialsWrapper
		expected   map[string]interface{}
	}{
		{
			name: "pass - valid authorize credentials",
			credential: AuthorizeCredentialsWrapper{
				Credential: AuthorizeCredentials{
					Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					CredentialType: "6D795F63726564656E7469616C",
				},
			},
			expected: map[string]interface{}{
				"Credential": map[string]interface{}{
					"Issuer":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					"CredentialType": "6D795F63726564656E7469616C",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.credential.Flatten()
			require.Equal(t, test.expected, result)
		})
	}
}
