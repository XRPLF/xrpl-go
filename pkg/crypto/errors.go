package crypto

import "errors"

var (
	// keypair

	// ErrValidatorKeypairDerivation is returned when a validator keypair is attempted to be derived
	ErrValidatorKeypairDerivation = errors.New("validator keypair derivation not supported")
	// ErrInvalidPrivateKey is returned when a private key is invalid
	ErrInvalidPrivateKey = errors.New("invalid private key")
	// ErrInvalidMessage is returned when a message is required but not provided
	ErrInvalidMessage = errors.New("message is required")
	// ErrValidatorNotSupported is returned when a validator keypair is used with the ED25519 algorithm.
	ErrValidatorNotSupported = errors.New("validator keypairs can not use Ed25519")
	// ErrDerivedKeyIsZero is returned when the derived private key scalar is zero (mod N).
	ErrDerivedKeyIsZero = errors.New("derived key is zero")
)
