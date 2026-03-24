//revive:disable:var-naming
package types

// Credential represents an XRPL credential, including the issuer address and credential type.
type Credential struct {
	// The issuer of the credential.
	Issuer Address
	// A hex-encoded value to identify the type of credential from the issuer.
	CredentialType CredentialType
}

// Flatten returns a map of the Credential fields for transaction encoding.
func (c Credential) Flatten() map[string]any {
	m := make(map[string]any)
	m["Issuer"] = c.Issuer.String()
	m["CredentialType"] = c.CredentialType.String()
	return m
}
