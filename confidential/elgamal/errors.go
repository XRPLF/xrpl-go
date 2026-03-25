package elgamal

import "errors"

var (
	// ErrInvalidKeyLength is returned when a key has an unexpected byte length.
	ErrInvalidKeyLength = errors.New("elgamal: invalid key length")
	// ErrInvalidCiphertextLength is returned when a ciphertext has an unexpected byte length.
	ErrInvalidCiphertextLength = errors.New("elgamal: invalid ciphertext length")
	// ErrInvalidBlindingFactorLength is returned when a blinding factor has an unexpected byte length.
	ErrInvalidBlindingFactorLength = errors.New("elgamal: invalid blinding factor length")
	// ErrEncryptFailed is returned when the underlying C encryption call fails.
	ErrEncryptFailed = errors.New("elgamal: encryption failed")
	// ErrDecryptFailed is returned when the underlying C decryption call fails.
	ErrDecryptFailed = errors.New("elgamal: decryption failed")
)
