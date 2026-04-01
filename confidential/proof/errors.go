package proof

import "errors"

var (
	// ErrInvalidAddress is returned when a classic XRPL address cannot be decoded.
	ErrInvalidAddress = errors.New("proof: invalid classic address")
	// ErrInvalidIssuanceID is returned when an issuance ID is not 48 hex characters (24 bytes).
	ErrInvalidIssuanceID = errors.New("proof: invalid issuance ID")
	// ErrInvalidContextHash is returned when a context hash has an unexpected byte length.
	ErrInvalidContextHash = errors.New("proof: invalid context hash")
	// ErrInvalidProof is returned when a proof has an unexpected byte length.
	ErrInvalidProof = errors.New("proof: invalid proof")
	// ErrInvalidPubKey is returned when a public key has an unexpected byte length.
	ErrInvalidPubKey = errors.New("proof: invalid public key")
	// ErrInvalidPrivKey is returned when a private key has an unexpected byte length.
	ErrInvalidPrivKey = errors.New("proof: invalid private key")
	// ErrInvalidCiphertext is returned when a ciphertext has an unexpected byte length.
	ErrInvalidCiphertext = errors.New("proof: invalid ciphertext")
	// ErrInvalidCommitment is returned when a commitment has an unexpected byte length.
	ErrInvalidCommitment = errors.New("proof: invalid commitment")
	// ErrInvalidBlindingFactor is returned when a blinding factor has an unexpected byte length.
	ErrInvalidBlindingFactor = errors.New("proof: invalid blinding factor length")
	// ErrProofGenerationFailed is returned when the underlying C proof generation call fails.
	ErrProofGenerationFailed = errors.New("proof: proof generation failed")
	// ErrProofVerificationFailed is returned when the underlying C proof verification call fails.
	ErrProofVerificationFailed = errors.New("proof: proof verification failed")
	// ErrContextHashFailed is returned when the underlying C context hash computation fails.
	ErrContextHashFailed = errors.New("proof: context hash computation failed")
)
