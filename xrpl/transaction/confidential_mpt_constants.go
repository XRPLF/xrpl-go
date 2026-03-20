package transaction

import "github.com/Peersyst/xrpl-go/pkg/typecheck"

// EncryptionKeyCompressedLen is the hex-encoded length of a compressed EC public key (33 bytes).
const EncryptionKeyCompressedLen = 66

// IsValidCompressedEncryptionKey checks if the given hex string is a valid
// 33-byte compressed EC public key (66 hex chars).
// Used for IssuerEncryptionKey and AuditorEncryptionKey per XLS-96.
func IsValidCompressedEncryptionKey(key string) bool {
	return len(key) == EncryptionKeyCompressedLen && typecheck.IsHex(key)
}
