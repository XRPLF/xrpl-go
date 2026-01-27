// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// Data represents arbitrary metadata in hex format.
// The field is limited to 512 characters.
type Data string

// Value returns the string representation of Data.
func (d *Data) Value() string {
	return string(*d)
}
