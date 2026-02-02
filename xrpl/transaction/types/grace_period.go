// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// GracePeriod represents the number of seconds after the Payment Due Date that the Loan can be Defaulted.
type GracePeriod uint32

// Value returns the uint32 representation of GracePeriod.
func (g *GracePeriod) Value() uint32 {
	return uint32(*g)
}
