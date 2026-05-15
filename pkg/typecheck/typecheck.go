// Package typecheck provides functions for runtime type assertions and numeric string validation.
package typecheck

import (
	"encoding/json"
	"math"
	"math/big"
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

// ToUint32 converts any integer or whole-number floating point value
// (including json.Number) to a uint32 when the exact value fits the
// [0, math.MaxUint32] range. The second return value reports success.
func ToUint32(v any) (uint32, bool) {
	switch n := v.(type) {
	case uint32:
		return n, true
	case uint:
		if uint64(n) > uint64(math.MaxUint32) {
			return 0, false
		}
		return uint32(n), true
	case uint8:
		return uint32(n), true
	case uint16:
		return uint32(n), true
	case uint64:
		if n > uint64(math.MaxUint32) {
			return 0, false
		}
		return uint32(n), true
	case int:
		return int64ToUint32(int64(n))
	case int8:
		return int64ToUint32(int64(n))
	case int16:
		return int64ToUint32(int64(n))
	case int32:
		return int64ToUint32(int64(n))
	case int64:
		return int64ToUint32(n)
	case float32:
		return floatToUint32(float64(n))
	case float64:
		return floatToUint32(n)
	case json.Number:
		return jsonNumberToUint32(n)
	default:
		return 0, false
	}
}

func jsonNumberToUint32(n json.Number) (uint32, bool) {
	s := n.String()
	if strings.Contains(s, "/") {
		return 0, false
	}
	v, ok := new(big.Rat).SetString(s)
	if !ok || !v.IsInt() {
		return 0, false
	}
	i := v.Num()
	if i.Sign() < 0 || i.BitLen() > 32 {
		return 0, false
	}
	//nolint:gosec // G115: i is range-checked above
	return uint32(i.Uint64()), true
}

func int64ToUint32(v int64) (uint32, bool) {
	if v < 0 || v > int64(math.MaxUint32) {
		return 0, false
	}
	//nolint:gosec // G115: v is range-checked above
	return uint32(v), true
}

func floatToUint32(v float64) (uint32, bool) {
	if math.IsNaN(v) || math.IsInf(v, 0) || v != math.Trunc(v) {
		return 0, false
	}
	if v < 0 || v > float64(math.MaxUint32) {
		return 0, false
	}
	//nolint:gosec // G115: v is range-checked above
	return uint32(int64(v)), true
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
