//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

var (
	// ErrMissingCurrencyLengthOption is returned when no length option is
	// provided to Currency.ToJSON.
	ErrMissingCurrencyLengthOption = errors.New("missing length option for Currency.ToJSON")
	// ErrInvalidCurrency is returned when the currency field is missing or invalid.
	ErrInvalidCurrency = errors.New("invalid currency")
)

// Currency handles encoding and decoding of currency values in the binary codec.
type Currency struct{}

// FromJSON parses a JSON value into its binary currency representation.
func (c *Currency) FromJSON(json any) ([]byte, error) {
	if str, ok := json.(string); ok {
		return c.fromString(str)
	}
	return nil, ErrInvalidCurrency
}

// ToJSON serializes a binary currency value into a JSON-compatible format.
// It requires a length option specifying the byte length to read.
func (c *Currency) ToJSON(p interfaces.BinaryParser, opts ...int) (any, error) {
	// default to 20 bytes, https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/oracle#currency-internal-format
	length := 20
	if len(opts) > 0 && opts[0] > 0 {
		length = opts[0]
	}

	currencyBytes, err := p.ReadBytes(length)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(currencyBytes, XRPBytes) {
		return "XRP", nil
	}

	// Check if bytes has exactly 3 non-zero bytes at positions 12-14
	nonZeroCount := 0
	var currencyStr strings.Builder
	for i := range currencyBytes {
		if currencyBytes[i] != 0 {
			if i >= 12 && i <= 14 {
				nonZeroCount++
				currencyStr.WriteString(string(currencyBytes[i]))
			} else {
				nonZeroCount = 0
				break
			}
		}
	}

	if nonZeroCount == 3 {
		return currencyStr.String(), nil
	}

	return hex.EncodeToString(currencyBytes), nil
}

func (c *Currency) fromString(str string) ([]byte, error) {
	if len(str) == 3 {
		var bytes [20]byte
		if str != "XRP" {
			isoBytes := []byte(str)
			copy(bytes[12:], isoBytes)
		}
		return bytes[:], nil
	}

	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
