//revive:disable:var-naming
package types

import (
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// Int64 represents a 64-bit signed integer.
type Int64Type struct{}

// ErrInvalidInt64 is returned when a value cannot be converted to Int64.
var ErrInvalidInt64 = errors.New("invalid Int64 value")

// FromJSON converts a JSON value into a serialized byte slice representing a 64-bit signed integer.
// The input value can be an int, int64, float64, or string representation of an integer.
func (i *Int64Type) FromJSON(value any) ([]byte, error) {
	var v int64

	switch val := value.(type) {
	case int:
		v = int64(val)
	case int64:
		v = val
	case float64:
		v = int64(val)
	case string:
		var err error
		v, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, ErrInvalidInt64
		}
	default:
		return nil, ErrInvalidInt64
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	return buf, nil
}

// ToJSON takes a BinaryParser and converts the serialized byte data back to a JSON string value.
// Int64 values are returned as strings to preserve precision in JSON.
func (i *Int64Type) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(8)
	if err != nil {
		return nil, err
	}

	v := int64(binary.BigEndian.Uint64(b))
	return strconv.FormatInt(v, 10), nil
}
