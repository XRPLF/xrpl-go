// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// PreviousPaymentDueDate represents a date in ripple epoch.
type PreviousPaymentDueDate uint32

// Value returns the uint32 representation of the PreviousPaymentDueDate.
func (n *PreviousPaymentDueDate) Value() uint32 {
	return uint32(*n)
}
