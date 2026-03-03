//revive:disable:var-naming
package types

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// ErrInvalidHashLength struct is used when the hash length does not meet the expected value.
type ErrInvalidHashLength struct {
	Expected int
}

// ErrInvalidHashType indicates the provided JSON value is not a valid hash type.
type ErrInvalidHashType struct{}

// ErrInvalidHexString indicates an error occurred decoding a hex string.
type ErrInvalidHexString struct {
	Err error
}

// Error method for ErrInvalidHashLength formats the error message.
func (e *ErrInvalidHashLength) Error() string {
	return fmt.Sprintf("invalid hash length expected length %v", e.Expected)
}

// Error method for ErrInvalidHashType formats the error message.
func (e *ErrInvalidHashType) Error() string {
	return "invalid hash type"
}

// Error method for ErrInvalidHexString formats the error message.
func (e *ErrInvalidHexString) Error() string {
	return "error decoding hex string: " + e.Err.Error()
}

// hashI interface combines the SerializedType interface and getLength method for hashes.
type hashI interface {
	SerializedType
	getLength() int
}

// hash struct represents a hash with a specific length.
type hash struct {
	Length int
}

// newHash is a constructor for creating a new hash with a specified length.
func newHash(l int) hash {
	return hash{
		Length: l,
	}
}

// getLength method for hash returns the hash's length.
func (h hash) getLength() int {
	return h.Length
}

// FromJSON method for hash converts a hexadecimal string from JSON to a byte array.
// It returns an error if the conversion fails or the length of the decoded byte array is not as expected.
func (h hash) FromJSON(json any) ([]byte, error) {
	v, ok := json.(string)
	if !ok {
		return nil, &ErrInvalidHashType{}
	}
	decoded, err := hex.DecodeString(v)
	if err != nil {
		return nil, &ErrInvalidHexString{Err: err}
	}
	if h.getLength() != len(decoded) {
		return nil, &ErrInvalidHashLength{Expected: h.getLength()}
	}
	return decoded, nil
}

// ToJSON method for hash reads a certain number of bytes from a BinaryParser and converts it into a hexadecimal string.
// It returns an error if the read operation fails.
func (h hash) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := p.ReadBytes(h.Length)
	if err != nil {
		return nil, err
	}
	return hexutil.EncodeToUpperHex(b), nil
}
