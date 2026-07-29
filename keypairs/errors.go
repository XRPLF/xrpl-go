package keypairs

import (
	"errors"
	"fmt"
)

var (
	// keypairs

	// ErrInvalidSignature is returned when the derived keypair did not generate a verifiable signature.
	ErrInvalidSignature = errors.New("derived keypair did not generate verifiable signature")

	// ErrRandomizerRequired is returned when GenerateSeed is called with empty entropy
	// and no randomizer to source entropy from.
	ErrRandomizerRequired = errors.New("keypairs: randomizer required when entropy is empty")

	// ErrInvalidEntropyLength is returned when caller-supplied entropy is not the expected length.
	// Use errors.Is to detect this case without importing the underlying codec package.
	ErrInvalidEntropyLength = errors.New("keypairs: invalid entropy length")

	// crypto

	// ErrInvalidCryptoImplementation is returned when the key does not match any crypto implementation.
	ErrInvalidCryptoImplementation = errors.New("not a valid crypto implementation")

	// ErrInvalidPrivateKeyFormat is returned when a private key does not match an accepted format.
	// It wraps ErrInvalidCryptoImplementation for backward-compatible errors.Is checks.
	ErrInvalidPrivateKeyFormat = fmt.Errorf("%w: invalid private key format", ErrInvalidCryptoImplementation)

	// ErrInvalidPublicKeyFormat is returned when a public key does not match an accepted format.
	// It wraps ErrInvalidCryptoImplementation for backward-compatible errors.Is checks.
	ErrInvalidPublicKeyFormat = fmt.Errorf("%w: invalid public key format", ErrInvalidCryptoImplementation)
)
