package mptcrypto

import (
	"errors"
	"fmt"
	"math"
)

// Ciphertext is a fixed-size ElGamal ciphertext.
type Ciphertext = [CiphertextSize]byte

// PrivateKey is a fixed-size ElGamal private key.
type PrivateKey = [PrivKeySize]byte

// ErrInvalidAmountRange is returned when a decryption search range is invalid.
var ErrInvalidAmountRange = errors.New("mptcrypto: invalid amount range")

func validateAmountRange(rangeLow, rangeHigh uint64) error {
	if rangeLow > rangeHigh {
		return fmt.Errorf("%w: low %d exceeds high %d", ErrInvalidAmountRange, rangeLow, rangeHigh)
	}
	if rangeHigh == math.MaxUint64 {
		return fmt.Errorf("%w: high must be less than %d", ErrInvalidAmountRange, uint64(math.MaxUint64))
	}
	return nil
}
