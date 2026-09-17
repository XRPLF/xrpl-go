package elgamal

import "errors"

var (
	// ErrInvalidKey is returned when a key is not valid hex or has an unexpected byte length.
	ErrInvalidKey = errors.New("elgamal: invalid key")
	// ErrInvalidCiphertext is returned when a ciphertext is not valid hex or has an unexpected byte length.
	ErrInvalidCiphertext = errors.New("elgamal: invalid ciphertext")
	// ErrInvalidAddress is returned when an account is not a valid classic or X-address.
	ErrInvalidAddress = errors.New("elgamal: invalid address")
	// ErrInvalidIssuanceID is returned when an issuance ID is not valid hex or has an unexpected byte length.
	ErrInvalidIssuanceID = errors.New("elgamal: invalid issuance ID")
	// ErrInvalidBlindingFactor is returned when a blinding factor is not valid hex or has an unexpected byte length.
	ErrInvalidBlindingFactor = errors.New("elgamal: invalid blinding factor")
	// ErrEncryptFailed is returned when the underlying C encryption call fails.
	ErrEncryptFailed = errors.New("elgamal: encryption failed")
	// ErrDecryptFailed is returned when the underlying C decryption call fails.
	ErrDecryptFailed = errors.New("elgamal: decryption failed")
	// ErrCiphertextArithmetic is returned when a homomorphic add or subtract has no result,
	// such as a subtraction of a ciphertext from itself.
	ErrCiphertextArithmetic = errors.New("elgamal: ciphertext arithmetic failed")
	// ErrInvalidAmountRange is returned when a decryption search range is invalid.
	ErrInvalidAmountRange = errors.New("elgamal: invalid amount range")
)
