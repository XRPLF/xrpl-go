// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// PaymentInterval represents the number of seconds between Loan payments.
// The minimum valid value is 60 seconds.
type PaymentInterval uint32

// Value returns the uint32 representation of PaymentInterval.
func (p *PaymentInterval) Value() uint32 {
	return uint32(*p)
}
