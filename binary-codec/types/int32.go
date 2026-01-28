//revive:disable:var-naming
package types

import (
	"encoding/binary"
	"errors"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// Int32 represents a 32-bit signed integer.
type Int32 struct{}

// ErrInvalidInt32 is returned when a value cannot be converted to Int32.
var ErrInvalidInt32 = errors.New("invalid Int32 value")

// FromJSON converts a JSON value into a serialized byte slice representing a 32-bit signed integer.
// The input value can be an int, int32, int64, or float64.
func (i *Int32) FromJSON(value any) ([]byte, error) {
	var v int32

	switch val := value.(type) {
	case int:
		v = int32(val)
	case int32:
		v = val
	case int64:
		v = int32(val)
	case float64:
		v = int32(val)
	default:
		return nil, ErrInvalidInt32
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(v))
	return buf, nil
}

// ToJSON takes a BinaryParser and converts the serialized byte data back to a JSON integer value.
func (i *Int32) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(4)
	if err != nil {
		return nil, err
	}

	v := int32(binary.BigEndian.Uint32(b))
	return int(v), nil
}
