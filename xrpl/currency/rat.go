package currency

import (
	"math/big"
	"strconv"
	"strings"
)

const (
	maxNativeAmountDigits = 18
	// maxDecimalRatInputLen bounds the work and allocations done by big.Rat.SetString.
	// Real fee inputs formatted from finite float64 values are shorter than 32 bytes,
	// while 1024 bytes leaves ample room for non-canonical fee text and exact user
	// multipliers. This is a parser safety limit, not an XRPL protocol limit.
	maxDecimalRatInputLen = 1024
	// maxDecimalRatExponent bounds scientific notation before parsing to keep conversion work proportional to native amounts
	// 1e17.
	maxDecimalRatExponent = maxNativeAmountDigits - 1
	// A finite float64 formatted as decimal text ranges from about 5e-324 to 1.8e308.
	// Use 324 as a symmetric limit and to prevent unbounded big.Rat allocation.
	maxMultiplierRatExponent = 324
)

func nativeAmountRat(value string) (*big.Rat, bool) {
	return limitedDecimalRat(value, maxDecimalRatExponent)
}

func multiplierRat(value string) (*big.Rat, bool) {
	return limitedDecimalRat(value, maxMultiplierRatExponent)
}

func limitedDecimalRat(value string, maxExponent int) (*big.Rat, bool) {
	if len(value) > maxDecimalRatInputLen || containsInvalidChar(value) {
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
