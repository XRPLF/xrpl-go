//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/binary"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// UInt16 represents a 16-bit unsigned integer.
type UInt16 struct{}

// FromJSON converts a JSON value into a serialized byte slice representing a 16-bit unsigned integer.
// If the input value is a string, it's assumed to be a transaction type or ledger entry type name, and the
// method will attempt to convert it into a corresponding type code. If the conversion fails, an error is returned.
func (u *UInt16) FromJSON(value any) ([]byte, error) {

	if _, ok := value.(string); ok {
		tc, err := definitions.Get().GetTransactionTypeCodeByTransactionTypeName(value.(string))
		if err != nil {
			tc, err = definitions.Get().GetLedgerEntryTypeCodeByLedgerEntryTypeName(value.(string))
			if err != nil {
				return nil, err
			}
		}
		value = int(tc)
	}

	var val uint16
	switch v := value.(type) {
	case int:
		val = uint16(v)
	case int64:
		val = uint16(v)
	case uint16:
		val = v
	case uint32:
		val = uint16(v)
	case float64:
		val = uint16(v)
	default:
		val = uint16(value.(int)) // will panic with meaningful error
	}

	buf := new(bytes.Buffer)
	//nolint:gosec // G115: Potential hardcoded credentials (gosec)
	err := binary.Write(buf, binary.BigEndian, val)

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ToJSON takes a BinaryParser and optional parameters, and converts the serialized byte data
// back into a JSON integer value. This method assumes the parser contains data representing
// a 16-bit unsigned integer. If the parsing fails, an error is returned.
func (u *UInt16) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(2)
	if err != nil {
		return nil, err
	}
	return int(binary.BigEndian.Uint16(b)), nil
}
