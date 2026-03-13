//revive:disable:var-naming
package types

import (
	"encoding/binary"
	"errors"
	"math/big"
	"regexp"
	"strings"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// Number represents the XRPL Number type (also known as STNumber in JS).
// It is encoded as 12 bytes: 8-byte signed mantissa + 4-byte signed exponent, both big-endian.
type Number struct{}

// Constants for mantissa and exponent normalization per XRPL Number spec.
var (
	minMantissa, _     = big.NewInt(0).SetString("1000000000000000000", 10) // 10^18
	maxMantissa, _     = big.NewInt(0).SetString("9999999999999999999", 10) // 10^19 - 1
	maxInt64, _        = big.NewInt(0).SetString("9223372036854775807", 10) // 2^63 - 1
	minExponent        = int32(-32768)
	maxExponent        = int32(32768)
	defaultZeroExp     = int32(-2147483648) // 0x80000000
	rangeLog           = 18
	ErrInvalidNumber   = errors.New("invalid Number string")
	ErrNumberOverflow  = errors.New("mantissa and exponent are too large")
	ErrInvalidExponent = errors.New("exponent out of range")
)

// numberRegex matches decimal/float/scientific number strings.
// Pattern: optional sign, integer part, optional decimal, optional exponent
var numberRegex = regexp.MustCompile(`^([-+]?)([0-9]+)(?:\.([0-9]+))?(?:[eE]([+-]?[0-9]+))?$`)

// FromJSON converts a JSON value (string) into a serialized 12-byte slice.
func (n *Number) FromJSON(value any) ([]byte, error) {
	s, ok := value.(string)
	if !ok {
		return nil, ErrInvalidNumber
	}

	mantissa, exponent, err := parseAndNormalize(s)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 12)
	writeInt64BE(buf, mantissa.Int64(), 0)
	writeInt32BE(buf, exponent, 8)

	return buf, nil
}

// ToJSON takes a BinaryParser and converts the serialized byte data back to a JSON string.
func (n *Number) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(12)
	if err != nil {
		return nil, err
	}

	mantissaRaw := readInt64BE(b, 0)
	exponent := readInt32BE(b, 8)

	mantissa := big.NewInt(mantissaRaw)

	// Special zero case
	if mantissa.Sign() == 0 && exponent == defaultZeroExp {
		return "0", nil
	}

	isNegative := mantissa.Sign() < 0
	mantissaAbs := new(big.Int).Abs(mantissa)
	ten := big.NewInt(10)

	// If mantissa < MIN_MANTISSA, it was shrunk for int64 serialization (mantissa > 2^63-1).
	// Restore it for proper string rendering to match rippled's internal representation.
	if mantissaAbs.Sign() != 0 && mantissaAbs.Cmp(minMantissa) < 0 {
		mantissaAbs.Mul(mantissaAbs, ten)
		exponent--
	}

	// Use scientific notation for exponents that are too small or too large

	//nolint:gosec // G115: integer overflow conversion int -> uint32 (gosec)
	if exponent != 0 && (exponent < -int32(rangeLog+10) || exponent > -int32(rangeLog-10)) {
		return formatScientificBig(mantissaAbs, exponent, isNegative), nil
	}

	// Decimal rendering
	return formatDecimalBig(mantissaAbs, exponent, isNegative), nil
}

// parseAndNormalize extracts mantissa, exponent from a string and normalizes them.
func parseAndNormalize(s string) (*big.Int, int32, error) {
	match := numberRegex.FindStringSubmatch(s)
	if match == nil {
		return nil, 0, ErrInvalidNumber
	}

	sign := match[1]
	intPart := match[2]
	fracPart := match[3]
	expPart := match[4]

	// Remove leading zeros (unless entire intPart is zeros)
	intPart = strings.TrimLeft(intPart, "0")
	if intPart == "" {
		intPart = "0"
	}

	mantissaStr := intPart
	exponent := int32(0)

	if fracPart != "" {
		mantissaStr += fracPart
		//nolint:gosec // G115: integer overflow conversion int -> int32 (gosec)
		exponent -= int32(len(fracPart))
	}

	if expPart != "" {
		var expVal int64
		_, err := parseIntFromString(expPart, &expVal)
		if err != nil {
			return nil, 0, err
		}
		//nolint:gosec // G115: integer overflow conversion int64 -> int32 (gosec)
		exponent += int32(expVal)
	}

	// Remove trailing zeros from mantissa and adjust exponent
	for len(mantissaStr) > 1 && mantissaStr[len(mantissaStr)-1] == '0' {
		mantissaStr = mantissaStr[:len(mantissaStr)-1]
		exponent++
	}

	mantissa := new(big.Int)
	mantissa.SetString(mantissaStr, 10)

	if sign == "-" {
		mantissa.Neg(mantissa)
	}

	// Check for zero
	if mantissa.Sign() == 0 {
		return big.NewInt(0), defaultZeroExp, nil
	}

	// Normalize
	mantissa, exponent, err := normalize(mantissa, exponent)
	if err != nil {
		return nil, 0, err
	}

	return mantissa, exponent, nil
}

