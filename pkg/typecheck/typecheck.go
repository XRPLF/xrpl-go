// Package typecheck provides functions for runtime type assertions and numeric string validation.
package typecheck

import (
	"regexp"
	"strconv"
	"strings"
)

// IsUint8 checks if the given interface is a uint8.
func IsUint8(num any) bool {
	_, ok := num.(uint8)
	return ok
}

// IsString checks if the given interface is a string.
func IsString(str any) bool {
	_, ok := str.(string)
	return ok
}

// IsUint32 checks if the given interface is a uint32.
func IsUint32(num any) bool {
	_, ok := num.(uint32)
	return ok
}

// IsUint64 checks if the given interface is a uint64.
func IsUint64(num any) bool {
	_, ok := num.(uint64)
	return ok
}

// IsUint checks if the given interface is a uint.
func IsUint(num any) bool {
	_, ok := num.(uint)
	return ok
}

// IsInt checks if the given interface is an int.
func IsInt(num any) bool {
	_, ok := num.(int)
	return ok
}

// IsBool checks if the given interface is a bool.
func IsBool(b any) bool {
	_, ok := b.(bool)
	return ok
}

// IsHex checks if the given string is a valid hexadecimal string.
// Empty strings return false.
func IsHex(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'A' && r <= 'F' {
			continue
		}
		if r >= 'a' && r <= 'f' {
			continue
		}
		return false
	}
	return true
}

// IsHexBlob reports whether s is a hex string that encodes whole bytes
// (valid hex characters and even length). Empty strings return false.
func IsHexBlob(s string) bool {
	return IsHex(s) && len(s)%2 == 0
}

// IsFloat32 checks if the given string is a valid Float32 number.
func IsFloat32(s string) bool {
	_, err := strconv.ParseFloat(s, 32)
	return err == nil
}

// IsFloat64 checks if the given string is a valid Float64 number.
func IsFloat64(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsStringNumericUint checks if the given string is a valid unsigned integer with the specified base and bit size.
func IsStringNumericUint(s string, base, bitSize int) bool {
	_, err := strconv.ParseUint(s, base, bitSize)
	return err == nil
}

// IsMap checks if the given interface is a map.
func IsMap(m any) bool {
	_, ok := m.(map[string]any)
	return ok
}

// xrplNumberPattern matches optional sign, digits, optional decimal, optional exponent (scientific).
// Allows leading zeros; rejects empty string, lone sign, or missing digits.
var xrplNumberPattern = regexp.MustCompile(`^[-+]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][-+]?\d+)?$`)

// IsXRPLNumber checks if the value is a valid XRPL number string.
// XRPL numbers are strings that represent numbers, including scientific notation.
func IsXRPLNumber(value any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	return xrplNumberPattern.MatchString(strings.TrimSpace(str))
}
