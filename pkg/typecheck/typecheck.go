// Package typecheck provides functions for runtime type assertions and numeric string validation.
package typecheck

import (
	"regexp"
	"strconv"
)

// IsUint8 checks if the given interface is a uint8.
func IsUint8(num interface{}) bool {
	_, ok := num.(uint8)
	return ok
}

// IsString checks if the given interface is a string.
func IsString(str interface{}) bool {
	_, ok := str.(string)
	return ok
}

// IsUint32 checks if the given interface is a uint32.
func IsUint32(num interface{}) bool {
	_, ok := num.(uint32)
	return ok
}

// IsUint64 checks if the given interface is a uint64.
func IsUint64(num interface{}) bool {
	_, ok := num.(uint64)
	return ok
}

// IsUint checks if the given interface is a uint.
func IsUint(num interface{}) bool {
	_, ok := num.(uint)
	return ok
}

// IsInt checks if the given interface is an int.
func IsInt(num interface{}) bool {
	_, ok := num.(int)
	return ok
}

// IsBool checks if the given interface is a bool.
func IsBool(b interface{}) bool {
	_, ok := b.(bool)
	return ok
}

// IsHex checks if the given string is a valid hexadecimal string.
func IsHex(s string) bool {
	// Define a regular expression for a valid hexadecimal string
	var validHexPattern = regexp.MustCompile(`^[0-9a-fA-F]+$`)
	return validHexPattern.MatchString(s)
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
