//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/binary"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// UInt32 represents a 32-bit unsigned integer.
type UInt32 struct{}

// FromJSON converts a JSON value into a serialized byte slice representing a 32-bit unsigned integer.
// The input value is assumed to be an integer. If the serialization fails, an error is returned.
func (u *UInt32) FromJSON(value any) ([]byte, error) {
	var val uint32
	switch v := value.(type) {
	case uint32:
		val = v
	case int:
		val = uint32(v)
	case int64:
		val = uint32(v)
	case uint64:
		val = uint32(v)
	case float64:
		val = uint32(v)
	default:
		val = value.(uint32) // will panic with meaningful error
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
