//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// UInt32 represents a 32-bit unsigned integer.
type UInt32 struct{}

// checkRange validates that a value fits within the uint32 range (0-4294967295).
func (u *UInt32) checkRange(value int64) error {
	if value < 0 || value > int64(math.MaxUint32) {
		return ErrUInt32OutOfRange
	}
	return nil
}

// checkRangeUint64 validates that a uint64 value fits within the uint32 range.
func (u *UInt32) checkRangeUint64(value uint64) error {
	if value > uint64(math.MaxUint32) {
		return ErrUInt32OutOfRange
	}
	return nil
}

// FromJSON converts a JSON value into a serialized byte slice representing a 32-bit unsigned integer.
// The input value is assumed to be an integer. If the serialization fails, an error is returned.
func (u *UInt32) FromJSON(value any) ([]byte, error) {
	var val uint32

	switch v := value.(type) {
	case uint32:
		val = v
	case int:
		int64Value := int64(v)
		if err := u.checkRange(int64Value); err != nil {
			return nil, err
		}
		//nolint:gosec // G115: integer overflow conversion int64 -> uint32 (gosec)
		val = uint32(int64Value)
	case int64:
		if err := u.checkRange(v); err != nil {
			return nil, err
		}
		//nolint:gosec // G115: integer overflow conversion int64 -> uint32 (gosec)
		val = uint32(v)
	case uint64:
		if err := u.checkRangeUint64(v); err != nil {
			return nil, err
		}
		//nolint:gosec // G115: integer overflow conversion uint64 -> uint32 (gosec)
		val = uint32(v)
	case float64:
		// Check if float64 represents a whole number
		if v != float64(int64(v)) {
			return nil, ErrUInt32OutOfRange
		}
		int64Value := int64(v)
		if err := u.checkRange(int64Value); err != nil {
			return nil, err
		}
		//nolint:gosec // G115: integer overflow conversion int64 -> uint32 (gosec)
		val = uint32(int64Value)
	default:
		return nil, ErrUInt32OutOfRange
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, val)

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ToJSON takes a BinaryParser and optional parameters, and converts the serialized byte data
// back into a JSON integer value. This method assumes the parser contains data representing
// a 32-bit unsigned integer. If the parsing fails, an error is returned.
func (u *UInt32) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(4)
	if err != nil {
		return nil, err
	}
	return binary.BigEndian.Uint32(b), nil
}
