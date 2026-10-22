//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// CurrencyCodeByteLength is the length of a currency code on the ledger.
const CurrencyCodeByteLength = 20

var iouCodeRegex = regexp.MustCompile("^" + IOUCodeRegex + "$")

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
		return ParseCurrencyCode(str)
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

// ParseCurrencyCode converts a JSON currency code to its 20-byte form, following
// rippled's to_currency. "XRP" is the native currency and encodes as 20 zero bytes.
// A 3-character code must use the IOU code alphabet and is placed at bytes 12 to 14.
// A 40-character hexadecimal code is taken verbatim. Anything else is an error.
func ParseCurrencyCode(code string) ([]byte, error) {
	switch len(code) {
	case 3:
		if code == "XRP" {
			return make([]byte, CurrencyCodeByteLength), nil
		}
		if !iouCodeRegex.MatchString(code) {
			return nil, errInvalidCurrencyCode
		}
		currencyBytes := make([]byte, CurrencyCodeByteLength)
		copy(currencyBytes[12:], code)
		return currencyBytes, nil
	case 2 * CurrencyCodeByteLength:
		return hex.DecodeString(code)
	}
	return nil, &InvalidCodeError{Disallowed: code}
}
