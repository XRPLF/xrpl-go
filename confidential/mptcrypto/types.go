// Package mptcrypto provides CGo bindings to the XRPLF/mpt-crypto C library
// for XLS-96 Confidential MPT Transfers: ElGamal encryption, ZK proofs,
// Pedersen commitments, and context hash computation.
package mptcrypto

import (
	"math"

	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
)

// MaxParticipants is the ceiling the C API imposes on a Confidential Send.
const MaxParticipants = math.MaxUint8 // C API uses uint8_t for participant count

// PrivateKey is a fixed-size ElGamal private key.
type PrivateKey [mptsizes.PrivKeySize]byte

// PublicKey is a fixed-size compressed ElGamal public key.
type PublicKey [mptsizes.PubKeySize]byte

// BlindingFactor is a fixed-size scalar used for encryption and commitments.
type BlindingFactor [mptsizes.BlindingFactorSize]byte

// Ciphertext is a fixed-size ElGamal ciphertext.
type Ciphertext [mptsizes.CiphertextSize]byte

// Commitment is a fixed-size compressed Pedersen commitment.
type Commitment [mptsizes.CommitmentSize]byte

// ContextHash binds a proof to its transaction context.
type ContextHash [mptsizes.HashOutputSize]byte

// Participant represents a party in a Confidential Send transaction.
type Participant struct {
	PubKey     PublicKey
	Ciphertext Ciphertext
}

// PedersenProofParams holds the parameters required to generate a Pedersen linkage proof.
type PedersenProofParams struct {
	Commitment     Commitment
	Amount         uint64
	Ciphertext     Ciphertext
	BlindingFactor BlindingFactor
}
