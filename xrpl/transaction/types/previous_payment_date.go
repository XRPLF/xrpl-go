// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// PreviousPaymentDate represents a date in ripple epoch.
type PreviousPaymentDate uint32

// Value returns the uint32 representation of the PreviousPaymentDate.
func (n *PreviousPaymentDate) Value() uint32 {
	return uint32(*n)
}
