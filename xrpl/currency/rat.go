package currency

import (
	"math/big"

	"github.com/Peersyst/xrpl-go/xrpl/internal/decimal"
)

const (
	maxNativeAmountDigits = 18
	// maxDecimalRatExponent bounds scientific notation before parsing to keep conversion work proportional to native amounts
	// 1e17.
	maxDecimalRatExponent = maxNativeAmountDigits - 1
	// A finite float64 formatted as decimal text ranges from about 5e-324 to 1.8e308.
	// Use 324 as a symmetric limit and to prevent unbounded big.Rat allocation.
	maxMultiplierRatExponent = 324
)

func nativeAmountRat(value string) (*big.Rat, bool) {
	return decimal.Parse(value, maxDecimalRatExponent)
}

func multiplierRat(value string) (*big.Rat, bool) {
	return decimal.Parse(value, maxMultiplierRatExponent)
}
