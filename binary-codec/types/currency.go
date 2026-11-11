//revive:disable:var-naming
package types

import (
	"encoding/hex"
	"errors"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// currencyBytesLength is the number of bytes of a serialized currency.
const currencyBytesLength = 20

var (
	// ErrMissingCurrencyLengthOption is not returned by any function.
	//
	// Deprecated: Currency.ToJSON does not need a length option.
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

// ToJSON reads a 20-byte currency from the parser and returns its JSON string.
// A currency has a fixed length, so the opts parameter is ignored.
// https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types/oracle#currency-internal-format
func (c *Currency) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	currencyBytes, err := p.ReadBytes(currencyBytesLength)
	if err != nil {
		return nil, err
	}

	return decodeCurrencyBytes(currencyBytes), nil
}

func (c *Currency) fromString(str string) ([]byte, error) {
	if len(str) == 3 {
		var bytes [currencyBytesLength]byte
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
	if len(bytes) != currencyBytesLength {
		return nil, ErrInvalidCurrency
	}

	return bytes, nil
}
