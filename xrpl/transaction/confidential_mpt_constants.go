package transaction

import (
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
)

// Every length here is hex-encoded, so twice the byte size mpt-crypto emits. Deriving them
// from mptsizes keeps the models and the vendored C library from drifting apart when
// `make update-mpt-crypto` bumps a proof format.

// CompressedPointLen is the hex-encoded length of a 33-byte compressed EC point.
// Used for compressed public keys (HolderEncryptionKey, IssuerEncryptionKey,
// AuditorEncryptionKey) and Pedersen commitments (BalanceCommitment, AmountCommitment).
// Derived from the byte length the point validation itself uses, so the two cannot drift.
const CompressedPointLen = 2 * crypto.CompressedSECP256K1PointByteLength

// CiphertextLen is the hex-encoded length of a 66-byte ElGamal ciphertext (two compressed EC points).
const CiphertextLen = 2 * mptsizes.CiphertextSize

// BlindingFactorLen is the hex-encoded length of a 32-byte blinding factor scalar.
const BlindingFactorLen = 2 * mptsizes.BlindingFactorSize

// SchnorrProofLen is the hex-encoded length of a 64-byte Schnorr proof of knowledge.
// Carried by ConfidentialMPTConvert when registering a holder encryption key. It proves
// ownership of the private key behind the ElGamal public key.
const SchnorrProofLen = 2 * mptsizes.SchnorrProofSize

// SendProofLen is the hex-encoded length of a 946-byte confidential send proof bundle:
// a 192-byte compact sigma proof and a 754-byte aggregated Bulletproof range proof.
const SendProofLen = 2 * mptsizes.SendProofSize

// ConvertBackProofLen is the hex-encoded length of an 816-byte confidential convert back
// proof bundle: a 128-byte compact sigma proof and a 688-byte Bulletproof range proof.
const ConvertBackProofLen = 2 * mptsizes.ConvertBackProofSize

// ClawbackProofLen is the hex-encoded length of a 64-byte AND-composed compact Clawback
// sigma proof. It proves that the issuer's encrypted balance mirror encrypts the plaintext
// MPTAmount. It matches SchnorrProofLen in size but proves a different statement.
const ClawbackProofLen = 2 * mptsizes.CompactClawbackProofSize
