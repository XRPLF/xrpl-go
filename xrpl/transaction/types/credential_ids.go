//revive:disable:var-naming
package types

import "encoding/hex"

// CredentialIDs represents a list of credential ID strings encoded as hexadecimal values.
type CredentialIDs []string

// IsValid checks for one to eight distinct, nonzero 256-bit hexadecimal IDs.
// IDs are compared by decoded value, so hex letter case does not affect uniqueness.
// This offline check applies the fixCleanup3_4_0 zero-ID rule without checking
// amendment activation. It does not check credential existence or authorization.
// Callers handle an omitted (nil) list separately.
func (c CredentialIDs) IsValid() bool {
	if len(c) == 0 || len(c) > 8 {
		return false
	}

	seen := make(map[[32]byte]struct{}, len(c))
	for _, id := range c {
		if len(id) != 64 {
			return false
		}
		var hash [32]byte
		if _, err := hex.Decode(hash[:], []byte(id)); err != nil || hash == [32]byte{} {
			return false
		}
		if _, duplicate := seen[hash]; duplicate {
			return false
		}
		seen[hash] = struct{}{}
	}
	return true
}

// Flatten returns the underlying slice of credential ID strings.
func (c CredentialIDs) Flatten() []string {
	return c
}
