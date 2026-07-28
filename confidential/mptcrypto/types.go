// Package mptcrypto provides CGo bindings to the XRPLF/mpt-crypto C library
// for XLS-96 Confidential MPT Transfers: ElGamal encryption, ZK proofs,
// Pedersen commitments, and context hash computation.
package mptcrypto

import "math"

// Size constants (bytes), matching mpt_utility.h defines.
const (
	PrivKeySize        = 32
	PubKeySize         = 33
	BlindingFactorSize = 32
	CiphertextSize     = 66 // two compressed EC points (C1 || C2)

	AccountIDSize  = 20
	IssuanceIDSize = 24
	HashOutputSize = 32 // kMPT_HALF_SHA_SIZE -- output size of context hash functions
	CommitmentSize = 33 // compressed Pedersen commitment point

	SchnorrProofSize            = 64
	SingleBulletproofSize       = 688
	DoubleBulletproofSize       = 754
	CompactClawbackProofSize    = 64
	CompactConvertBackProofSize = 128
	CompactSendProofSize        = 192
	ConvertBackProofSize        = CompactConvertBackProofSize + SingleBulletproofSize
	SendProofSize               = CompactSendProofSize + DoubleBulletproofSize

	MaxParticipants = math.MaxUint8 // C API uses uint8_t for participant count
)

// Participant represents a party in a Confidential Send transaction.
type Participant struct {
	PubKey     [PubKeySize]byte
	Ciphertext [CiphertextSize]byte
}

// PedersenProofParams holds the parameters required to generate a Pedersen linkage proof.
type PedersenProofParams struct {
	Commitment     [CommitmentSize]byte
	Amount         uint64
	Ciphertext     [CiphertextSize]byte
	BlindingFactor [BlindingFactorSize]byte
}
