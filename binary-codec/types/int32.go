//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// Int32 represents a 32-bit signed integer.
type Int32 struct{}

// checkRange validates that a value fits within the int32 range (-2147483648 to 2147483647).
func (i *Int32) checkRange(value int64) error {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return ErrInt32OutOfRange
	}
	return nil
}

// FromJSON converts a JSON value into a serialized byte slice representing a 32-bit signed integer.
// The input value can be int, int32, int64, or float64 (if it represents a whole number).
// Negative values are encoded using two's complement representation.
// If the serialization fails or the value is out of range, an error is returned.
func (i *Int32) FromJSON(value any) ([]byte, error) {
	var val int32

	switch v := value.(type) {
	case int32:
		val = v
	case int:
		int64Value := int64(v)
		if err := i.checkRange(int64Value); err != nil {
			return nil, err
		}
		//nolint:gosec // G115: integer overflow conversion int64 -> int32 (gosec)
		val = int32(int64Value)
	case int64:
		if err := i.checkRange(v); err != nil {
			return nil, err
		}
		//nolint:gosec // G115: integer overflow conversion int64 -> int32 (gosec)
		val = int32(v)
	case uint32:
		// Check if uint32 value fits in int32 range
		if v > math.MaxInt32 {
			return nil, ErrInt32OutOfRange
		}
		//nolint:gosec // G115: integer overflow conversion uint32 -> int32 (gosec)
		val = int32(v)
	case uint64:
		// Check if uint64 value fits in int32 range
		if v > math.MaxInt32 {
			return nil, ErrInt32OutOfRange
		}
		//nolint:gosec // G115: integer overflow conversion uint64 -> int32 (gosec)
		val = int32(v)
	case float64:
		// Check if float64 represents a whole number
		intV := int64(v)
		if v != float64(intV) {
			return nil, ErrInt32OutOfRange
		}
		if err := i.checkRange(intV); err != nil {
			return nil, err
		}
		//nolint:gosec // G115: integer overflow conversion int64 -> int32 (gosec)
		val = int32(intV)
	default:
		return nil, ErrInt32OutOfRange
	}

	// Use BigEndian encoding for consistency with XRPL binary format
	// Two's complement representation handles negative numbers automatically
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, val)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ToJSON takes a BinaryParser and optional parameters, and converts the serialized byte data
// back into a JSON integer value. This method assumes the parser contains data representing
// a 32-bit signed integer in two's complement format.
// If the parsing fails, an error is returned.
func (i *Int32) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(4)
	if err != nil {
		return nil, err
	}

	// Read as signed int32 using BigEndian byte order
	// This automatically handles two's complement for negative numbers
	//
	// How it works: binary.BigEndian.Uint32() reads 4 bytes as unsigned,
	// then int32() cast reinterprets those bits as signed (two's complement).
	//
	// Examples:
	// Bytes                    | As Uint32    | As Int32      | Represents
	// [0x00, 0x00, 0x00, 0x01] | 1            | 1             | Positive 1
	// [0x7F, 0xFF, 0xFF, 0xFF] | 2147483647   | 2147483647    | INT32_MAX
	// [0x80, 0x00, 0x00, 0x00] | 2147483648   | -2147483648   | INT32_MIN
	// [0xFF, 0xFF, 0xFF, 0xFF] | 4294967295   | -1            | Negative 1
	// [0xFF, 0xFF, 0xFF, 0x9C] | 4294967196   | -100          | Negative 100
	//nolint:gosec // G115: integer overflow conversion uint32 -> int32 (gosec)
	val := int32(binary.BigEndian.Uint32(b))

	return val, nil
}
