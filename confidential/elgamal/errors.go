package elgamal

import "errors"

var (
	// ErrInvalidKey is returned when a key is not valid hex or has an unexpected byte length.
	ErrInvalidKey = errors.New("elgamal: invalid key")
	// ErrInvalidCiphertext is returned when a ciphertext is not valid hex or has an unexpected byte length.
	ErrInvalidCiphertext = errors.New("elgamal: invalid ciphertext")
	// ErrInvalidBlindingFactor is returned when a blinding factor is not valid hex or has an unexpected byte length.
	ErrInvalidBlindingFactor = errors.New("elgamal: invalid blinding factor")
	// ErrEncryptFailed is returned when the underlying C encryption call fails.
	ErrEncryptFailed = errors.New("elgamal: encryption failed")
	// ErrDecryptFailed is returned when the underlying C decryption call fails.
	ErrDecryptFailed = errors.New("elgamal: decryption failed")
)
