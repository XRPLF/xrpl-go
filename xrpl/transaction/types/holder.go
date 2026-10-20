// Package types provides core transaction types and helpers for the XRPL Go library.
//
// revive:disable:var-naming
package types

// Holder returns a pointer to an Address representing the Holder field (optional).
func Holder(address Address) *Address {
	return new(address)
}
