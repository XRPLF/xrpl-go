package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// Number represents a legacy XRPL number type (also known as STNumber).
// This type is deprecated in XRPL but still appears in historical ledgers.
// It encodes as 12 bytes: 8-byte signed mantissa + 4-byte signed exponent.
type Number struct{}

var ErrInvalidNumber = errors.New("invalid Number value")

const (
	// XRPL Number format constants
	numberMinMantissa  = 1000000000000000 // 10^15
	numberMaxMantissa  = 9999999999999999 // 10^16 - 1
	numberMinExponent  = -32768           // min signed 16-bit
	numberMaxExponent  = 32768            // max signed 16-bit
	numberZeroExponent = -2147483648      // special exponent for zero
	numberByteLength   = 12               // 8 bytes mantissa + 4 bytes exponent
)

var numberRegex = regexp.MustCompile(`^([+-]?)(\d*)(?:\.(\d*))?(?:[eE]([+-]?\d+))?$`)

// FromJSON converts a JSON value into a serialized byte slice representing a Number.
// The input value should be a string representation of a number.
func (n *Number) FromJSON(value any) ([]byte, error) {
	strValue, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: must be a string", ErrInvalidNumber)
	}

	// Parse the number string
	mantissa, exponent, err := parseNumberString(strValue)
	if err != nil {
		return nil, err
	}

	// Normalize mantissa and exponent to XRPL constraints
	mantissa, exponent, err = normalizeNumber(mantissa, exponent)
	if err != nil {
		return nil, err
	}

	// Encode as 12 bytes: 8-byte mantissa (big-endian) + 4-byte exponent (big-endian)
	buf := make([]byte, numberByteLength)
	binary.BigEndian.PutUint64(buf[0:8], uint64(mantissa))
	binary.BigEndian.PutUint32(buf[8:12], uint32(exponent))

	return buf, nil
}

// ToJSON takes a BinaryParser and converts the serialized byte data
// back into a JSON string value. The Number type uses 12 bytes.
func (n *Number) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(numberByteLength)
	if err != nil {
		return nil, err
	}

	// Decode mantissa (8 bytes) and exponent (4 bytes)
	mantissa := int64(binary.BigEndian.Uint64(b[0:8]))
	exponent := int32(binary.BigEndian.Uint32(b[8:12]))

	// Handle zero case
	if exponent == numberZeroExponent {
		return "0", nil
	}

	// Convert to string representation
	return formatNumber(mantissa, exponent), nil
}

// parseNumberString parses a string representation of a number and returns mantissa and exponent.
func parseNumberString(s string) (int64, int32, error) {
	s = strings.TrimSpace(s)
	if s == "0" || s == "" {
		return 0, numberZeroExponent, nil
	}

	matches := numberRegex.FindStringSubmatch(s)
	if matches == nil {
		return 0, 0, fmt.Errorf("%w: invalid format", ErrInvalidNumber)
	}

	sign := matches[1]
	intPart := matches[2]
	fracPart := matches[3]
	expPart := matches[4]

	// Combine integer and fractional parts
	numStr := intPart + fracPart
	if numStr == "" {
		return 0, numberZeroExponent, nil
	}

	// Parse as big.Int to handle large numbers
	mantissa := new(big.Int)
	_, ok := mantissa.SetString(numStr, 10)
	if !ok {
		return 0, 0, fmt.Errorf("%w: cannot parse mantissa", ErrInvalidNumber)
	}

	// Calculate initial exponent based on decimal point position
	exponent := int32(len(intPart)) - int32(len(numStr))

	// Adjust for explicit exponent if present
	if expPart != "" {
		exp, err := strconv.ParseInt(expPart, 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("%w: invalid exponent", ErrInvalidNumber)
		}
		exponent += int32(exp)
	}

	// Apply sign
	if sign == "-" {
		mantissa.Neg(mantissa)
	}

	// Convert to int64 (may need normalization)
	if !mantissa.IsInt64() {
		return 0, 0, fmt.Errorf("%w: mantissa overflow", ErrInvalidNumber)
	}

	return mantissa.Int64(), exponent, nil
}

// normalizeNumber adjusts mantissa and exponent to fit XRPL Number constraints.
func normalizeNumber(mantissa int64, exponent int32) (int64, int32, error) {
	if mantissa == 0 {
		return 0, numberZeroExponent, nil
	}

	// Get absolute value for normalization
	absMantissa := mantissa
	if absMantissa < 0 {
		absMantissa = -absMantissa
	}

	// Normalize: shift mantissa to be within valid range
	for absMantissa > 0 && absMantissa < numberMinMantissa {
		absMantissa *= 10
		exponent--
		if exponent < numberMinExponent {
			return 0, 0, fmt.Errorf("%w: exponent underflow", ErrInvalidNumber)
		}
	}

	for absMantissa > numberMaxMantissa {
		absMantissa /= 10
		exponent++
		if exponent > numberMaxExponent {
			return 0, 0, fmt.Errorf("%w: exponent overflow", ErrInvalidNumber)
		}
	}

	// Reapply sign
	if mantissa < 0 {
		absMantissa = -absMantissa
	}

	return absMantissa, exponent, nil
}

// formatNumber converts mantissa and exponent back to string representation.
func formatNumber(mantissa int64, exponent int32) string {
	if exponent == numberZeroExponent {
		return "0"
	}

	// Determine if we should use decimal or scientific notation
	// Use decimal notation for exponents between -25 and -5 (like XRPL.js)
	useDecimal := exponent >= -25 && exponent <= -5

	mantissaStr := strconv.FormatInt(mantissa, 10)
	sign := ""
	if mantissa < 0 {
		sign = "-"
		mantissaStr = mantissaStr[1:] // Remove negative sign
	}

	if useDecimal {
		// Decimal notation
		totalExp := int(exponent) + len(mantissaStr)
		if totalExp <= 0 {
			// Add leading zeros: 0.000...mantissa
			return sign + "0." + strings.Repeat("0", -totalExp) + mantissaStr
		} else if totalExp < len(mantissaStr) {
			// Insert decimal point: mantissa[0:totalExp].mantissa[totalExp:]
			return sign + mantissaStr[:totalExp] + "." + mantissaStr[totalExp:]
		} else {
			// Add trailing zeros: mantissa000...
			return sign + mantissaStr + strings.Repeat("0", totalExp-len(mantissaStr))
		}
	}

	// Scientific notation
	if len(mantissaStr) == 1 {
		return sign + mantissaStr + "e" + strconv.FormatInt(int64(exponent), 10)
	}
	return sign + mantissaStr[0:1] + "." + mantissaStr[1:] + "e" + strconv.FormatInt(int64(exponent)+int64(len(mantissaStr))-1, 10)
}
