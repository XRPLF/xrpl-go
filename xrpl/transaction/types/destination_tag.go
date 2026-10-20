//revive:disable:var-naming
package types

// DestinationTag returns a pointer to the provided uint32 value for use as a destination tag in transactions.
func DestinationTag(value uint32) *uint32 {
	return new(value)
}
