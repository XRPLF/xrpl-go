package mptcrypto

import (
	"errors"
	"fmt"
	"math"
)

var (
	// ErrCgoRequired is returned when native confidential MPT operations are unavailable.
	ErrCgoRequired = errors.New(
		"mptcrypto: CGo is required for confidential MPT operations; " +
			"rebuild with CGO_ENABLED=1 and vendored mpt-crypto libraries",
	)
	// ErrInvalidAmountRange is returned when a decryption search range is invalid.
	ErrInvalidAmountRange = errors.New("mptcrypto: invalid amount range")
)

func validateAmountRange(rangeLow, rangeHigh uint64) error {
	if rangeLow > rangeHigh {
		return fmt.Errorf("%w: low %d exceeds high %d", ErrInvalidAmountRange, rangeLow, rangeHigh)
	}
	if rangeHigh == math.MaxUint64 {
		return fmt.Errorf("%w: high must be less than %d", ErrInvalidAmountRange, uint64(math.MaxUint64))
	}
	return nil
}
