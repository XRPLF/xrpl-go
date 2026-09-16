//revive:disable:var-naming
package types

import (
	"encoding/hex"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
)

// CredentialIDs represents a list of credential ID strings encoded as hexadecimal values.
type CredentialIDs []string

// IsValid checks that the CredentialIDs slice is non-empty and contains only valid hex strings.
func (c CredentialIDs) IsValid() bool {
	if len(c) == 0 {
		return false
	}

	for _, id := range c {
		if !typecheck.IsHex(id) {
			return false
		}
	}

	return true
}

// IsValidForWithdrawal checks for one to eight distinct, nonzero 256-bit IDs.
// VaultWithdraw and LoanBrokerCoverWithdraw require these checks under
// Credentials and fixCleanup3_4_0. IsValid retains its legacy behavior.
// Callers handle an omitted (nil) list separately.
func (c CredentialIDs) IsValidForWithdrawal() bool {
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
