package types

// HexBlob returns a pointer to the given hex-encoded blob string.
// Used as a convenience for setting optional hex blob fields in transactions
// (e.g. AuditorEncryptedAmount, ZKProof).
func HexBlob(value string) *string {
	return &value
}
