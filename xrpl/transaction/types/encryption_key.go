package types

// EncryptionKey returns a pointer to the given encryption key string.
// Used as a convenience for setting optional encryption key fields in transactions.
func EncryptionKey(key string) *string {
	return &key
}
