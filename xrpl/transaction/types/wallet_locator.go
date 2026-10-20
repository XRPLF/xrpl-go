//revive:disable:var-naming
package types

// WalletLocator returns a pointer to a Hash256 representing an arbitrary 256-bit value (optional).
// If specified, the value is stored as part of the account but has no inherent meaning or requirements.
func WalletLocator(value Hash256) *Hash256 {
	return new(value)
}
