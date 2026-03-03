//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// UInt64 represents a 64-bit unsigned integer.
type UInt64 struct{}

// ErrInvalidUInt64String is returned when a value is not a valid string representation of a UInt64.
var ErrInvalidUInt64String = errors.New("invalid UInt64 string, value should be a string representation of a UInt64")

// checkRange validates that a numeric string represents a value that fits within the uint64 range.
func (u *UInt64) checkRange(numericStr string) error {
	// ParseUint with bitSize 64 already validates the range (0 to max uint64)
	_, err := strconv.ParseUint(numericStr, 10, 64)
	if err != nil {
		// Check if it's an overflow/underflow error
		if numErr, ok := err.(*strconv.NumError); ok && numErr.Err == strconv.ErrRange {
			return ErrUInt64OutOfRange
		}
		return err
	}
	return nil
}

// FromJSON converts a JSON value into a serialized byte slice representing a 64-bit unsigned integer.
// The input value is assumed to be a string representation of an integer. If the serialization fails, an error is returned.
func (u *UInt64) FromJSON(value any) ([]byte, error) {

	var buf = new(bytes.Buffer)

	if _, ok := value.(string); !ok {
		return nil, ErrInvalidUInt64String
	}

	strValue := value.(string)

	if !isNumeric(strValue) {
		// Handle hex strings - validate they don't exceed 16 hex characters (8 bytes)
		hexStr := strings.ToUpper(strValue)
		if len(hexStr) > 16 {
			return nil, ErrUInt64OutOfRange
		}
		hexBytes, err := hex.DecodeString(hexStr)
		if err != nil {
			return nil, err
		}
		// Ensure the decoded hex doesn't exceed 8 bytes
		if len(hexBytes) > 8 {
			return nil, ErrUInt64OutOfRange
		}
		buf.Write(hexBytes)
		return buf.Bytes(), nil
	}

	// Validate numeric string fits in uint64 range
	if err := u.checkRange(strValue); err != nil {
		return nil, err
	}

	// Right justify the string to 16 hex characters (8 bytes)
	strValue = strings.Repeat("0", 16-len(strValue)) + strValue
	decoded, err := hex.DecodeString(strValue)
	if err != nil {
		return nil, err
	}
	buf.Write(decoded)

	return buf.Bytes(), nil
}

// ToJSON takes a BinaryParser and optional parameters, and converts the serialized byte data
// back into a JSON string value. This method assumes the parser contains data representing
// a 64-bit unsigned integer. If the parsing fails, an error is returned.
func (u *UInt64) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(8)
	if err != nil {
		return nil, err
	}
	return hexutil.EncodeToUpperHex(b), nil
}

// isNumeric checks if a string only contains numerical values.
func isNumeric(s string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", s)
	return match
}
