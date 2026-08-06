package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetworkFeeXRP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		baseFeeXRP float64
		loadFactor float64
		cushion    float64
		maxFeeXRP  string
		expected   string
	}{
		{name: "default fee", baseFeeXRP: 0.00001, loadFactor: 1, cushion: 1.2, maxFeeXRP: "2", expected: "0.000012"},
		{name: "explicit zero base fee", baseFeeXRP: 0, loadFactor: 1, cushion: 1.2, maxFeeXRP: "2", expected: "0"},
		{name: "fractional load factor", baseFeeXRP: 0.00001, loadFactor: 1.5, cushion: 1.2, maxFeeXRP: "2", expected: "0.000018"},
		{name: "explicit zero load factor", baseFeeXRP: 0.00001, loadFactor: 0, cushion: 1.2, maxFeeXRP: "2", expected: "0"},
		{name: "half drop rounds upward", baseFeeXRP: 0.000001, loadFactor: 10, cushion: 1.05, maxFeeXRP: "2", expected: "0.000011"},
		{name: "maximum fee is applied before rounding", baseFeeXRP: 1, loadFactor: 1000, cushion: 1.2, maxFeeXRP: "2", expected: "2"},
		{name: "decimal maximum fee", baseFeeXRP: 1, loadFactor: 1000, cushion: 1.2, maxFeeXRP: "0.123456", expected: "0.123456"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, err := NetworkFeeXRP(test.baseFeeXRP, test.loadFactor, test.cushion, test.maxFeeXRP)
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestNetworkFeeXRPRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	_, err := NetworkFeeXRP(-1, 1, 1.2, "2")
	require.ErrorIs(t, err, ErrInvalidFeeValue)

	_, err = NetworkFeeXRP(0.00001, -1, 1.2, "2")
	require.ErrorIs(t, err, ErrInvalidFeeValue)

	for _, maxFee := range []string{"invalid", "1/2", "0x10", "-1"} {
		_, err = NetworkFeeXRP(0.00001, 1, 1.2, maxFee)
		require.ErrorIs(t, err, ErrInvalidFeeValue)
	}
}

func TestFeeArithmetic(t *testing.T) {
	t.Parallel()

	base, err := NewFeeFromDrops("10")
	require.NoError(t, err)

	emptyFulfillmentEscrow, err := base.MultiplyFraction(33*16, 16)
	require.NoError(t, err)
	require.Equal(t, "330", emptyFulfillmentEscrow.CeilDrops())

	escrow, err := base.MultiplyFraction(33*16+4, 16)
	require.NoError(t, err)
	require.Equal(t, "333", escrow.CeilDrops())

	multisigned := base.Add(base.Multiply(2))
	require.Equal(t, "30", multisigned.CeilDrops())

	maxFee, err := NewFeeFromXRP("0.000025")
	require.NoError(t, err)
	require.Equal(t, "25", multisigned.Min(maxFee).CeilDrops())

	_, err = NewFeeFromXRP("0.0000001")
	require.ErrorIs(t, err, ErrFeeHasTooManyDecimals)
}
