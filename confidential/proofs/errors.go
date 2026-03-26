package proofs

import "errors"

var (
	// ErrInvalidAddress is returned when a classic XRPL address cannot be decoded.
	ErrInvalidAddress = errors.New("proofs: invalid classic address")
	// ErrInvalidIssuanceIDLength is returned when an issuance ID is not 48 hex characters (24 bytes).
	ErrInvalidIssuanceIDLength = errors.New("proofs: issuance ID must be 48 hex characters (24 bytes)")
	// ErrInvalidContextHash is returned when a context hash has an unexpected byte length.
	ErrInvalidContextHash = errors.New("proofs: invalid context hash")
	// ErrInvalidProofLength is returned when a proof has an unexpected byte length.
	ErrInvalidProofLength = errors.New("proofs: invalid proof length")
	// ErrInvalidPubKeyLength is returned when a public key has an unexpected byte length.
	ErrInvalidPubKeyLength = errors.New("proofs: invalid public key length")
	// ErrInvalidPrivKeyLength is returned when a private key has an unexpected byte length.
	ErrInvalidPrivKeyLength = errors.New("proofs: invalid private key length")
	// ErrInvalidCiphertextLength is returned when a ciphertext has an unexpected byte length.
	ErrInvalidCiphertextLength = errors.New("proofs: invalid ciphertext length")
	// ErrInvalidCommitmentLength is returned when a commitment has an unexpected byte length.
	ErrInvalidCommitmentLength = errors.New("proofs: invalid commitment length")
	// ErrInvalidBlindingFactor is returned when a blinding factor has an unexpected byte length.
	ErrInvalidBlindingFactor = errors.New("proofs: invalid blinding factor length")
	// ErrProofGenerationFailed is returned when the underlying C proof generation call fails.
	ErrProofGenerationFailed = errors.New("proofs: proof generation failed")
	// ErrProofVerificationFailed is returned when the underlying C proof verification call fails.
	ErrProofVerificationFailed = errors.New("proofs: proof verification failed")
	// ErrContextHashFailed is returned when the underlying C context hash computation fails.
	ErrContextHashFailed = errors.New("proofs: context hash computation failed")
)
