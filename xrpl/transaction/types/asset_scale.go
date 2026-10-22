//revive:disable:var-naming
package types

// AssetScale returns a pointer to the provided uint8 value, used for asset scale parameters.
func AssetScale(value uint8) *uint8 {
	return new(value)
}
