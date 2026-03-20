//revive:disable:var-naming
package types

// AuthorizeCredentialList represents a list of AuthorizeCredential entries with validation and flattening.
type AuthorizeCredentialList []AuthorizeCredential

// Validate checks that the list is non-empty, within allowed size, has no duplicates, and each credential is valid.
func (ac *AuthorizeCredentialList) Validate() error {
	if len(*ac) == 0 {
		return ErrEmptyCredentials
	}
	if len(*ac) > MaxAcceptedCredentials {
		return ErrInvalidCredentialCount
	}
	seen := make(map[string]bool)
	for _, cred := range *ac {
		key := cred.Credential.Issuer.String() + cred.Credential.CredentialType.String()
		if seen[key] {
			return ErrDuplicateCredentials
		}
		seen[key] = true

		if err := cred.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Flatten returns a slice of maps representing each AuthorizeCredential in JSON-like format.
func (ac *AuthorizeCredentialList) Flatten() []map[string]any {
	acs := make([]map[string]any, len(*ac))
	for i, c := range *ac {
		acs[i] = c.Flatten()
	}
	return acs
}
