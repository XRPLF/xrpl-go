//revive:disable:var-naming
package types

import (
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
)

const (
	uint64JSONBaseDecimal = 10
	uint64JSONBaseHex     = 16
)

// uint64JSONBaseForField mirrors rippled's kSmdBaseTen field metadata, which is
// not available in the definitions consumed by the codec.
func uint64JSONBaseForField(fieldName string) int {
	switch fieldName {
	case "MaximumAmount", "OutstandingAmount", "MPTAmount", "LockedAmount":
		return uint64JSONBaseDecimal
	default:
		return uint64JSONBaseHex
	}
}

// UInt64 represents a 64-bit unsigned integer. Its JSON radix depends on the
// field: MPT amount fields use decimal strings, while other fields use hex.
type UInt64 struct{}

// ErrInvalidUInt64String is returned when a value is not a valid UInt64 string
// for the field's JSON radix.
var ErrInvalidUInt64String = errors.New("invalid UInt64 string")

// FromJSON converts a hexadecimal JSON string into its 8-byte UInt64 representation.
// Field-aware STObject serialization uses decimal parsing for MPT amount fields.
func (u *UInt64) FromJSON(value any) ([]byte, error) {
	return u.fromJSON(value, uint64JSONBaseHex)
}

func (u *UInt64) fromJSON(value any, base int) ([]byte, error) {
	strValue, ok := value.(string)
	if !ok || strValue == "" {
		return nil, ErrInvalidUInt64String
	}

	switch base {
	case uint64JSONBaseDecimal:
		for i := range len(strValue) {
			if strValue[i] < '0' || strValue[i] > '9' {
				return nil, ErrInvalidUInt64String
			}
		}
	case uint64JSONBaseHex:
		if len(strValue) > 16 || !typecheck.IsHex(strValue) {
			return nil, ErrInvalidUInt64String
		}
	default:
		return nil, ErrInvalidUInt64String
	}

	parsed, err := strconv.ParseUint(strValue, base, 64)
	if err != nil {
		return nil, ErrInvalidUInt64String
	}

	serialized := make([]byte, 8)
	binary.BigEndian.PutUint64(serialized, parsed)
	return serialized, nil
}

// ToJSON reads an 8-byte UInt64 and returns its JSON string representation.
// The optional base is used internally for field-aware decimal MPT amounts.
// Direct calls retain the canonical hexadecimal representation.
func (u *UInt64) ToJSON(p interfaces.BinaryParser, opts ...int) (any, error) {
	b, err := p.ReadBytes(8)
	if err != nil {
		return nil, err
	}
	if len(opts) > 0 && opts[0] == uint64JSONBaseDecimal {
		return strconv.FormatUint(binary.BigEndian.Uint64(b), uint64JSONBaseDecimal), nil
	}
	return hexutil.EncodeToUpperHex(b), nil
}
