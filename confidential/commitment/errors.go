package commitment

import "errors"

var (
	// ErrInvalidBlindingFactor is returned when a blinding factor is not valid hex or has an unexpected byte length.
	ErrInvalidBlindingFactor = errors.New("commitment: invalid blinding factor")
	// ErrCommitmentFailed is returned when the underlying C commitment computation fails.
	ErrCommitmentFailed = errors.New("commitment: computation failed")
)
