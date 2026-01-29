//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// UInt8 represents an 8-bit unsigned integer.
type UInt8 struct{}

// checkRange validates that a value fits within the uint8 range (0-255).
func (u *UInt8) checkRange(value int64) error {
	if value < 0 || value > int64(math.MaxUint8) {
		return ErrUInt8OutOfRange
	}
	return nil
}

// FromJSON converts a JSON value into a serialized byte slice representing an 8-bit unsigned integer.
// If the input value is a string, it's assumed to be a transaction result name, and the method will
// attempt to convert it into a transaction result type code. If the conversion fails, an error is returned.
func (u *UInt8) FromJSON(value any) ([]byte, error) {
	if s, ok := value.(string); ok {
		tc, err := definitions.Get().GetTransactionResultTypeCodeByTransactionResultName(s)
		if err != nil {
			return nil, err
		}
		value = tc
	}

	var int64Value int64

	switch v := value.(type) {
	case int:
		int64Value = int64(v)
	case int32:
		int64Value = int64(v)
	case int64:
		int64Value = v
	case uint8:
		int64Value = int64(v)
	case float64:
		// Check if float64 represents a whole number
		if v != float64(int64(v)) {
			return nil, ErrUInt8OutOfRange
		}
		int64Value = int64(v)
	default:
		return nil, ErrUInt8OutOfRange
	}

	// Check range before casting
	if err := u.checkRange(int64Value); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	// TODO: Check if this is still needed
	err := binary.Write(buf, binary.BigEndian, byte(int64Value))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ToJSON takes a BinaryParser and optional parameters, and converts the serialized byte data
// back into a JSON integer value. This method assumes the parser contains data representing
// an 8-bit unsigned integer. If the parsing fails, an error is returned.
func (u *UInt8) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(1)
	if err != nil {
		return nil, err
	}
	return int(b[0]), nil
}
