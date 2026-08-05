package client

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
)

var decimalFeePattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)

// Fee is an exact, non-negative fee value denominated in drops. It can retain a
// fractional drop until the final XRPL-required ceiling operation.
type Fee struct {
	drops *big.Rat
}

// NetworkFeeXRP calculates the load-adjusted and capped network fee. Inputs use
// binary64 precision, the cap is applied before rounding, and the result is
// rounded half-up to six XRP decimal places.
func NetworkFeeXRP(baseFeeXRP, loadFactor, cushion float64, maxFeeXRP string) (string, error) {
	baseFee, err := ratFromFloat64(baseFeeXRP)
	if err != nil {
		return "", fmt.Errorf("%w: base fee", ErrInvalidFeeValue)
	}
	feeCushion, err := ratFromFloat64(cushion)
	if err != nil {
		return "", fmt.Errorf("%w: fee cushion", ErrInvalidFeeValue)
	}
	maxFee, err := ratFromDecimal(maxFeeXRP)
	if err != nil {
		return "", fmt.Errorf("%w: maximum fee", ErrInvalidFeeValue)
	}

	load, err := ratFromFloat64(loadFactor)
	if err != nil {
		return "", fmt.Errorf("%w: load factor", ErrInvalidFeeValue)
	}

	fee := new(big.Rat).Mul(baseFee, load)
	fee.Mul(fee, feeCushion)
	if fee.Cmp(maxFee) > 0 {
		fee.Set(maxFee)
	}

	scaledDrops := new(big.Rat).Mul(fee, new(big.Rat).SetInt64(currency.DropsPerXRP))
	drops := roundHalfUp(scaledDrops)
	return dropsToXRP(drops), nil
}

// NewFeeFromDrops creates an exact fee from a whole-number drops string.
func NewFeeFromDrops(drops string) (*Fee, error) {
	value, ok := new(big.Int).SetString(drops, 10)
	if !ok || value.Sign() < 0 {
		return nil, fmt.Errorf("%w: drops", ErrInvalidFeeValue)
	}
	return &Fee{drops: new(big.Rat).SetInt(value)}, nil
}

// NewFeeFromUint64 creates an exact fee from a whole-number drops value.
func NewFeeFromUint64(drops uint64) *Fee {
	return &Fee{drops: new(big.Rat).SetUint64(drops)}
}

// NewFeeFromXRP creates an exact fee from an XRP decimal string. The XRP value
// must contain no fractional drops.
func NewFeeFromXRP(xrp string) (*Fee, error) {
	value, err := ratFromDecimal(xrp)
	if err != nil {
		return nil, fmt.Errorf("%w: XRP", ErrInvalidFeeValue)
	}
	value.Mul(value, new(big.Rat).SetInt64(currency.DropsPerXRP))
	if !value.IsInt() {
		return nil, ErrFeeHasTooManyDecimals
	}
	return &Fee{drops: value}, nil
}

// Add returns the exact sum of two fees.
func (f *Fee) Add(other *Fee) *Fee {
	return &Fee{drops: new(big.Rat).Add(f.drops, other.drops)}
}

// Multiply returns the fee multiplied by an integer.
func (f *Fee) Multiply(multiplier uint64) *Fee {
	return &Fee{drops: new(big.Rat).Mul(f.drops, new(big.Rat).SetUint64(multiplier))}
}

// MultiplyFraction returns the fee multiplied by numerator/denominator.
func (f *Fee) MultiplyFraction(numerator, denominator uint64) (*Fee, error) {
	if denominator == 0 {
		return nil, fmt.Errorf("%w: zero denominator", ErrInvalidFeeValue)
	}
	factor := new(big.Rat).SetFrac(
		new(big.Int).SetUint64(numerator),
		new(big.Int).SetUint64(denominator),
	)
	return &Fee{drops: new(big.Rat).Mul(f.drops, factor)}, nil
}

// Min returns the smaller of two fees.
func (f *Fee) Min(other *Fee) *Fee {
	if f.drops.Cmp(other.drops) <= 0 {
		return &Fee{drops: new(big.Rat).Set(f.drops)}
	}
	return &Fee{drops: new(big.Rat).Set(other.drops)}
}

// CeilDrops returns the fee rounded upward to a whole number of drops.
func (f *Fee) CeilDrops() string {
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(f.drops.Num(), f.drops.Denom(), remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.String()
}

func ratFromFloat64(value float64) (*big.Rat, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return nil, ErrInvalidFeeValue
	}
	return ratFromDecimal(strconv.FormatFloat(value, 'g', -1, 64))
}

func ratFromDecimal(value string) (*big.Rat, error) {
	if !decimalFeePattern.MatchString(value) {
		return nil, ErrInvalidFeeValue
	}
	result, ok := new(big.Rat).SetString(value)
	if !ok || result.Sign() < 0 {
		return nil, ErrInvalidFeeValue
	}
	return result, nil
}

func roundHalfUp(value *big.Rat) *big.Int {
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(value.Num(), value.Denom(), remainder)
	doubledRemainder := new(big.Int).Lsh(new(big.Int).Set(remainder), 1)
	if doubledRemainder.Cmp(value.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

func dropsToXRP(drops *big.Int) string {
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(drops, big.NewInt(currency.DropsPerXRP), remainder)
	if remainder.Sign() == 0 {
		return quotient.String()
	}
	fraction := fmt.Sprintf("%06s", remainder.String())
	return quotient.String() + "." + strings.TrimRight(fraction, "0")
}
