package currency

import (
	"fmt"
	"math/big"
	"strings"
)

// Drops is an exact, non-negative native XRP amount denominated in drops.
// It can contain a fractional drop during intermediate calculations. Its zero
// value is zero drops. All arithmetic methods return a new value.
type Drops struct {
	_   [0]func() // Prevent ==, which would compare rational pointers instead of values.
	rat *big.Rat
}

// DropsFromString creates Drops from a decimal drops string. The value must be
// non-negative and equal to a whole number of drops. Input is limited to 1024
// bytes and a decimal exponent magnitude of 17 to keep exact parsing bounded.
func DropsFromString(value string) (Drops, error) {
	drops, err := parseNativeAmount(value)
	if err != nil {
		return Drops{}, err
	}
	if !drops.IsInt() {
		return Drops{}, ErrFractionalDrops
	}
	return dropsFromRat(drops), nil
}

// DropsFromUint64 creates Drops from a whole-number drops value.
func DropsFromUint64(value uint64) Drops {
	return dropsFromRat(new(big.Rat).SetUint64(value))
}

// DropsFromXRP creates Drops from a decimal XRP string. The value must be
// non-negative and equal to a whole number of drops. Input is limited to 1024
// bytes and a decimal exponent magnitude of 17 to keep exact parsing bounded.
func DropsFromXRP(value string) (Drops, error) {
	xrp, err := parseNativeAmount(value)
	if err != nil {
		return Drops{}, err
	}

	drops := new(big.Rat).Mul(xrp, dropsPerXRPRat)
	if !drops.IsInt() {
		return Drops{}, ErrFractionalDrops
	}
	return dropsFromRat(drops), nil
}

// Add returns the exact sum of d and other.
func (d Drops) Add(other Drops) Drops {
	return dropsFromRat(new(big.Rat).Add(d.ratValue(), other.ratValue()))
}

// Mul returns d multiplied by an integer.
func (d Drops) Mul(multiplier uint64) Drops {
	factor := new(big.Rat).SetUint64(multiplier)
	return dropsFromRat(new(big.Rat).Mul(d.ratValue(), factor))
}

// MulDecimal returns d multiplied by an exact, non-negative decimal value.
// Input is limited to 1024 bytes and a decimal exponent magnitude of 324 to
// keep exact parsing bounded while accepting every finite float64 fee input.
func (d Drops) MulDecimal(multiplier string) (Drops, error) {
	factor, ok := multiplierRat(multiplier)
	if !ok || factor.Sign() < 0 {
		return Drops{}, ErrInvalidDecimalMultiplier
	}
	return dropsFromRat(new(big.Rat).Mul(d.ratValue(), factor)), nil
}

// MulRat returns d multiplied by numerator divided by denominator.
func (d Drops) MulRat(numerator, denominator uint64) (Drops, error) {
	if denominator == 0 {
		return Drops{}, ErrDropsDivisionByZero
	}

	factor := new(big.Rat).SetFrac(
		new(big.Int).SetUint64(numerator),
		new(big.Int).SetUint64(denominator),
	)
	return dropsFromRat(new(big.Rat).Mul(d.ratValue(), factor)), nil
}

// Min returns the smaller of d and other.
func (d Drops) Min(other Drops) Drops {
	if d.Cmp(other) <= 0 {
		return dropsFromRat(new(big.Rat).Set(d.ratValue()))
	}
	return dropsFromRat(new(big.Rat).Set(other.ratValue()))
}

// Cmp compares d and other and returns -1, 0, or 1.
func (d Drops) Cmp(other Drops) int {
	return d.ratValue().Cmp(other.ratValue())
}

// Ceil returns d rounded upward to a whole number of drops.
func (d Drops) Ceil() Drops {
	quotient, remainder := new(big.Int), new(big.Int)
	value := d.ratValue()
	quotient.QuoRem(value.Num(), value.Denom(), remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return dropsFromRat(new(big.Rat).SetInt(quotient))
}

// RoundHalfUp returns d rounded to the nearest whole number of drops. A half
// drop is rounded upward.
func (d Drops) RoundHalfUp() Drops {
	quotient, remainder := new(big.Int), new(big.Int)
	value := d.ratValue()
	quotient.QuoRem(value.Num(), value.Denom(), remainder)
	doubledRemainder := new(big.Int).Lsh(new(big.Int).Set(remainder), 1)
	if doubledRemainder.Cmp(value.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return dropsFromRat(new(big.Rat).SetInt(quotient))
}

// IsWhole reports whether d contains a whole number of drops.
func (d Drops) IsWhole() bool {
	return d.rat == nil || d.rat.IsInt()
}

// WholeString returns d as a whole-number drops string.
func (d Drops) WholeString() (string, error) {
	if !d.IsWhole() {
		return "", ErrFractionalDrops
	}
	if d.rat == nil {
		return "0", nil
	}
	return d.rat.Num().String(), nil
}

// XRPString returns d as an XRP decimal string. The value must contain a whole
// number of drops.
func (d Drops) XRPString() (string, error) {
	if !d.IsWhole() {
		return "", ErrFractionalDrops
	}
	if d.rat == nil {
		return "0", nil
	}

	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(d.rat.Num(), dropsPerXRPBigInt, remainder)
	if remainder.Sign() == 0 {
		return quotient.String(), nil
	}

	fraction := fmt.Sprintf("%06d", remainder)
	return quotient.String() + "." + strings.TrimRight(fraction, "0"), nil
}

func parseNativeAmount(value string) (*big.Rat, error) {
	amount, ok := nativeAmountRat(value)
	if !ok {
		return nil, ErrInvalidNativeAmount
	}
	if amount.Sign() < 0 {
		return nil, ErrNegativeNativeAmount
	}
	return amount, nil
}

func dropsFromRat(value *big.Rat) Drops {
	if value.Sign() == 0 {
		return Drops{}
	}
	return Drops{rat: value}
}

func (d Drops) ratValue() *big.Rat {
	if d.rat == nil {
		return new(big.Rat)
	}
	return d.rat
}
