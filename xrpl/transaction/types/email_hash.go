//revive:disable:var-naming
package types

// EmailHash returns an arbitrary 128-bit value. Conventionally, clients treat this as the md5 hash of an email address to use for displaying a Gravatar image.
func EmailHash(value Hash128) *Hash128 {
	return new(value)
}
