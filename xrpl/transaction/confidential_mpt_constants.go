package transaction

import "github.com/Peersyst/xrpl-go/pkg/typecheck"

// CompressedPointLen is the hex-encoded length of a 33-byte compressed EC point.
// Used for compressed public keys (IssuerEncryptionKey, AuditorEncryptionKey) and
// Pedersen commitments (BalanceCommitment, AmountCommitment).
const CompressedPointLen = 66

// CiphertextLen is the hex-encoded length of a 66-byte ElGamal ciphertext (two compressed EC points).
const CiphertextLen = 2 * CompressedPointLen

// BlindingFactorLen is the hex-encoded length of a 32-byte blinding factor scalar.
const BlindingFactorLen = 64

// SchnorrProofLen is the hex-encoded length of a 65-byte Schnorr proof of knowledge (T[33] || s[32]).
const SchnorrProofLen = 130

// IsValidCompressedEncryptionKey checks if the given hex string is a valid
// 33-byte compressed EC public key (66 hex chars).
// Used for IssuerEncryptionKey and AuditorEncryptionKey per XLS-96.
func IsValidCompressedEncryptionKey(key string) bool {
	return len(key) == CompressedPointLen && typecheck.IsHex(key)
}

// IsValidBlindingFactor checks if the given hex string is a valid 32-byte blinding factor (64 hex chars).
func IsValidBlindingFactor(bf string) bool {
	return len(bf) == BlindingFactorLen && typecheck.IsHex(bf)
}

// IsValidSchnorrProof checks if the given hex string is a valid 65-byte Schnorr proof (130 hex chars).
func IsValidSchnorrProof(proof string) bool {
	return len(proof) == SchnorrProofLen && typecheck.IsHex(proof)
}

// IsValidCiphertext checks if the given hex string is a valid 66-byte ElGamal ciphertext (132 hex chars).
func IsValidCiphertext(s string) bool {
	return len(s) == CiphertextLen && typecheck.IsHex(s)
}

// IsValidCommitment checks if the given hex string is a valid 33-byte Pedersen commitment (66 hex chars).
func IsValidCommitment(s string) bool {
	return len(s) == CompressedPointLen && typecheck.IsHex(s)
}

// IsValidHexBlob checks if the given string is a non-empty valid hex-encoded blob.
// Used for variable-length fields like ZKProof bundles where the spec does not define a fixed size.
func IsValidHexBlob(s string) bool {
	return len(s) > 0 && typecheck.IsHex(s)
}
