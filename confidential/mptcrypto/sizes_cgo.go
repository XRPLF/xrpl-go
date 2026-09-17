//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package mptcrypto

/*
#include <secp256k1_mpt.h>
*/
import "C"

import "github.com/Peersyst/xrpl-go/pkg/mptsizes"

// The wrappers in this package hand pointers to Go arrays sized from pkg/mptsizes to C
// functions that take fixed-length array parameters. Those parameters decay to bare
// pointers, so the C library performs no bounds check: a vendored header that grows an
// output size without pkg/mptsizes following would let C write past the Go allocation.
//
// confidential/deps/update.sh replaces the headers wholesale and cannot know about the Go
// side, so every constant is pinned to its define here instead. sizeMatch is indexed by
// the difference between the two, which makes any divergence a compile error: a larger C
// value indexes out of range, a smaller one indexes negatively.
var sizeMatch [1]struct{}

// Sizes defined by mpt_protocol.h.
var (
	_ = sizeMatch[C.kMPT_PRIVKEY_SIZE-mptsizes.PrivKeySize]
	_ = sizeMatch[C.kMPT_PUBKEY_SIZE-mptsizes.PubKeySize]
	_ = sizeMatch[C.kMPT_BLINDING_FACTOR_SIZE-mptsizes.BlindingFactorSize]
	_ = sizeMatch[C.kMPT_ELGAMAL_TOTAL_SIZE-mptsizes.CiphertextSize]
	_ = sizeMatch[C.kMPT_ACCOUNT_ID_SIZE-mptsizes.AccountIDSize]
	_ = sizeMatch[C.kMPT_ISSUANCE_ID_SIZE-mptsizes.IssuanceIDSize]
	_ = sizeMatch[C.kMPT_HALF_SHA_SIZE-mptsizes.HashOutputSize]
	_ = sizeMatch[C.kMPT_PEDERSEN_COMMIT_SIZE-mptsizes.CommitmentSize]
	_ = sizeMatch[C.kMPT_SCHNORR_PROOF_SIZE-mptsizes.SchnorrProofSize]
	_ = sizeMatch[C.kMPT_SINGLE_BULLETPROOF_SIZE-mptsizes.SingleBulletproofSize]
	_ = sizeMatch[C.kMPT_DOUBLE_BULLETPROOF_SIZE-mptsizes.DoubleBulletproofSize]
)

// Compact proof sizes defined by secp256k1_mpt.h.
var (
	_ = sizeMatch[C.SECP256K1_COMPACT_CLAWBACK_PROOF_SIZE-mptsizes.CompactClawbackProofSize]
	_ = sizeMatch[C.SECP256K1_COMPACT_CONVERTBACK_PROOF_SIZE-mptsizes.CompactConvertBackProofSize]
	_ = sizeMatch[C.SECP256K1_COMPACT_STANDARD_PROOF_SIZE-mptsizes.CompactSendProofSize]
)

// Proof blob sizes, pinned to the buffer lengths the C declarations spell out.
var (
	_ = sizeMatch[C.SECP256K1_COMPACT_CONVERTBACK_PROOF_SIZE+C.kMPT_SINGLE_BULLETPROOF_SIZE-mptsizes.ConvertBackProofSize]
	_ = sizeMatch[C.SECP256K1_COMPACT_STANDARD_PROOF_SIZE+C.kMPT_DOUBLE_BULLETPROOF_SIZE-mptsizes.SendProofSize]
)
