//revive:disable:var-naming
package types

// Expiration returns a pointer to a uint32 representing the Expiration field (optional).
func Expiration(value uint32) *uint32 {
	return new(value)
}
