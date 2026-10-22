//revive:disable:var-naming
package types

// TransferFee returns a pointer to a uint16 representing the TransferFee field.
func TransferFee(value uint16) *uint16 {
	return new(value)
}
