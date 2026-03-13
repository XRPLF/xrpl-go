// nolint
package types

// VaultWithdrawalPolicy represents the withdrawal strategy used by a Vault.
type VaultWithdrawalPolicy uint8

const (
	// VaultStrategyFirstComeFirstServe is the default withdrawal policy.
	// Withdrawals are processed on a first-come, first-served basis.
	VaultStrategyFirstComeFirstServe VaultWithdrawalPolicy = 0x0001
)

// Value returns the underlying uint8 value.
func (v VaultWithdrawalPolicy) Value() uint8 {
	return uint8(v)
}
