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
	uint64JSONBaseDecimal int = 10
	uint64JSONBaseHex     int = 16
)

// uint64JSONBaseForField mirrors rippled's kSmdBaseTen metadata, which is not
// included in the server definitions copied from xrpl.js.
func uint64JSONBaseForField(fieldName string) int {
	switch fieldName {
	case "MaximumAmount", "OutstandingAmount", "MPTAmount", "LockedAmount", "ConfidentialOutstandingAmount":
		return uint64JSONBaseDecimal
	default:
		return uint64JSONBaseHex
	}
}

// UInt64 represents a 64-bit unsigned integer serialized from a hex JSON string.
type UInt64 struct{}

// ErrInvalidUInt64String is returned when a value is not a valid UInt64 hex string.
var ErrInvalidUInt64String = errors.New("invalid UInt64 string, value should be a 1 to 16 character hex string")

// FromJSON converts a JSON value into a serialized byte slice representing a 64-bit unsigned integer.
// The input value must be a 1 to 16 character hex string.
//
// Note: decimal-looking inputs are parsed as hex. "10" is read as 0x10 (= 16),
// not decimal 10. Callers wanting decimal semantics must hex-encode first.
//
// Returns ErrInvalidUInt64String when the input is not a string, contains non-hex
// characters, or exceeds 16 characters.
func (u *UInt64) FromJSON(value any) ([]byte, error) {
	return u.fromJSON(value, uint64JSONBaseHex)
}

func (u *UInt64) fromJSON(value any, base int) ([]byte, error) {
	strValue, ok := value.(string)
	if !ok {
		return nil, ErrInvalidUInt64String
	}

	if base == uint64JSONBaseHex && (len(strValue) > 16 || !typecheck.IsHex(strValue)) {
		return nil, ErrInvalidUInt64String
	}

	parsed, err := strconv.ParseUint(strValue, base, 64)
	if err != nil {
		var numErr *strconv.NumError
		if errors.As(err, &numErr) && errors.Is(numErr.Err, strconv.ErrRange) {
			return nil, ErrInvalidUInt64String
		}
		return nil, err
	}

	serialized := make([]byte, 8)
	binary.BigEndian.PutUint64(serialized, parsed)
	return serialized, nil
}

// ToJSON takes a BinaryParser and optional parameters, and converts the serialized byte data
// back into a JSON string value. This method assumes the parser contains data representing
// a 64-bit unsigned integer. If the parsing fails, an error is returned.
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
