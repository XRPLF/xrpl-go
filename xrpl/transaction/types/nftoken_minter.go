//revive:disable:var-naming
package types

// NFTokenMinter returns a pointer to a string specifying an alternate account allowed to mint NFTokens on this account's behalf (optional).
// It sets the `Issuer` field for NFTokenMint transactions.
func NFTokenMinter(value string) *string {
	return new(value)
}
