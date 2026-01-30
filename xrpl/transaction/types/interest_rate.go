// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// InterestRate represents the interest rate in 1/10th basis points.
// Valid values are between 0 and 100000 inclusive. (0 - 100%)
type InterestRate uint32

// Value returns the uint32 representation of InterestRate.
func (i *InterestRate) Value() uint32 {
	return uint32(*i)
}
