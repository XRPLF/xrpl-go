// Package mptcrypto provides CGo bindings to the XRPLF/mpt-crypto C library
// for XLS-96 Confidential MPT Transfers: ElGamal encryption, ZK proofs,
// Pedersen commitments, and context hash computation.
package mptcrypto

// Size constants (bytes), matching mpt_utility.h defines.
const (
	PrivKeySize        = 32
	PubKeySize         = 33
	BlindingFactorSize = 32
	CiphertextSize     = 66 // two compressed EC points (C1 || C2)
)
