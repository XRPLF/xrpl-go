//revive:disable:var-naming
package types

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
)

// AuthorizeCredentialsWrapper wraps AuthorizeCredentials for optional inclusion in a transaction.
type AuthorizeCredentialsWrapper struct {
	Credential AuthorizeCredentials
}

// AuthorizeCredentials represents the credentials to be authorized, including issuer and credential type.
type AuthorizeCredentials struct {
	// The issuer of the credential.
	Issuer Address
	// The credential type of the credential.
	CredentialType CredentialType
}

// IsValid returns true if the authorize credentials are valid.
func (a *AuthorizeCredentials) IsValid() bool {
	return addresscodec.IsValidAddress(a.Issuer.String()) && a.CredentialType.IsValid()
}

// Flatten returns a map of the authorize credentials.
func (a *AuthorizeCredentialsWrapper) Flatten() map[string]any {
	flattened := make(map[string]any)

	flattened["Credential"] = a.Credential.Flatten()

	return flattened
}

// Flatten returns a map of the authorize credentials.
func (a *AuthorizeCredentials) Flatten() map[string]any {
	flattened := make(map[string]any)

	if a.Issuer != "" {
		flattened["Issuer"] = a.Issuer.String()
	}
	if a.CredentialType != "" {
		flattened["CredentialType"] = a.CredentialType.String()
	}

	return flattened
}
