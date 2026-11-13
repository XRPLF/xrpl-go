// Package decimal parses untrusted decimal text into exact rational values.
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
