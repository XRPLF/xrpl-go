package client

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetworkFeeDrops(t *testing.T) {
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
		{name: "maximum float load factor", baseFeeXRP: 0.00001, loadFactor: math.MaxFloat64, cushion: 1, maxFeeXRP: "2", expected: "2"},
		{name: "minimum float load factor", baseFeeXRP: 0.00001, loadFactor: math.SmallestNonzeroFloat64, cushion: 1, maxFeeXRP: "2", expected: "0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			maxFee, err := ParseFeeXRP(test.maxFeeXRP)
			require.NoError(t, err)

			actual, err := NetworkFeeDrops(test.baseFeeXRP, test.loadFactor, test.cushion, maxFee)
			require.NoError(t, err)
			actualXRP, err := actual.RoundHalfUp().XRPString()
			require.NoError(t, err)
			require.Equal(t, test.expected, actualXRP)
		})
	}
}

func TestNetworkFeeDropsFractionalBaseFee(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		baseFeeXRP    float64
		loadFactor    float64
		expectedXRP   string
		expectedDrops string
	}{
		{name: "half drop rounds upward", baseFeeXRP: 0.0000125, loadFactor: 1, expectedXRP: "0.000013", expectedDrops: "13"},
		{name: "half drop is multiplied before rounding", baseFeeXRP: 0.0000125, loadFactor: 2, expectedXRP: "0.000025", expectedDrops: "25"},
		{name: "less than half drop rounds downward", baseFeeXRP: 0.0000124, loadFactor: 1, expectedXRP: "0.000012", expectedDrops: "12"},
	}

	maxFee, err := ParseFeeXRP("2")
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := NetworkFeeDrops(test.baseFeeXRP, test.loadFactor, 1, maxFee)
			require.NoError(t, err)

			actualXRP, err := actual.RoundHalfUp().XRPString()
			require.NoError(t, err)
			require.Equal(t, test.expectedXRP, actualXRP)

			actualDrops, err := actual.RoundHalfUp().WholeString()
			require.NoError(t, err)
			require.Equal(t, test.expectedDrops, actualDrops)
		})
	}
}

// TestNetworkFeeDropsKeepsFractionalDrops pins the contract callers rely on:
// the network fee keeps its fractional drop so a base-fee factor multiplies the
// exact value. Rounding first would scale the rounding error by the factor.
// CalculateFee owns the rounding, and TestCalculateFee pins the drops it pays.
func TestNetworkFeeDropsKeepsFractionalDrops(t *testing.T) {
	t.Parallel()

	maxFee, err := ParseFeeXRP("2")
	require.NoError(t, err)

	// A ten drop base fee under a fractional load factor is 12.4 drops.
	netFee, err := NetworkFeeDrops(0.00001, 1.24, 1, maxFee)
	require.NoError(t, err)
	require.False(t, netFee.IsWhole())

	tenthDrops, err := netFee.Mul(10).WholeString()
	require.NoError(t, err)
	require.Equal(t, "124", tenthDrops)
}

func TestNetworkFeeDropsRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	maxFee, err := ParseFeeXRP("2")
	require.NoError(t, err)

	_, err = NetworkFeeDrops(-1, 1, 1.2, maxFee)
	require.ErrorIs(t, err, ErrInvalidFeeValue)

	_, err = NetworkFeeDrops(math.NaN(), 1, 1.2, maxFee)
	require.ErrorIs(t, err, ErrInvalidFeeValue)

	_, err = NetworkFeeDrops(math.Inf(1), 1, 1.2, maxFee)
	require.ErrorIs(t, err, ErrInvalidFeeValue)

	_, err = NetworkFeeDrops(0.00001, -1, 1.2, maxFee)
	require.ErrorIs(t, err, ErrInvalidFeeValue)
}

func TestParseFeeXRP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "canonical fee", value: "0.000025", expected: "25"},
		{name: "long non-canonical fee", value: "2.000000000000000000000000", expected: "2000000"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fee, err := ParseFeeXRP(test.value)
			require.NoError(t, err)

			actual, err := fee.WholeString()
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestParseFeeXRPRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	_, err := ParseFeeXRP("0.0000001")
	require.ErrorIs(t, err, ErrFeeHasTooManyDecimals)

	for _, value := range []string{"invalid", "1/2", "0x10", "-1"} {
		_, err = ParseFeeXRP(value)
		require.ErrorIs(t, err, ErrInvalidFeeValue)
	}
}
