package types

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestAuthorizeCredential_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		ac       AuthorizeCredential
		expected string
	}{
		{
			name: "pass - valid credential",
			ac: AuthorizeCredential{
				Credential: Credential{
					Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					CredentialType: CredentialType("1234"),
				},
			},
			expected: `{
				"Credential": {
					"Issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					"CredentialType": "1234"
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tt.ac.Flatten(), []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAuthorizeCredential_Validate(t *testing.T) {
	tests := []struct {
		name          string
		ac            AuthorizeCredential
		expected      bool
		expectedError error
	}{
		{
			name: "pass - valid credential",
			ac: AuthorizeCredential{
				Credential: Credential{
					Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					CredentialType: CredentialType("1234"),
				},
			},
			expected: true,
		},
		{
			name: "fail - invalid credential",
			ac: AuthorizeCredential{
				Credential: Credential{
					Issuer:         "",
					CredentialType: CredentialType("1234"),
				},
			},
			expected:      false,
			expectedError: ErrInvalidCredentialIssuer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ac.Validate()
			if err != nil {
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			}
		})
	}
}
