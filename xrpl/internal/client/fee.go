package client

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
)

// NetworkFeeDrops calculates the exact load-adjusted and capped network fee for
// one base fee. Inputs use binary64 precision and the result keeps any
// fractional drop, because rippled scales a transaction's whole base fee for
// load in one step. Callers multiply by the transaction's base-fee factor and
// round once, so a fractional drop is never amplified by that factor.
func NetworkFeeDrops(baseFeeXRP, loadFactor, cushion float64, maxFee currency.Drops) (currency.Drops, error) {
	baseFeeText, err := decimalFromFloat64(baseFeeXRP)
	if err != nil {
		return currency.Drops{}, fmt.Errorf("%w: base fee", ErrInvalidFeeValue)
	}
	// Start with one XRP in drops so the base fee can keep fractional drops
	// until the final rounding step.
	oneXRP := currency.DropsFromUint64(currency.DropsPerXRP)
	baseFee, err := oneXRP.MulDecimal(baseFeeText)
	if err != nil {
		return currency.Drops{}, fmt.Errorf("%w: base fee", ErrInvalidFeeValue)
	}

	cushionText, err := decimalFromFloat64(cushion)
	if err != nil {
		return currency.Drops{}, fmt.Errorf("%w: fee cushion", ErrInvalidFeeValue)
	}

	loadFactorText, err := decimalFromFloat64(loadFactor)
	if err != nil {
		return currency.Drops{}, fmt.Errorf("%w: load factor", ErrInvalidFeeValue)
	}

	fee, err := baseFee.MulDecimal(loadFactorText)
	if err != nil {
		return currency.Drops{}, fmt.Errorf("%w: load factor", ErrInvalidFeeValue)
	}
	fee, err = fee.MulDecimal(cushionText)
	if err != nil {
		return currency.Drops{}, fmt.Errorf("%w: fee cushion", ErrInvalidFeeValue)
	}

	return fee.Min(maxFee), nil
}

// ParseFeeXRP creates an exact drops value from an XRP fee string while
// preserving the client fee error vocabulary.
func ParseFeeXRP(value string) (currency.Drops, error) {
	drops, err := currency.DropsFromXRP(value)
	if err == nil {
		return drops, nil
	}
	if errors.Is(err, currency.ErrFractionalDrops) {
		return currency.Drops{}, ErrFeeHasTooManyDecimals
	}
	return currency.Drops{}, fmt.Errorf("%w: XRP", ErrInvalidFeeValue)
}

func decimalFromFloat64(value float64) (string, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return "", ErrInvalidFeeValue
	}
	return strconv.FormatFloat(value, 'g', -1, 64), nil
}
