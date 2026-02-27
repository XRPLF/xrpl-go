// Package hexutil provides utility functions for hexadecimal encoding.
package hexutil

import (
	"encoding/hex"
	"strings"
)

// EncodeToUpperHex encodes bytes to an uppercase hexadecimal string.
func EncodeToUpperHex(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}
