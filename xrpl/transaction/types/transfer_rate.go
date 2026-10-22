//revive:disable:var-naming
package types

// TransferRate returns a pointer to a uint32 representing the TransferRate field (optional).
// The fee to charge when users transfer this account's tokens, represented as billionths of a unit.
// Cannot be more than 2000000000 or less than 1000000000, except for the special case 0 meaning no fee.
func TransferRate(value uint32) *uint32 {
	return new(value)
}
