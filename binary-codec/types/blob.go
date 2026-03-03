//revive:disable:var-naming
package types

import (
	"encoding/hex"
	"errors"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// ErrNoLengthPrefix error is raised when no length prefix size is given.
var ErrNoLengthPrefix = errors.New("no length prefix size given")

// Blob struct is used for manipulating hexadecimal data.
type Blob struct{}

// FromJSON method for Blob converts a hexadecimal string from JSON to a byte array.
func (b *Blob) FromJSON(json any) ([]byte, error) {
	// Convert hexadecimal string to byte array.
	// Return an error if the conversion fails.
	v, err := hex.DecodeString(json.(string))
	if err != nil {
		return nil, err
	}
	return v, nil
}

// ToJSON method for Blob reads a certain number of bytes from a BinaryParser
// and converts it into a hexadecimal string.
// It returns an error if no length prefix is specified or if the read operation fails.
func (b *Blob) ToJSON(p interfaces.BinaryParser, opts ...int) (any, error) {
	// If no length prefix is specified, return an error.
	if opts == nil {
		return nil, ErrNoLengthPrefix
	}
	// Read the specified number of bytes.
	// If the read operation fails, return an error.
	val, err := p.ReadBytes(opts[0])
	if err != nil {
		return nil, err
	}
	// Convert the bytes to a hexadecimal string and return it.
	return hexutil.EncodeToUpperHex(val), nil
}
