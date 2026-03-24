package transaction

import "github.com/Peersyst/xrpl-go/pkg/typecheck"

// EncryptionKeyCompressedLen is the hex-encoded length of a compressed EC public key (33 bytes).
const EncryptionKeyCompressedLen = 66

// EncryptionKeyUncompressedLen is the hex-encoded length of an uncompressed EC public key (64 bytes).
// Used for HolderEncryptionKey in ConfidentialMPTConvert per XLS-96.
const EncryptionKeyUncompressedLen = 128

// BlindingFactorLen is the hex-encoded length of a 32-byte blinding factor scalar.
const BlindingFactorLen = 64

// SchnorrProofLen is the hex-encoded length of a 65-byte Schnorr proof of knowledge (T[33] || s[32]).
const SchnorrProofLen = 130

// IsValidCompressedEncryptionKey checks if the given hex string is a valid
// 33-byte compressed EC public key (66 hex chars).
// Used for IssuerEncryptionKey and AuditorEncryptionKey per XLS-96.
func IsValidCompressedEncryptionKey(key string) bool {
	return len(key) == EncryptionKeyCompressedLen && typecheck.IsHex(key)
}

// IsValidUncompressedEncryptionKey checks if the given hex string is a valid
// 64-byte uncompressed EC public key (128 hex chars).
// Used for HolderEncryptionKey in ConfidentialMPTConvert per XLS-96.
func IsValidUncompressedEncryptionKey(key string) bool {
	return len(key) == EncryptionKeyUncompressedLen && typecheck.IsHex(key)
}

// IsValidBlindingFactor checks if the given hex string is a valid 32-byte blinding factor (64 hex chars).
func IsValidBlindingFactor(bf string) bool {
	return len(bf) == BlindingFactorLen && typecheck.IsHex(bf)
}

// IsValidSchnorrProof checks if the given hex string is a valid 65-byte Schnorr proof (130 hex chars).
func IsValidSchnorrProof(proof string) bool {
	return len(proof) == SchnorrProofLen && typecheck.IsHex(proof)
}

// IsValidHexBlob checks if the given string is a non-empty valid hex-encoded blob.
func IsValidHexBlob(s string) bool {
	return len(s) > 0 && typecheck.IsHex(s)
}