// normalize adjusts mantissa and exponent to XRPL constraints.
func normalize(mantissa *big.Int, exponent int32) (*big.Int, int32, error) {
	isNegative := mantissa.Sign() < 0
	m := new(big.Int).Abs(mantissa)
	ten := big.NewInt(10)
	five := big.NewInt(5)

	// Handle zero
	if m.Sign() == 0 {
		return big.NewInt(0), defaultZeroExp, nil
	}

	// Scale up if too small
	for m.Cmp(minMantissa) < 0 && exponent > minExponent {
		exponent--
		m.Mul(m, ten)
	}

	// Scale down if too large, tracking last digit for rounding
	var lastDigit *big.Int
	for m.Cmp(maxMantissa) > 0 {
		if exponent >= maxExponent {
			return nil, 0, ErrNumberOverflow
		}
		exponent++
		lastDigit = new(big.Int).Mod(m, ten)
		m.Div(m, ten)
	}

	// Underflow check
	if exponent < minExponent || m.Cmp(minMantissa) < 0 {
		return nil, 0, errors.New("underflow: value too small to represent")
	}

	// Exponent overflow check
	if exponent > maxExponent {
		return nil, 0, ErrInvalidExponent
	}

	// MAX_INT64 check: mantissa must fit in signed int64
	if m.Cmp(maxInt64) > 0 {
		if exponent >= maxExponent {
			return nil, 0, ErrInvalidExponent
		}
		exponent++
		lastDigit = new(big.Int).Mod(m, ten)
		m.Div(m, ten)
	}

	// Rounding: if last discarded digit >= 5, round up
	if lastDigit != nil && lastDigit.Cmp(five) >= 0 {
		m.Add(m, big.NewInt(1))
		// Re-check after rounding may push mantissa over MAX_INT64
		if m.Cmp(maxInt64) > 0 {
			if exponent >= maxExponent {
				return nil, 0, ErrInvalidExponent
			}
			lastDigit = new(big.Int).Mod(m, ten)
			exponent++
			m.Div(m, ten)
			if lastDigit.Cmp(five) >= 0 {
				m.Add(m, big.NewInt(1))
			}
		}
	}

	if isNegative {
		m.Neg(m)
	}

	return m, exponent, nil
}

// formatScientificBig formats mantissa and exponent as scientific notation string.
// Strips trailing zeros from mantissa to match rippled behavior.
func formatScientificBig(mantissaAbs *big.Int, exponent int32, isNegative bool) string {
	m := new(big.Int).Set(mantissaAbs)
	ten := big.NewInt(10)
	zero := big.NewInt(0)
	exp := exponent

	// Strip trailing zeros from mantissa
	for m.Cmp(zero) != 0 && new(big.Int).Mod(m, ten).Sign() == 0 && exp < maxExponent {
		m.Div(m, ten)
		exp++
	}

	sign := ""
	if isNegative {
		sign = "-"
	}
	return sign + m.String() + "e" + itoa(int(exp))
}

// formatDecimalBig formats mantissa and exponent as a decimal string.
func formatDecimalBig(mantissaAbs *big.Int, exponent int32, isNegative bool) string {
	mantissaStr := mantissaAbs.String()

	padPrefix := rangeLog + 12 // 30
	padSuffix := rangeLog + 8  // 26
	rawValue := strings.Repeat("0", padPrefix) + mantissaStr + strings.Repeat("0", padSuffix)

	offset := int(exponent) + padPrefix + rangeLog + 1 // exponent + 49
	if offset < 0 {
		offset = 0
	}
	if offset > len(rawValue) {
		offset = len(rawValue)
	}

	integerPart := strings.TrimLeft(rawValue[:offset], "0")
	if integerPart == "" {
		integerPart = "0"
	}

	fractionPart := strings.TrimRight(rawValue[offset:], "0")

	result := integerPart
	if fractionPart != "" {
		result += "." + fractionPart
	}

	if isNegative {
		result = "-" + result
	}

	return result
}

// Helper functions for big-endian signed integer I/O

func writeInt64BE(buf []byte, v int64, offset int) {
	//nolint:gosec // G115: integer overflow conversion int64 -> uint64 (gosec)
	binary.BigEndian.PutUint64(buf[offset:], uint64(v))
}

func writeInt32BE(buf []byte, v int32, offset int) {
	//nolint:gosec // G115: integer overflow conversion int32 -> uint32 (gosec)
	binary.BigEndian.PutUint32(buf[offset:], uint32(v))
}

func readInt64BE(buf []byte, offset int) int64 {
	//nolint:gosec // G115: integer overflow conversion uint64 -> int64 (gosec)
	return int64(binary.BigEndian.Uint64(buf[offset:]))
}

func readInt32BE(buf []byte, offset int) int32 {
	//nolint:gosec // G115: integer overflow conversion uint32 -> int32 (gosec)
	return int32(binary.BigEndian.Uint32(buf[offset:]))
}

func parseIntFromString(s string, result *int64) (bool, error) {
	n := new(big.Int)
	_, ok := n.SetString(s, 10)
	if !ok {
		return false, ErrInvalidNumber
	}
	*result = n.Int64()
	return true, nil
}

func itoa(n int) string {
	if n < 0 {
		return "-" + uitoa(uint(-n))
	}
	return uitoa(uint(n))
}

func uitoa(n uint) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
