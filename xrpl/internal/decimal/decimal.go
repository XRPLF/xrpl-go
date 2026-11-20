// Package decimal parses untrusted decimal text into exact rational values.
//
// Use it when the value is needed for arithmetic, for example a balance
// difference. pkg/big-decimal has a different purpose: it splits a decimal
// string into sign, digits, and scale for the binary codec, it does no
// arithmetic, and it rejects zero.
package decimal

import (
	"math/big"
	"strconv"
	"strings"
)

// MaxInputLen bounds the work and allocations done by big.Rat.SetString.
// XRPL amounts and fee values are much shorter, while 1024 bytes leaves ample
// room for non-canonical text. This is a parser safety limit, not an XRPL
// protocol limit.
const MaxInputLen = 1024

// maxAmountExponent bounds the exponent of a token amount. xrpld emits
// issued-currency exponents from -96 to 80, so no valid amount is rejected.
const maxAmountExponent = 256

// ParseAmount converts the text of a token amount to an exact rational. It
// applies the rules of Parse with the exponent limit for token amounts.
func ParseAmount(value string) (*big.Rat, bool) {
	return Parse(value, maxAmountExponent)
}

// Parse converts value to an exact rational. It accepts an optional sign,
// decimal digits with an optional point, and an optional exponent. It rejects
// values longer than MaxInputLen and exponents beyond ±maxExponent, so the
// cost of a conversion stays bounded.
func Parse(value string, maxExponent int) (*big.Rat, bool) {
	if len(value) > MaxInputLen || containsInvalidChar(value) {
		return nil, false
	}

	if i := strings.IndexAny(value, "eE"); i >= 0 {
		exp, err := strconv.Atoi(value[i+1:])
		if err != nil || exp < -maxExponent || exp > maxExponent {
			return nil, false
		}
	}

	return new(big.Rat).SetString(value)
}

// Exponent bounds of plain notation in Format. They match the defaults of
// bignumber.js toString, which xrpl.js uses for token values.
const (
	minPlainExponent = -6
	maxPlainExponent = 20
)

// Format returns value as decimal text. It uses plain notation when the
// decimal exponent of value is from -6 to 20 and exponential notation
// otherwise, for example "0.000001", "1e-7", and "1e+21". It reports false
// when value has no finite decimal expansion, for example 1/3.
func Format(value *big.Rat) (string, bool) {
	precision, exact := value.FloatPrec()
	if !exact {
		return "", false
	}
	if value.Sign() == 0 {
		return "0", true
	}

	// scaled is |value| * 10^precision, which is an integer because the
	// expansion is finite.
	scaled := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)), nil)
	scaled.Mul(scaled, value.Num())
	scaled.Quo(scaled, value.Denom())
	digits := scaled.Abs(scaled).String()

	exponent := len(digits) - 1 - precision
	if exponent >= minPlainExponent && exponent <= maxPlainExponent {
		return value.FloatString(precision), true
	}

	var text strings.Builder
	if value.Sign() < 0 {
		text.WriteByte('-')
	}
	digits = strings.TrimRight(digits, "0")
	text.WriteString(digits[:1])
	if len(digits) > 1 {
		text.WriteByte('.')
		text.WriteString(digits[1:])
	}
	text.WriteByte('e')
	if exponent > 0 {
		text.WriteByte('+')
	}
	text.WriteString(strconv.Itoa(exponent))
	return text.String(), true
}

func containsInvalidChar(value string) bool {
	if value == "" {
		return true
	}

	for i := 0; i < len(value); i++ {
		c := value[i]
		switch {
		case c >= '0' && c <= '9', c == '.', c == 'e', c == 'E':
			// always valid
		case c == '+' || c == '-':
			if i != 0 && value[i-1] != 'e' && value[i-1] != 'E' {
				return true
			}
		default:
			return true
		}
	}

	return false
}
