// Package hexutil provides utility functions for hexadecimal encoding.
package hexutil

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// EncodeToUpperHex encodes bytes to an uppercase hexadecimal string.
func EncodeToUpperHex(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}

// DecodeFixedHex decodes a hex string and validates it decodes to exactly size bytes.
func DecodeFixedHex(hexStr string, size int) ([]byte, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	if len(b) != size {
		return nil, fmt.Errorf("expected %d bytes, got %d", size, len(b))
	}
	return b, nil
}
