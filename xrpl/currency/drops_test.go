package currency

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDropsZeroValue(t *testing.T) {
	t.Parallel()

	var zero Drops

	whole, err := zero.WholeString()
	require.NoError(t, err)
	require.Equal(t, "0", whole)

	xrp, err := zero.XRPString()
	require.NoError(t, err)
	require.Equal(t, "0", xrp)

	require.True(t, zero.IsWhole())
	require.Equal(t, 0, zero.Cmp(DropsFromUint64(0)))
	require.Equal(t, 0, zero.Add(zero).Cmp(zero))
	require.Equal(t, 0, zero.Mul(10).Cmp(zero))
	require.Equal(t, 0, zero.Ceil().Cmp(zero))
	require.Equal(t, 0, zero.RoundHalfUp().Cmp(zero))
}

func TestDropsFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "zero", value: "0", expected: "0"},
		{name: "whole number", value: "12", expected: "12"},
		{name: "fractional zeros", value: "12.0", expected: "12"},
		{name: "scientific notation", value: "1e6", expected: "1000000"},
		{name: "above native maximum", value: "100000000000000001", expected: "100000000000000001"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			drops, err := DropsFromString(test.value)
			require.NoError(t, err)

			actual, err := drops.WholeString()
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestDropsFromStringRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       string
		expectedErr error
	}{
		{name: "empty", value: "", expectedErr: ErrInvalidNativeAmount},
		{name: "invalid", value: "invalid", expectedErr: ErrInvalidNativeAmount},
		{name: "fraction syntax", value: "1/2", expectedErr: ErrInvalidNativeAmount},
		{name: "negative", value: "-1", expectedErr: ErrNegativeNativeAmount},
		{name: "fractional drop", value: "1.1", expectedErr: ErrFractionalDrops},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := DropsFromString(test.value)
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestDropsFromXRP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "zero", value: "0", expected: "0"},
		{name: "one drop", value: "0.000001", expected: "1"},
		{name: "decimal XRP", value: "1.234567", expected: "1234567"},
		{name: "scientific notation", value: "1e-6", expected: "1"},
		{name: "above native maximum", value: "100000000001", expected: "100000000001000000"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			drops, err := DropsFromXRP(test.value)
			require.NoError(t, err)

			actual, err := drops.WholeString()
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestDropsFromXRPRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       string
		expectedErr error
	}{
		{name: "empty", value: "", expectedErr: ErrInvalidNativeAmount},
		{name: "invalid", value: "invalid", expectedErr: ErrInvalidNativeAmount},
		{name: "negative", value: "-0.000001", expectedErr: ErrNegativeNativeAmount},
		{name: "fractional drop", value: "0.0000001", expectedErr: ErrFractionalDrops},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := DropsFromXRP(test.value)
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestDropsFromUint64(t *testing.T) {
	t.Parallel()

	drops := DropsFromUint64(math.MaxUint64)
	actual, err := drops.WholeString()
	require.NoError(t, err)
	require.Equal(t, "18446744073709551615", actual)
}

func TestDropsArithmeticIsExactAndImmutable(t *testing.T) {
	t.Parallel()

	base := DropsFromUint64(10)
	sum := base.Add(DropsFromUint64(5))
	product := base.Mul(3)
	fraction, err := base.MulRat(532, 16)
	require.NoError(t, err)

	baseString, err := base.WholeString()
	require.NoError(t, err)
	require.Equal(t, "10", baseString)

	sumString, err := sum.WholeString()
	require.NoError(t, err)
	require.Equal(t, "15", sumString)

	productString, err := product.WholeString()
	require.NoError(t, err)
	require.Equal(t, "30", productString)

	_, err = fraction.WholeString()
	require.ErrorIs(t, err, ErrFractionalDrops)

	ceilString, err := fraction.Ceil().WholeString()
	require.NoError(t, err)
	require.Equal(t, "333", ceilString)
}

func TestDropsMulDecimal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		base       uint64
		multiplier string
		expected   string
	}{
		{name: "decimal", base: 10, multiplier: "1.2", expected: "12"},
		{
			name:       "high precision",
			base:       1,
			multiplier: "1234567890123456789012345678901234567890",
			expected:   "1234567890123456789012345678901234567890",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			product, err := DropsFromUint64(test.base).MulDecimal(test.multiplier)
			require.NoError(t, err)

			actual, err := product.WholeString()
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestDropsMulDecimalFractionalResult(t *testing.T) {
	t.Parallel()

	fraction, err := DropsFromUint64(10).MulDecimal("1.05")
	require.NoError(t, err)
	require.False(t, fraction.IsWhole())

	rounded, err := fraction.RoundHalfUp().WholeString()
	require.NoError(t, err)
	require.Equal(t, "11", rounded)
}

func TestDropsMulDecimalExponentBoundaries(t *testing.T) {
	t.Parallel()

	base := DropsFromUint64(10)
	huge, err := base.MulDecimal("1e308")
	require.NoError(t, err)
	require.Equal(t, 1, huge.Cmp(base))

	tiny, err := base.MulDecimal("5e-324")
	require.NoError(t, err)
	require.Equal(t, 1, tiny.Cmp(Drops{}))
}

func TestDropsMulDecimalRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       string
		expectedErr error
	}{
		{name: "empty", value: "", expectedErr: ErrInvalidDecimalMultiplier},
		{name: "invalid", value: "invalid", expectedErr: ErrInvalidDecimalMultiplier},
		{name: "exponent too large", value: "1e325", expectedErr: ErrInvalidDecimalMultiplier},
		{name: "input too long", value: strings.Repeat("1", maxDecimalRatInputLen+1), expectedErr: ErrInvalidDecimalMultiplier},
		{name: "negative", value: "-1", expectedErr: ErrInvalidDecimalMultiplier},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := DropsFromUint64(1).MulDecimal(test.value)
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestDropsMulRatRejectsZeroDenominator(t *testing.T) {
	t.Parallel()

	_, err := DropsFromUint64(1).MulRat(1, 0)
	require.ErrorIs(t, err, ErrDropsDivisionByZero)
}

func TestDropsRoundHalfUp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		numerator   uint64
		denominator uint64
		expected    string
	}{
		{name: "below half", numerator: 124, denominator: 10, expected: "12"},
		{name: "exactly half", numerator: 125, denominator: 10, expected: "13"},
		{name: "above half", numerator: 126, denominator: 10, expected: "13"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			value, err := DropsFromUint64(1).MulRat(test.numerator, test.denominator)
			require.NoError(t, err)

			actual, err := value.RoundHalfUp().WholeString()
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestDropsMinAndCmp(t *testing.T) {
	t.Parallel()

	ten := DropsFromUint64(10)
	twenty := DropsFromUint64(20)

	require.Equal(t, -1, ten.Cmp(twenty))
	require.Equal(t, 0, ten.Cmp(ten))
	require.Equal(t, 1, twenty.Cmp(ten))
	require.Equal(t, 0, ten.Min(twenty).Cmp(ten))
	require.Equal(t, 0, twenty.Min(ten).Cmp(ten))
}

func TestDropsXRPString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		drops    uint64
		expected string
	}{
		{name: "zero", drops: 0, expected: "0"},
		{name: "one drop", drops: 1, expected: "0.000001"},
		{name: "fractional XRP", drops: 1_234_567, expected: "1.234567"},
		{name: "trailing zeros", drops: 1_200_000, expected: "1.2"},
		{name: "whole XRP", drops: 2_000_000, expected: "2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := DropsFromUint64(test.drops).XRPString()
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestDropsXRPStringRejectsFractionalDrops(t *testing.T) {
	t.Parallel()

	fraction, err := DropsFromUint64(1).MulRat(1, 2)
	require.NoError(t, err)

	_, err = fraction.XRPString()
	require.ErrorIs(t, err, ErrFractionalDrops)
}
