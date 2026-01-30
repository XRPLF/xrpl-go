// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// PaymentTotal represents the total number of payments to be made against the Loan.
type PaymentTotal uint32

// Value returns the uint32 representation of PaymentTotal.
func (p *PaymentTotal) Value() uint32 {
	return uint32(*p)
}
