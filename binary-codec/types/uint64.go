//revive:disable:var-naming
package types

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
)

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
	strValue, ok := value.(string)
	if !ok {
		return nil, ErrInvalidUInt64String
	}

	if len(strValue) > 16 || !typecheck.IsHex(strValue) {
		return nil, ErrInvalidUInt64String
	}

	// Right justify the string to 16 hex characters (8 bytes)
	strValue = strings.Repeat("0", 16-len(strValue)) + strValue
	decoded, err := hex.DecodeString(strValue)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

// ToJSON takes a BinaryParser and optional parameters, and converts the serialized byte data
// back into a JSON string value. This method assumes the parser contains data representing
// a 64-bit unsigned integer. If the parsing fails, an error is returned.
func (u *UInt64) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(8)
	if err != nil {
		return nil, err
	}
	return hexutil.EncodeToUpperHex(b), nil
}
