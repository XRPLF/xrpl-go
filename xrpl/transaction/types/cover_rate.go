// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// CoverRate represents a cover rate in 1/10th basis points.
// Valid values are between 0 and 100000 inclusive. A value of 1 is equivalent to 1/10 bps or 0.001%.
type CoverRate uint32

// Value returns the uint32 representation of CoverRate.
func (cr *CoverRate) Value() uint32 {
	return uint32(*cr)
}
