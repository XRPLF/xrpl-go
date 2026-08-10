//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
)

// UInt16 represents a 16-bit unsigned integer.
type UInt16 struct{}

// checkRange validates that a value fits within the uint16 range (0-65535).
func (u *UInt16) checkRange(value int64) error {
	if value < 0 || value > int64(math.MaxUint16) {
		return ErrUInt16OutOfRange
	}
	return nil
}

// FromJSON converts a JSON value into a serialized byte slice representing a 16-bit unsigned integer.
// If the input value has a string underlying type, it is treated as a transaction type or ledger entry type name.
// The method returns an error when the name does not have a corresponding type code.
func (u *UInt16) FromJSON(value any) ([]byte, error) {
	if stringValue, ok := typecheck.ToString(value); ok {
		tc, err := definitions.Get().GetTransactionTypeCodeByTransactionTypeName(stringValue)
		if err != nil {
			tc, err = definitions.Get().GetLedgerEntryTypeCodeByLedgerEntryTypeName(stringValue)
			if err != nil {
				return nil, err
			}
		}
		value = int(tc)
	}

	var int64Value int64

	switch v := value.(type) {
	case int:
		int64Value = int64(v)
	case int64:
		int64Value = v
	case uint16:
		int64Value = int64(v)
	case uint32:
		int64Value = int64(v)
	case float64:
		// Check if float64 represents a whole number
		if v != float64(int64(v)) {
			return nil, ErrUInt16OutOfRange
		}
		int64Value = int64(v)
	default:
		return nil, ErrUInt16OutOfRange
	}

	// Check range before casting
	if err := u.checkRange(int64Value); err != nil {
		return nil, err
	}

	//nolint:gosec // G115: integer overflow conversion int64 -> uint16 (gosec)
	val := uint16(int64Value)
	buf := new(bytes.Buffer)
	//nolint:gosec // G115 false positive, binary.Write with uint16 value
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
