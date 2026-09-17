package mptcrypto

import "errors"

var (
	// ErrCgoRequired is returned when native confidential MPT operations are unavailable.
	ErrCgoRequired = errors.New(
		"mptcrypto: CGo is required for confidential MPT operations; " +
			"rebuild with CGO_ENABLED=1 and vendored mpt-crypto libraries",
	)
	// ErrInvalidAmountRange is returned when a decryption search range is invalid.
	ErrInvalidAmountRange = errors.New("mptcrypto: invalid amount range")
	// ErrInvalidCiphertext is returned when a ciphertext does not decode to two curve
	// points, or when a homomorphic result has no compressed encoding.
	ErrInvalidCiphertext = errors.New("mptcrypto: invalid ciphertext")
	// ErrInvalidPublicKey is returned when a public key does not decode to a curve point.
	ErrInvalidPublicKey = errors.New("mptcrypto: invalid public key")
)
