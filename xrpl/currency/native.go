// Package currency provides utilities for working with XRP native currency conversions and calculations.
package currency

import (
	"math/big"
	"strconv"
	"strings"
)

const (
	// DropsPerXrp is the number of drops equivalent to one XRP.
	//
	// Deprecated: use XrpToDrops and DropsToXrp for native amount conversions.
	// The conversion helpers use exact rational arithmetic internally instead of float64.
	DropsPerXrp float64 = 1000000
	// MaxFractionLength is the maximum allowed decimal places in an XRP value.
	MaxFractionLength int = 6
	// NativeCurrencySymbol is the symbol representing the native XRP currency.
	NativeCurrencySymbol string = "XRP"
)

const (
	dropsPerXRP = int64(1000000)
	maxDrops    = uint64(100000000000000000)

	maxNativeAmountDigits = 18
	// maxDecimalRatInputLen bounds plain decimal input to the largest drop amount plus a decimal point and XRP fraction
	// 18 + 1 (decimal point) + 6 (fraction length).
	maxDecimalRatInputLen = maxNativeAmountDigits + 1 + MaxFractionLength
	// maxDecimalRatExponent bounds scientific notation before parsing to keep conversion work proportional to native amounts
	// 1e17.
	maxDecimalRatExponent = maxNativeAmountDigits - 1
)

var (
	maxDropsInt       = new(big.Int).SetUint64(maxDrops)
	dropsPerXRPBigInt = big.NewInt(dropsPerXRP)
	dropsPerXRPRat    = big.NewRat(dropsPerXRP, 1)
	bigIntOne         = big.NewInt(1)
)

// XrpToDrops converts an amount in XRP to an amount in drops.
func XrpToDrops(value string) (string, error) {
	xrp, ok := decimalRat(value)
	if !ok {
		return "", ErrXrpToDropsInvalidValue
	}

	if xrp.Sign() < 0 {
		return "", ErrXrpToDropsNegativeValue
	}

	drops := new(big.Rat).Mul(xrp, dropsPerXRPRat)
	if drops.Denom().Cmp(bigIntOne) != 0 {
		return "", ErrXrpToDropsTooManyDecimals
	}

	if drops.Num().Cmp(maxDropsInt) > 0 {
		return "", ErrXrpToDropsExceedsMax
	}

	return drops.Num().String(), nil
}

// DropsToXrp converts an amount of drops into an amount of XRP.
func DropsToXrp(value string) (string, error) {
	drops, ok := decimalRat(value)
	if !ok {
		return "", ErrDropsToXrpInvalidValue
	}

	if drops.Sign() < 0 {
		return "", ErrDropsToXrpNegativeValue
	}

	if drops.Denom().Cmp(bigIntOne) != 0 {
		return "", ErrDropsToXrpFractionalDrops
	}

	dropInt := drops.Num()
	if dropInt.Cmp(maxDropsInt) > 0 {
		return "", ErrDropsToXrpExceedsMax
	}

	whole := new(big.Int).Div(dropInt, dropsPerXRPBigInt)
	fraction := new(big.Int).Mod(dropInt, dropsPerXRPBigInt)
	if fraction.Sign() == 0 {
		return whole.String(), nil
	}

	fractionString := fraction.String()
	if len(fractionString) < MaxFractionLength {
		fractionString = strings.Repeat("0", MaxFractionLength-len(fractionString)) + fractionString
	}

	return whole.String() + "." + strings.TrimRight(fractionString, "0"), nil
}

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
