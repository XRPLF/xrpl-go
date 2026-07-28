package transaction

import "github.com/Peersyst/xrpl-go/pkg/typecheck"

// CompressedPointLen is the hex-encoded length of a 33-byte compressed EC point.
// Used for compressed public keys (IssuerEncryptionKey, AuditorEncryptionKey) and
// Pedersen commitments (BalanceCommitment, AmountCommitment).
const CompressedPointLen = 66

// CiphertextLen is the hex-encoded length of a 66-byte ElGamal ciphertext (two compressed EC points).
const CiphertextLen = 2 * CompressedPointLen

// PrivKeyLen is the hex-encoded length of a 32-byte private key scalar.
const PrivKeyLen = 64

// BlindingFactorLen is the hex-encoded length of a 32-byte blinding factor scalar.
const BlindingFactorLen = 64

// SchnorrProofLen is the hex-encoded length of a 64-byte compact Schnorr proof of knowledge.
const SchnorrProofLen = 128

// SendProofLen is the hex-encoded length of a 946-byte confidential send proof bundle.
const SendProofLen = 1892

// ConvertBackProofLen is the hex-encoded length of an 816-byte confidential convert back proof bundle.
const ConvertBackProofLen = 1632

// ClawbackProofLen is the hex-encoded length of a 64-byte confidential clawback proof.
const ClawbackProofLen = 128

// IsValidPrivKey checks if the given hex string is a valid 32-byte private key scalar (64 hex chars).
func IsValidPrivKey(key string) bool {
	return len(key) == PrivKeyLen && typecheck.IsHex(key)
}

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

// IsValidSchnorrProof checks if the given hex string is a valid 64-byte Schnorr proof (128 hex chars).
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

// IsValidFixedHexBlob checks if the given string is valid hex with an exact encoded length.
func IsValidFixedHexBlob(s string, hexLen int) bool {
	return len(s) == hexLen && typecheck.IsHex(s)
}
