// Package currency provides utilities for working with XRP native currency conversions and calculations.
package currency

import (
	"errors"
	"strconv"
	"strings"
)

const (
	// DropsPerXrp is the number of drops equivalent to one XRP.
	DropsPerXrp float64 = 1000000
	// MaxFractionLength is the maximum allowed decimal places in an XRP value.
	MaxFractionLength uint = 6
	// NativeCurrencySymbol is the symbol representing the native XRP currency.
	NativeCurrencySymbol string = "XRP"
)

// XrpToDrops converts an amount in XRP to an amount in drops.
func XrpToDrops(value string) (string, error) {
	if _, after, ok := strings.Cut(value, "."); ok && len(after) > int(MaxFractionLength) {
		return "", errors.New("xrp to drops: value has too many decimals")
	}

	xrpFloat, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", err
	}

	dropsFloat := xrpFloat * DropsPerXrp
	return strconv.FormatFloat(dropsFloat, 'f', -1, 64), nil
}

// DropsToXrp converts an amount of drops into an amount of XRP.
func DropsToXrp(value string) (string, error) {
	dropUint, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return "", err
	}

	xrpFloat := float64(dropUint) / DropsPerXrp

	return strconv.FormatFloat(xrpFloat, 'f', -1, 64), nil
}
