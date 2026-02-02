// Package types provides core transaction types and helpers for the XRPL Go library.
//
//revive:disable:var-naming
package types

// OwnerCount represents the number of active objects owned by an account or entity.
type OwnerCount uint32

// Value returns the uint32 representation of OwnerCount.
func (o *OwnerCount) Value() uint32 {
	return uint32(*o)
}
