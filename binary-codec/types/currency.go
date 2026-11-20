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
	// ErrMissingCurrencyLengthOption is retained for compatibility.
	//
	// Deprecated: Currency.ToJSON reads the fixed-width currency code.
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

// ToJSON reads the 20-byte XRPL currency code and returns its JSON representation.
func (c *Currency) ToJSON(p interfaces.BinaryParser, _ DecodeContext) (any, error) {
	currencyBytes, err := p.ReadBytes(20)
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
