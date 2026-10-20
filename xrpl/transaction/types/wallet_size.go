//revive:disable:var-naming
package types

// WalletSize returns a pointer to a uint32 representing the wallet size (optional).
// Not used. This field is valid in AccountSet transactions but does nothing.
func WalletSize(value uint32) *uint32 {
	return new(value)
}
