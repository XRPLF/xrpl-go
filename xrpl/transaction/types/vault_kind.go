package types

// VaultKind identifies when a vault accepts deposits and redemptions.
type VaultKind uint8

const (
	// VaultKindOpen permits redemption at any time and is the default kind.
	VaultKindOpen VaultKind = 0
	// VaultKindClosed restricts deposits and redemptions to their respective periods.
	VaultKindClosed VaultKind = 1
)
