package crypto

// Algorithm defines the interface for cryptographic algorithms used in XRPL key generation and signing.
type Algorithm interface {
	Prefix() byte
	FamilySeedPrefix() []byte
}
