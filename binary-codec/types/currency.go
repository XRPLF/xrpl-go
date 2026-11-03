//revive:disable:var-naming
package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"regexp"

	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
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

	return formatCurrencyCode(currencyBytes), nil
}

// formatCurrencyCode converts currency bytes to their JSON form, following rippled's
// to_string(Currency). It is the inverse of ParseCurrencyCode. 20 zero bytes are "XRP".
// The standard layout of an IOU code other than "XRP" is that 3-character code.
// Anything else, including the reserved standard layout of "XRP", is uppercase hexadecimal.
// This intentionally differs from rippled for noCurrency: rippled returns "1", but this
// returns the 40 hex characters, because ParseCurrencyCode("1") fails and the hex form
// keeps the bytes round-trippable.
func formatCurrencyCode(currencyBytes []byte) string {
	if bytes.Equal(currencyBytes, XRPBytes) {
		return "XRP"
	}
	if len(currencyBytes) == CurrencyCodeByteLength &&
		bytes.Equal(currencyBytes[:12], XRPBytes[:12]) &&
		bytes.Equal(currencyBytes[15:], XRPBytes[15:]) &&
		!bytes.Equal(currencyBytes, isoXRPBytes) &&
		iouCodeRegex.Match(currencyBytes[12:15]) {
		return string(currencyBytes[12:15])
	}
	return hexutil.EncodeToUpperHex(currencyBytes)
}

// ParseCurrencyCode converts a JSON currency code to its 20-byte form, following
// rippled's to_currency. "XRP" is the native currency and encodes as 20 zero bytes.
// A 3-character code must use the IOU code alphabet and is placed at bytes 12 to 14;
// a code outside that alphabet returns an error that matches ErrInvalidCurrency.
// A 40-character hexadecimal code is taken verbatim. Any other length returns an *InvalidCodeError.
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
