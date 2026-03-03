//revive:disable:var-naming
package types

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// HashLengthBytes is the byte length of a hash in Vector256.
const HashLengthBytes = 32

// ErrInvalidVector256Type represents an error when a Vector256 is constructed from an unexpected type.
type ErrInvalidVector256Type struct {
	Got string
}

// Error implements the error interface, providing a descriptive error message for ErrInvalidVector256Type.
func (e *ErrInvalidVector256Type) Error() string {
	return fmt.Sprintf("Invalid type to construct Vector256 from. Expected []string, got %v", e.Got)
}

// Vector256 represents a 256 bit vector.
type Vector256 struct{}

// FromJSON converts a JSON value into a serialized byte slice representing a Vector256.
// The input value is assumed to be an array of strings representing Hash256 values.
// If the serialization fails, an error is returned.
func (v *Vector256) FromJSON(json any) ([]byte, error) {
	var strSlice []string

	switch val := json.(type) {
	case []string:
		strSlice = val
	case []any:
		// Convert []interface{} to []string (common when unmarshalling JSON)
		strSlice = make([]string, len(val))
		for i, item := range val {
			s, ok := item.(string)
			if !ok {
				return nil, &ErrInvalidVector256Type{fmt.Sprintf("%T (element %d is %T)", json, i, item)}
			}
			strSlice[i] = s
		}
	default:
		return nil, &ErrInvalidVector256Type{fmt.Sprintf("%T", json)}
	}

	b, err := vector256FromValue(strSlice)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// vector256FromValue takes a slice of strings representing Hash256 values,
// serializes them, and returns the combined byte slice. If an error occurs during serialization, it is returned.
func vector256FromValue(value []string) ([]byte, error) {
	b := make([]byte, 0)
	for _, s := range value {
		hash256, err := NewHash256().FromJSON(s)

		if err != nil {
			return nil, err
		}

		b = append(b, hash256...)

	}
	return b, nil
}

// ToJSON takes a BinaryParser and optional parameters, and converts the serialized byte data
// back into an array of JSON string values representing Hash256 values.
// If the parsing fails, an error is returned.
func (v *Vector256) ToJSON(p interfaces.BinaryParser, opts ...int) (any, error) {

	b, err := p.ReadBytes(opts[0])
	if err != nil {
		return nil, err
	}
	var value []string

	for i := 0; i < len(b); i += HashLengthBytes {
		value = append(value, hexutil.EncodeToUpperHex(b[i:i+HashLengthBytes]))
	}

	return value, nil
}
