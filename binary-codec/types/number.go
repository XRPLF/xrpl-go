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
	minMantissa        = big.NewInt(1000000000000000)  // 10^15
	maxMantissa        = big.NewInt(9999999999999999) // 10^16 - 1
	minExponent        = int32(-32768)
	maxExponent        = int32(32768)
	defaultZeroExp     = int32(-2147483648) // 0x80000000
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

	mantissa := readInt64BE(b, 0)
	exponent := readInt32BE(b, 8)

	// Special zero case
	if mantissa == 0 && exponent == defaultZeroExp {
		return "0", nil
	}

	if exponent == 0 {
		return big.NewInt(mantissa).String(), nil
	}

	// Use scientific notation for very small/large exponents
	if exponent < -25 || exponent > -5 {
		return formatScientific(mantissa, exponent), nil
	}

	// Decimal rendering for -25 <= exp <= -5
	return formatDecimal(mantissa, exponent), nil
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
		exponent -= int32(len(fracPart))
	}

	if expPart != "" {
		var expVal int64
		_, err := parseIntFromString(expPart, &expVal)
		if err != nil {
			return nil, 0, err
		}
		exponent += int32(expVal)
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

	// Scale up if too small
	for m.Sign() != 0 && m.Cmp(minMantissa) < 0 && exponent > minExponent {
		exponent--
		m.Mul(m, ten)
	}

	// Scale down if too large
	for m.Cmp(maxMantissa) > 0 {
		if exponent >= maxExponent {
			return nil, 0, ErrNumberOverflow
		}
		exponent++
		m.Div(m, ten)
	}

	if isNegative {
		m.Neg(m)
	}

	return m, exponent, nil
}

// formatScientific formats mantissa and exponent as scientific notation string.
func formatScientific(mantissa int64, exponent int32) string {
	m := big.NewInt(mantissa)
	if exponent >= 0 {
		return m.String() + "e" + itoa(int(exponent))
	}
	return m.String() + "e" + itoa(int(exponent))
}

// formatDecimal formats mantissa and exponent as a decimal string.
func formatDecimal(mantissa int64, exponent int32) string {
	isNegative := mantissa < 0
	if isNegative {
		mantissa = -mantissa
	}

	mantissaStr := big.NewInt(mantissa).String()

	// Pad with zeros for proper decimal placement
	const padPrefix = 27
	const padSuffix = 23
	rawValue := strings.Repeat("0", padPrefix) + mantissaStr + strings.Repeat("0", padSuffix)

	offset := int(exponent) + 43
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
	binary.BigEndian.PutUint64(buf[offset:], uint64(v))
}

func writeInt32BE(buf []byte, v int32, offset int) {
	binary.BigEndian.PutUint32(buf[offset:], uint32(v))
}

func readInt64BE(buf []byte, offset int) int64 {
	return int64(binary.BigEndian.Uint64(buf[offset:]))
}

func readInt32BE(buf []byte, offset int) int32 {
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
