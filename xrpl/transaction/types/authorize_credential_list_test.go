package types

import (
	"testing"
)

func TestAuthorizeCredentialList_Validate(t *testing.T) {
	tests := []struct {
		name        string
		ac          AuthorizeCredentialList
		wantValid   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name: "pass - valid credential list",
			ac: AuthorizeCredentialList{
				AuthorizeCredential{
					Credential: Credential{
						Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						CredentialType: CredentialType("1234"),
					},
				},
			},
			wantValid:   true,
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "fail - empty credential list",
			ac:          AuthorizeCredentialList{},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrEmptyCredentials,
		},
		{
			name: "fail - invalid credential",
			ac: AuthorizeCredentialList{
				AuthorizeCredential{
					Credential: Credential{
						Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						CredentialType: CredentialType("1234"),
					},
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidCredentialType,
		},
		{
			name: "fail - duplicated credential",
			ac: AuthorizeCredentialList{
				AuthorizeCredential{
					Credential: Credential{
						Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						CredentialType: CredentialType("1234"),
					},
				},
				AuthorizeCredential{
					Credential: Credential{
						Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						CredentialType: CredentialType("1234"),
					},
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrDuplicateCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ac.Validate()
			if err != nil {
				if err != tt.expectedErr {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestAuthorizeCredentialList_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		ac       AuthorizeCredentialList
		expected string
	}{
		{
			name: "pass - valid credential list",
			ac: AuthorizeCredentialList{
				AuthorizeCredential{
					Credential: Credential{
						Issuer:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						CredentialType: CredentialType("1234"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattened := tt.ac.Flatten()
			if flattened == nil {
				t.Errorf("expected flattened credential list, got nil")
			}
			if len(flattened) != 1 {
				t.Errorf("expected 1 credential, got %d", len(flattened))
			}
		})
	}
}
