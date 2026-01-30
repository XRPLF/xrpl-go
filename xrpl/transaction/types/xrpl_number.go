// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// XRPLNumber represents an XRPL number as a string.
// XRPL numbers are strings that represent numbers, including scientific notation.
type XRPLNumber string

// String returns the string representation of the XRPLNumber.
func (n *XRPLNumber) String() string {
	return string(*n)
}
