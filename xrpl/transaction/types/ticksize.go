// Package types defines helper types for transaction optional fields.
// revive:disable:var-naming
package types

// TickSize returns an optional tick size to use for offers involving a currency issued by this address.
// The exchange rates of those offers are rounded to this many significant digits. Valid values are 3 to 15 inclusive, or 0 to disable.
func TickSize(value uint8) *uint8 {
	return new(value)
}
