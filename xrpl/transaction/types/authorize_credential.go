//revive:disable:var-naming
package types

// MaxAcceptedCredentials is the maximum number of accepted credentials.
const MaxAcceptedCredentials int = 10

// AuthorizeCredential represents an accepted credential for PermissionedDomainSet transactions.
type AuthorizeCredential struct {
	Credential Credential
}

// Validate checks if the AuthorizeCredential is valid.
func (a AuthorizeCredential) Validate() error {
	if a.Credential.Issuer.String() == "" {
		return ErrInvalidCredentialIssuer
	}
	if !a.Credential.CredentialType.IsValid() {
		return ErrInvalidCredentialType
	}
	return nil
}

// Flatten returns a flattened map representation of the AuthorizeCredential.
func (a AuthorizeCredential) Flatten() map[string]any {
	m := make(map[string]any)
	m["Credential"] = a.Credential.Flatten()
	return m
}
