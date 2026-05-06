package currency

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXrpToDrops(t *testing.T) {
	tt := []struct {
		name        string
		xrp         string
		drops       string
		expectedErr error
	}{
		{
			name:        "XRP to Drops (no decimals)",
			xrp:         "1",
			drops:       "1000000",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (decimals)(1)",
			xrp:         "3.456789",
			drops:       "3456789",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (decimals)(2)",
			xrp:         "0.000001",
			drops:       "1",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (minimum XRP)",
			xrp:         "0",
			drops:       "0",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (minimum XRP with fractional zeros)",
			xrp:         "0.000000",
			drops:       "0",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (decimals)(4)",
			xrp:         "3.400000",
			drops:       "3400000",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (scientific notation)",
			xrp:         "1e-6",
			drops:       "1",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (uppercase scientific notation)",
			xrp:         "1E-6",
			drops:       "1",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (leading zeros)",
			xrp:         "000001.000000",
			drops:       "1000000",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (largest exactly representable drops)",
			xrp:         "9007199254.740992",
			drops:       "9007199254740992",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (first unsafe integer drops)",
			xrp:         "9007199254.740993",
			drops:       "9007199254740993",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (valid high value preserves drops)",
			xrp:         "12345678901.234567",
			drops:       "12345678901234567",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (maximum XRP minus one drop)",
			xrp:         "99999999999.999999",
			drops:       "99999999999999999",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (maximum XRP)",
			xrp:         "100000000000",
			drops:       "100000000000000000",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (maximum XRP with fractional zeros)",
			xrp:         "100000000000.000000",
			drops:       "100000000000000000",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (too many decimals)",
			xrp:         "0.0000001",
			drops:       "1",
			expectedErr: ErrXrpToDropsTooManyDecimals,
		},
		{
			name:        "XRP to Drops (scientific notation too precise)",
			xrp:         "1e-7",
			drops:       "",
			expectedErr: ErrXrpToDropsTooManyDecimals,
		},
		{
			name:        "XRP to Drops (one drop below minimum XRP)",
			xrp:         "-0.000001",
			drops:       "",
			expectedErr: ErrXrpToDropsNegativeValue,
		},
		{
			name:        "XRP to Drops (negative XRP)",
			xrp:         "-1",
			drops:       "",
			expectedErr: ErrXrpToDropsNegativeValue,
		},
		{
			name:        "XRP to Drops (one drop over maximum XRP)",
			xrp:         "100000000000.000001",
			drops:       "",
			expectedErr: ErrXrpToDropsExceedsMax,
		},
		{
			name:        "XRP to Drops (over maximum XRP)",
			xrp:         "100000000001",
			drops:       "",
			expectedErr: ErrXrpToDropsExceedsMax,
		},
		{
			name:        "XRP to Drops (invalid input)",
			xrp:         "abc",
			drops:       "",
			expectedErr: ErrXrpToDropsInvalidValue,
		},
		{
			name:        "XRP to Drops (fraction syntax)",
			xrp:         "1/2",
			drops:       "",
			expectedErr: ErrXrpToDropsInvalidValue,
		},
		{
			name:        "XRP to Drops (hexadecimal syntax)",
			xrp:         "0x1",
			drops:       "",
			expectedErr: ErrXrpToDropsInvalidValue,
		},
		{
			name:        "XRP to Drops (NaN)",
			xrp:         "NaN",
			drops:       "",
			expectedErr: ErrXrpToDropsInvalidValue,
		},
		{
			name:        "XRP to Drops (positive infinity)",
			xrp:         "+Inf",
			drops:       "",
			expectedErr: ErrXrpToDropsInvalidValue,
		},
		{
			name:        "XRP to Drops (negative infinity)",
			xrp:         "-Inf",
			drops:       "",
			expectedErr: ErrXrpToDropsInvalidValue,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			drops, err := XrpToDrops(tc.xrp)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.drops, drops)
			}
		})
	}
}

func TestDropsToXrp(t *testing.T) {
	tt := []struct {
		name        string
		drops       string
		xrp         string
		expectedErr error
	}{
		{
			name:        "Drops to XRP (whole number)",
			drops:       "1000000",
			xrp:         "1",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (decimal)",
			drops:       "1234567",
			xrp:         "1.234567",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (minimum drops)",
			drops:       "0",
			xrp:         "0",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (small amount)",
			drops:       "1",
			xrp:         "0.000001",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (large amount)",
			drops:       "123456789000000",
			xrp:         "123456789",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (positive exponent)",
			drops:       "1e6",
			xrp:         "1",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (leading zeros)",
			drops:       "000001",
			xrp:         "0.000001",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (whole drops with fractional zeros)",
			drops:       "1.0",
			xrp:         "0.000001",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (largest exactly representable drops)",
			drops:       "9007199254740992",
			xrp:         "9007199254.740992",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (first unsafe integer drops)",
			drops:       "9007199254740993",
			xrp:         "9007199254.740993",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (valid high value preserves drops)",
			drops:       "12345678901234567",
			xrp:         "12345678901.234567",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (maximum drops minus one drop)",
			drops:       "99999999999999999",
			xrp:         "99999999999.999999",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (maximum drops)",
			drops:       "100000000000000000",
			xrp:         "100000000000",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (one drop below minimum drops)",
			drops:       "-1",
			xrp:         "",
			expectedErr: ErrDropsToXrpNegativeValue,
		},
		{
			name:        "Drops to XRP (one drop over maximum drops)",
			drops:       "100000000000000001",
			xrp:         "",
			expectedErr: ErrDropsToXrpExceedsMax,
		},
		{
			name:        "Drops to XRP (fractional drops)",
			drops:       "1.1",
			xrp:         "",
			expectedErr: ErrDropsToXrpFractionalDrops,
		},
		{
			name:        "Drops to XRP (invalid input)",
			drops:       "abc",
			xrp:         "",
			expectedErr: ErrDropsToXrpInvalidValue,
		},
		{
			name:        "Drops to XRP (empty input)",
			drops:       "",
			xrp:         "",
			expectedErr: ErrDropsToXrpInvalidValue,
		},
		{
			name:        "Drops to XRP (NaN)",
			drops:       "NaN",
			xrp:         "",
			expectedErr: ErrDropsToXrpInvalidValue,
		},
		{
			name:        "Drops to XRP (fraction syntax)",
			drops:       "1/2",
			xrp:         "",
			expectedErr: ErrDropsToXrpInvalidValue,
		},
		{
			name:        "Drops to XRP (hexadecimal syntax)",
			drops:       "0x1",
			xrp:         "",
			expectedErr: ErrDropsToXrpInvalidValue,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			xrp, err := DropsToXrp(tc.drops)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.xrp, xrp)
			}
		})
	}
}

func TestNativeCurrencyRoundTripPreservesDrops(t *testing.T) {
	tt := []struct {
		name  string
		drops string
	}{
		{
			name:  "near precision boundary",
			drops: "9007199254740991",
		},
		{
			name:  "first unsafe integer drops",
			drops: "9007199254740993",
		},
		{
			name:  "valid high value",
			drops: "12345678901234567",
		},
		{
			name:  "minimum drops",
			drops: "0",
		},
		{
			name:  "minimum positive drops",
			drops: "1",
		},
		{
			name:  "maximum drops minus one drop",
			drops: "99999999999999999",
		},
		{
			name:  "maximum drops",
			drops: "100000000000000000",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			xrp, err := DropsToXrp(tc.drops)
			require.NoError(t, err)

			drops, err := XrpToDrops(xrp)
			require.NoError(t, err)

			require.Equal(t, tc.drops, drops)
		})
	}
}
