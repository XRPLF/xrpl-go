package currency

import (
	"math/big"
	"strconv"
	"strings"
)

const (
	maxNativeAmountDigits = 18
	// maxDecimalRatInputLen bounds plain decimal input to the largest drop amount plus a decimal point and XRP fraction
	// 18 + 1 (decimal point) + 6 (fraction length).
	maxDecimalRatInputLen = maxNativeAmountDigits + 1 + MaxFractionLength
	// maxDecimalRatExponent bounds scientific notation before parsing to keep conversion work proportional to native amounts
	// 1e17.
	maxDecimalRatExponent = maxNativeAmountDigits - 1
)

func decimalRat(value string) (*big.Rat, bool) {
	if len(value) > maxDecimalRatInputLen || containsInvalidChar(value) {
		return nil, false
	}

	if i := strings.IndexAny(value, "eE"); i >= 0 {
		exp, err := strconv.Atoi(value[i+1:])
		if err != nil || exp < -maxDecimalRatExponent || exp > maxDecimalRatExponent {
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
