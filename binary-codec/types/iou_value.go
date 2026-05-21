package types

import (
	"strings"

	bigdecimal "github.com/Peersyst/xrpl-go/pkg/big-decimal"
)

// XRPL token (IOU) amounts are decimal strings encoded as a sign plus
// mantissa * 10^exponent. Non-zero values must fit in 16 significant digits
// once normalized, with a canonical exponent between -96 and 80. Zero is
// encoded separately as ZeroCurrencyAmountHex (0x8000000000000000).
// Reference: https://xrpl.org/docs/references/protocol/binary-format#token-amount-format

// xrplStringNumberAllowedChars is the full set of characters that can legally
// appear anywhere in an XRPL String Number. '+' is only meaningful as an
// exponent sign, but is included here so the character-set check only fires
// for inputs that contain truly out-of-set characters.
const xrplStringNumberAllowedChars = "0123456789.-+eE"

// VerifyIOUValue validates that value is an XRPL String Number within the IOU
// precision and exponent bounds. A numeric zero (for example "0", "0.0",
// "-0", or "0e5") is valid per the XRPL protocol. The first return reports
// whether the validated value is numerically zero, so callers do not need to
// repeat the grammar check to distinguish signed zero from negative non-zero.
func VerifyIOUValue(value string) (bool, error) {
	_, isZero, err := parseIOUValue(value)
	return isZero, err
}

// parseIOUValue validates value once and returns its decoded form. For a
// numeric zero it returns (nil, true, nil); the caller must check isZero
// before dereferencing bigDecimal. It is the single parse shared by
// VerifyIOUValue and SerializeIssuedCurrencyValue.
func parseIOUValue(value string) (bigDecimal *bigdecimal.BigDecimal, isZero bool, err error) {
	if !isXRPLStringNumber(value) {
		if !hasOnlyXRPLStringNumberChars(value) {
			return nil, false, bigdecimal.ErrInvalidCharacter{Allowed: bigdecimal.AllowedCharacters}
		}
		return nil, false, ErrInvalidStringNumber
	}
	if isZeroMantissa(value) {
		return nil, true, nil
	}

	bigDecimal, err = bigdecimal.NewBigDecimal(value)
	if err != nil {
		return nil, false, err
	}

	if bigDecimal.Precision > MaxIOUPrecision {
		return nil, false, &OutOfRangeError{Type: "Precision"}
	}

	// XRPL stores a token amount as a mantissa normalized to exactly
	// MaxIOUPrecision (16) significant digits times 10^exponent. BigDecimal
	// holds the value as UnscaledValue (Precision digits) * 10^Scale, so
	// renormalizing to a 16-digit mantissa shifts the exponent to
	// Scale + Precision - MaxIOUPrecision. That canonical exponent must fall
	// within the protocol's [MinIOUExponent, MaxIOUExponent] range.
	canonicalExp := bigDecimal.Scale + bigDecimal.Precision - MaxIOUPrecision

	if canonicalExp < MinIOUExponent || canonicalExp > MaxIOUExponent {
		return nil, false, &OutOfRangeError{Type: "Exponent"}
	}

	return bigDecimal, false, nil
}

// isZeroMantissa reports whether value's mantissa is all zeros. It assumes
// value has already passed isXRPLStringNumber, so the mantissa is non-empty.
func isZeroMantissa(value string) bool {
	mantissa, _, _ := splitXRPLStringNumber(value)
	for _, r := range mantissa {
		if r != '.' && r != '0' {
			return false
		}
	}

	return true
}

// isXRPLStringNumber reports whether value matches the XRPL String Number
// grammar: an optional leading '-', a non-zero-prefaced decimal mantissa
// (e.g. "0.1" ok, "00.1" not), and an optional "e"/"E" integer exponent.
func isXRPLStringNumber(value string) bool {
	if value == "" {
		return false
	}

	mantissa, exponent, hasExponent := splitXRPLStringNumber(value)
	if !isValidXRPLMantissa(mantissa) {
		return false
	}

	return !hasExponent || isValidXRPLExponent(exponent)
}

func splitXRPLStringNumber(value string) (mantissa, exponent string, hasExponent bool) {
	value = strings.TrimPrefix(value, "-")
	if value == "" {
		return "", "", false
	}

	expIndex := strings.IndexAny(value, "eE")
	if expIndex == -1 {
		return value, "", false
	}

	return value[:expIndex], value[expIndex+1:], true
}

func isValidXRPLMantissa(mantissa string) bool {
	whole, fraction, hasDecimal := strings.Cut(mantissa, ".")

	if whole == "" || !isAllDecimalDigits(whole) {
		return false
	}

	// XRPL String Numbers are non-zero-prefaced: "0.1" is valid, "00.1" is not.
	if len(whole) > 1 && whole[0] == '0' {
		return false
	}

	if hasDecimal {
		return fraction != "" && isAllDecimalDigits(fraction)
	}

	return true
}

func isValidXRPLExponent(exponent string) bool {
	if len(exponent) > 0 && (exponent[0] == '+' || exponent[0] == '-') {
		exponent = exponent[1:]
	}

	return exponent != "" && isAllDecimalDigits(exponent)
}

func isAllDecimalDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func hasOnlyXRPLStringNumberChars(value string) bool {
	for _, r := range value {
		if !strings.ContainsRune(xrplStringNumberAllowedChars, r) {
			return false
		}
	}

	return true
}
