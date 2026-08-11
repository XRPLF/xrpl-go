// Package currency provides utilities for working with XRP native currency conversions and calculations.
package currency

import (
	"math/big"
	"strings"
)

const (
	// DropsPerXRP is the number of drops equivalent to one XRP.
	// Use XrpToDrops and DropsToXrp for conversions. This constant is for reference only.
	DropsPerXRP = 1_000_000
	// MaxNativeDrops is the maximum native XRP amount in drops.
	MaxNativeDrops uint64 = 100_000_000_000_000_000
	// MaxFractionLength is the maximum allowed decimal places in an XRP value.
	MaxFractionLength int = 6
	// NativeCurrencySymbol is the symbol representing the native XRP currency.
	NativeCurrencySymbol string = "XRP"
)

var (
	maxDropsInt       = new(big.Int).SetUint64(MaxNativeDrops)
	dropsPerXRPBigInt = big.NewInt(DropsPerXRP)
	dropsPerXRPRat    = big.NewRat(DropsPerXRP, 1)
	bigIntOne         = big.NewInt(1)
)

// XrpToDrops converts an amount in XRP to an amount in drops.
func XrpToDrops(value string) (string, error) {
	xrp, ok := nativeAmountRat(value)
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
	drops, ok := nativeAmountRat(value)
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
