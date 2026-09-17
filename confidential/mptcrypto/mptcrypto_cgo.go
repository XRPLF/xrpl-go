//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package mptcrypto

/*
#cgo CFLAGS: -I${SRCDIR}/../deps/include -I${SRCDIR}/../deps/include/utility
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../deps/libs/linux-amd64 -lmpt-crypto -lsecp256k1 -lcrypto -lstdc++ -lz -lm -ldl -lpthread
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../deps/libs/linux-arm64 -lmpt-crypto -lsecp256k1 -lcrypto -lstdc++ -lz -lm -ldl -lpthread
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/../deps/libs/darwin-arm64 -lmpt-crypto -lsecp256k1 -lcrypto -lc++ -lz -lm
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/../deps/libs/darwin-amd64 -lmpt-crypto -lsecp256k1 -lcrypto -lc++ -lz -lm

#include "mpt_utility.h"
*/
import "C"

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
)

func uint8Ptr(p *byte) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(p))
}

// region C struct helpers
func toAccountID(id [mptsizes.AccountIDSize]byte) C.account_id {
	var c C.account_id
	for i, b := range id {
		c.bytes[i] = C.uint8_t(b)
	}
	return c
}

func toIssuanceID(id [mptsizes.IssuanceIDSize]byte) C.mpt_issuance_id {
	var c C.mpt_issuance_id
	for i, b := range id {
		c.bytes[i] = C.uint8_t(b)
	}
	return c
}

func toParticipant(p Participant) C.mpt_confidential_participant {
	var c C.mpt_confidential_participant
	for i, b := range p.PubKey {
		c.pubkey[i] = C.uint8_t(b)
	}
	for i, b := range p.Ciphertext {
		c.ciphertext[i] = C.uint8_t(b)
	}
	return c
}

func toProofParams(p PedersenProofParams) C.mpt_pedersen_proof_params {
	var c C.mpt_pedersen_proof_params
	for i, b := range p.Commitment {
		c.pedersen_commitment[i] = C.uint8_t(b)
	}
	c.amount = C.uint64_t(p.Amount)
	for i, b := range p.Ciphertext {
		c.ciphertext[i] = C.uint8_t(b)
	}
	for i, b := range p.BlindingFactor {
		c.blinding_factor[i] = C.uint8_t(b)
	}
	return c
}

// endregion

// region ElGamal

// GenerateKeypair creates a new secp256k1 ElGamal keypair.
// Returns a 32-byte private key and a 33-byte compressed public key.
func GenerateKeypair() (privkey PrivateKey, pubkey PublicKey, err error) {
	ret := C.mpt_generate_keypair(
		uint8Ptr(&privkey[0]),
		uint8Ptr(&pubkey[0]),
	)
	if ret != 0 {
		return privkey, pubkey, fmt.Errorf("mpt_generate_keypair failed with code %d", ret)
	}
	return
}

// GenerateBlindingFactor returns a random 32-byte scalar suitable for ElGamal encryption.
func GenerateBlindingFactor() (bf BlindingFactor, err error) {
	ret := C.mpt_generate_blinding_factor(
		uint8Ptr(&bf[0]),
	)
	if ret != 0 {
		return bf, fmt.Errorf("mpt_generate_blinding_factor failed with code %d", ret)
	}
	return
}

// EncryptAmount encrypts a uint64 amount under a compressed public key using a blinding factor.
// Returns a 66-byte ciphertext (two compressed EC points: C1 || C2).
func EncryptAmount(amount uint64, pubkey PublicKey, bf BlindingFactor) (ct Ciphertext, err error) {
	ret := C.mpt_encrypt_amount(
		C.uint64_t(amount),
		uint8Ptr(&pubkey[0]),
		uint8Ptr(&bf[0]),
		uint8Ptr(&ct[0]),
	)
	if ret != 0 {
		return ct, fmt.Errorf("mpt_encrypt_amount failed with code %d", ret)
	}
	return
}

// DecryptAmount decrypts a 66-byte ElGamal ciphertext using a private key.
// It searches the inclusive [rangeLow, rangeHigh] interval with linear cost.
func DecryptAmount(ciphertext Ciphertext, privateKey PrivateKey, rangeLow, rangeHigh uint64) (uint64, error) {
	if err := validateAmountRange(rangeLow, rangeHigh); err != nil {
		return 0, err
	}

	var amount C.uint64_t
	ret := C.mpt_decrypt_amount(
		uint8Ptr(&ciphertext[0]),
		uint8Ptr(&privateKey[0]),
		&amount,
		C.uint64_t(rangeLow),
		C.uint64_t(rangeHigh),
	)
	if ret != 0 {
		return 0, fmt.Errorf("mpt_decrypt_amount failed with code %d", ret)
	}
	return uint64(amount), nil
}

func validateAmountRange(rangeLow, rangeHigh uint64) error {
	if rangeLow > rangeHigh {
		return fmt.Errorf("%w: low %d exceeds high %d", ErrInvalidAmountRange, rangeLow, rangeHigh)
	}
	if rangeHigh == math.MaxUint64 {
		return fmt.Errorf("%w: high must be less than %d", ErrInvalidAmountRange, uint64(math.MaxUint64))
	}
	return nil
}

// endregion

// region Context hashes

// ConvertContextHash computes the context hash for a ConfidentialMPTConvert transaction.
func ConvertContextHash(account [mptsizes.AccountIDSize]byte, iss [mptsizes.IssuanceIDSize]byte, seq uint32) (hash ContextHash, err error) {
	ret := C.mpt_get_convert_context_hash(
		toAccountID(account),
		toIssuanceID(iss),
		C.uint32_t(seq),
		uint8Ptr(&hash[0]),
	)
	if ret != 0 {
		return hash, fmt.Errorf("mpt_get_convert_context_hash failed with code %d", ret)
	}
	return
}

// ConvertBackContextHash computes the context hash for a ConfidentialMPTConvertBack transaction.
func ConvertBackContextHash(account [mptsizes.AccountIDSize]byte, iss [mptsizes.IssuanceIDSize]byte, seq, ver uint32) (hash ContextHash, err error) {
	ret := C.mpt_get_convert_back_context_hash(
		toAccountID(account),
		toIssuanceID(iss),
		C.uint32_t(seq),
		C.uint32_t(ver),
		uint8Ptr(&hash[0]),
	)
	if ret != 0 {
		return hash, fmt.Errorf("mpt_get_convert_back_context_hash failed with code %d", ret)
	}
	return
}

// SendContextHash computes the context hash for a ConfidentialMPTSend transaction.
func SendContextHash(account [mptsizes.AccountIDSize]byte, iss [mptsizes.IssuanceIDSize]byte, seq uint32, dest [mptsizes.AccountIDSize]byte, ver uint32) (hash ContextHash, err error) {
	ret := C.mpt_get_send_context_hash(
		toAccountID(account),
		toIssuanceID(iss),
		C.uint32_t(seq),
		toAccountID(dest),
		C.uint32_t(ver),
		uint8Ptr(&hash[0]),
	)
	if ret != 0 {
		return hash, fmt.Errorf("mpt_get_send_context_hash failed with code %d", ret)
	}
	return
}

// ClawbackContextHash computes the context hash for a ConfidentialMPTClawback transaction.
func ClawbackContextHash(account [mptsizes.AccountIDSize]byte, iss [mptsizes.IssuanceIDSize]byte, seq uint32, holder [mptsizes.AccountIDSize]byte) (hash ContextHash, err error) {
	ret := C.mpt_get_clawback_context_hash(
		toAccountID(account),
		toIssuanceID(iss),
		C.uint32_t(seq),
		toAccountID(holder),
		uint8Ptr(&hash[0]),
	)
	if ret != 0 {
		return hash, fmt.Errorf("mpt_get_clawback_context_hash failed with code %d", ret)
	}
	return
}

// endregion

// region Pedersen commitment

// PedersenCommitment computes a Pedersen commitment for the given amount and blinding factor.
func PedersenCommitment(amount uint64, bf BlindingFactor) (commitment Commitment, err error) {
	ret := C.mpt_get_pedersen_commitment(
		C.uint64_t(amount),
		uint8Ptr(&bf[0]),
		uint8Ptr(&commitment[0]),
	)
	if ret != 0 {
		return commitment, fmt.Errorf("mpt_get_pedersen_commitment failed with code %d", ret)
	}
	return
}

// endregion

// region Proof generation

// GenerateConvertProof generates a Schnorr proof of knowledge for a ConfidentialMPTConvert transaction.
func GenerateConvertProof(pubkey PublicKey, privkey PrivateKey, ctxHash ContextHash) (proof [mptsizes.SchnorrProofSize]byte, err error) {
	ret := C.mpt_get_convert_proof(
		uint8Ptr(&pubkey[0]),
		uint8Ptr(&privkey[0]),
		uint8Ptr(&ctxHash[0]),
		uint8Ptr(&proof[0]),
	)
	if ret != 0 {
		return proof, fmt.Errorf("mpt_get_convert_proof failed with code %d", ret)
	}
	return
}

// GenerateConvertBackProof generates a compact AND-composed sigma proof over the balance
// witness, followed by a single Bulletproof range proof over the remainder commitment,
// for a ConfidentialMPTConvertBack transaction.
func GenerateConvertBackProof(privkey PrivateKey, pubkey PublicKey, ctxHash ContextHash, amount uint64, params PedersenProofParams) (proof [mptsizes.ConvertBackProofSize]byte, err error) {
	cParams := toProofParams(params)
	ret := C.mpt_get_convert_back_proof(
		uint8Ptr(&privkey[0]),
		uint8Ptr(&pubkey[0]),
		uint8Ptr(&ctxHash[0]),
		C.uint64_t(amount),
		&cParams,
		uint8Ptr(&proof[0]),
	)
	if ret != 0 {
		return proof, fmt.Errorf("mpt_get_convert_back_proof failed with code %d", ret)
	}
	return
}

// GenerateClawbackProof generates an equality proof for a ConfidentialMPTClawback transaction.
func GenerateClawbackProof(privkey PrivateKey, pubkey PublicKey, ctxHash ContextHash, amount uint64, ciphertext Ciphertext) (proof [mptsizes.CompactClawbackProofSize]byte, err error) {
	ret := C.mpt_get_clawback_proof(
		uint8Ptr(&privkey[0]),
		uint8Ptr(&pubkey[0]),
		uint8Ptr(&ctxHash[0]),
		C.uint64_t(amount),
		uint8Ptr(&ciphertext[0]),
		uint8Ptr(&proof[0]),
	)
	if ret != 0 {
		return proof, fmt.Errorf("mpt_get_clawback_proof failed with code %d", ret)
	}
	return
}

// GenerateSendProof generates a compact AND-composed sigma proof + aggregated Bulletproof range proof
// for a ConfidentialMPTSend transaction.
func GenerateSendProof(privkey PrivateKey, pubkey PublicKey, amount uint64, participants []Participant, txBF BlindingFactor, ctxHash ContextHash, amountCommitment Commitment, balanceParams PedersenProofParams) ([]byte, error) {
	n := len(participants)
	if n == 0 {
		return nil, fmt.Errorf("mptcrypto: at least one participant is required")
	}
	if n > MaxParticipants {
		return nil, fmt.Errorf("mptcrypto: too many participants: %d (max %d)", n, MaxParticipants)
	}
	proof := make([]byte, mptsizes.SendProofSize)
	outLen := C.size_t(mptsizes.SendProofSize)

	cParts := make([]C.mpt_confidential_participant, n)
	for i, p := range participants {
		cParts[i] = toParticipant(p)
	}

	cBalance := toProofParams(balanceParams)

	ret := C.mpt_get_confidential_send_proof(
		uint8Ptr(&privkey[0]),
		uint8Ptr(&pubkey[0]),
		C.uint64_t(amount),
		&cParts[0],
		C.size_t(n),
		uint8Ptr(&txBF[0]),
		uint8Ptr(&ctxHash[0]),
		uint8Ptr(&amountCommitment[0]),
		&cBalance,
		uint8Ptr(&proof[0]),
		&outLen,
	)
	if ret != 0 {
		return nil, fmt.Errorf("mpt_get_confidential_send_proof failed with code %d", ret)
	}
	return proof[:outLen], nil
}

// endregion

// region Proof verification (top-level)

// VerifyConvertProof verifies a Schnorr proof for a ConfidentialMPTConvert transaction.
func VerifyConvertProof(proof [mptsizes.SchnorrProofSize]byte, pubkey PublicKey, ctxHash ContextHash) error {
	ret := C.mpt_verify_convert_proof(
		uint8Ptr(&proof[0]),
		uint8Ptr(&pubkey[0]),
		uint8Ptr(&ctxHash[0]),
	)
	if ret != 0 {
		return fmt.Errorf("mpt_verify_convert_proof failed with code %d", ret)
	}
	return nil
}

// VerifyConvertBackProof verifies a linkage + range proof for a ConfidentialMPTConvertBack transaction.
// balanceCommit must be the original balance commitment, not the remainder after subtraction,
// the C library internally subtracts the transparent amount before checking the range proof.
func VerifyConvertBackProof(proof [mptsizes.ConvertBackProofSize]byte, pubkey PublicKey, ciphertext Ciphertext, balanceCommit Commitment, amount uint64, ctxHash ContextHash) error {
	ret := C.mpt_verify_convert_back_proof(
		uint8Ptr(&proof[0]),
		uint8Ptr(&pubkey[0]),
		uint8Ptr(&ciphertext[0]),
		uint8Ptr(&balanceCommit[0]),
		C.uint64_t(amount),
		uint8Ptr(&ctxHash[0]),
	)
	if ret != 0 {
		return fmt.Errorf("mpt_verify_convert_back_proof failed with code %d", ret)
	}
	return nil
}

// VerifySendProof verifies the full proof for a ConfidentialMPTSend transaction.
func VerifySendProof(proof []byte, participants []Participant, senderCt Ciphertext, amountCommit, balanceCommit Commitment, ctxHash ContextHash) error {
	if len(proof) != mptsizes.SendProofSize {
		return fmt.Errorf("mptcrypto: proof must be %d bytes, got %d", mptsizes.SendProofSize, len(proof))
	}
	if len(participants) == 0 {
		return fmt.Errorf("mptcrypto: at least one participant is required")
	}
	if len(participants) > MaxParticipants {
		return fmt.Errorf("mptcrypto: too many participants: %d (max %d)", len(participants), MaxParticipants)
	}
	cParts := make([]C.mpt_confidential_participant, len(participants))
	for i, p := range participants {
		cParts[i] = toParticipant(p)
	}
	ret := C.mpt_verify_send_proof(
		uint8Ptr(&proof[0]),
		&cParts[0],
		C.uint8_t(len(participants)),
		uint8Ptr(&senderCt[0]),
		uint8Ptr(&amountCommit[0]),
		uint8Ptr(&balanceCommit[0]),
		uint8Ptr(&ctxHash[0]),
	)
	if ret != 0 {
		return fmt.Errorf("mpt_verify_send_proof failed with code %d", ret)
	}
	return nil
}

// VerifyClawbackProof verifies an equality proof for a ConfidentialMPTClawback transaction.
func VerifyClawbackProof(proof [mptsizes.CompactClawbackProofSize]byte, amount uint64, pubkey PublicKey, ciphertext Ciphertext, ctxHash ContextHash) error {
	ret := C.mpt_verify_clawback_proof(
		uint8Ptr(&proof[0]),
		C.uint64_t(amount),
		uint8Ptr(&pubkey[0]),
		uint8Ptr(&ciphertext[0]),
		uint8Ptr(&ctxHash[0]),
	)
	if ret != 0 {
		return fmt.Errorf("mpt_verify_clawback_proof failed with code %d", ret)
	}
	return nil
}

// endregion

// region Internal component verifiers

// VerifyRevealedAmount verifies that a revealed amount and blinding factor are consistent
// with the participants' ciphertexts. auditor may be nil if no auditor is present.
func VerifyRevealedAmount(amount uint64, bf BlindingFactor, holder, issuer Participant, auditor *Participant) error {
	cHolder := toParticipant(holder)
	cIssuer := toParticipant(issuer)
	var cAuditor *C.mpt_confidential_participant
	if auditor != nil {
		a := toParticipant(*auditor)
		cAuditor = &a
	}
	ret := C.mpt_verify_revealed_amount(
		C.uint64_t(amount),
		uint8Ptr(&bf[0]),
		&cHolder,
		&cIssuer,
		cAuditor,
	)
	if ret != 0 {
		return fmt.Errorf("mpt_verify_revealed_amount failed with code %d", ret)
	}
	return nil
}

// VerifySendRangeProof verifies that the transfer amount and remaining balance are within [0, 2^64-1].
func VerifySendRangeProof(proof [mptsizes.DoubleBulletproofSize]byte, amountCommit, balanceCommitment Commitment, ctxHash ContextHash) error {
	ret := C.mpt_verify_send_range_proof(
		uint8Ptr(&proof[0]),
		uint8Ptr(&amountCommit[0]),
		uint8Ptr(&balanceCommitment[0]),
		uint8Ptr(&ctxHash[0]),
	)
	if ret != 0 {
		return fmt.Errorf("mpt_verify_send_range_proof failed with code %d", ret)
	}
	return nil
}

// endregion

// region Utilities

// ComputeConvertBackRemainder subtracts a transparent amount from a hidden Pedersen commitment.
func ComputeConvertBackRemainder(commitmentIn Commitment, amount uint64) (commitmentOut Commitment, err error) {
	ret := C.mpt_compute_convert_back_remainder(
		uint8Ptr(&commitmentIn[0]),
		C.uint64_t(amount),
		uint8Ptr(&commitmentOut[0]),
	)
	if ret != 0 {
		return commitmentOut, fmt.Errorf("mpt_compute_convert_back_remainder failed with code %d", ret)
	}
	return
}

// endregion

// region Homomorphic ciphertext arithmetic

// AddCiphertexts homomorphically adds two ElGamal ciphertexts encrypted under the same public key,
// returning an encryption of the sum of their plaintexts.
func AddCiphertexts(a, b Ciphertext) (sum Ciphertext, err error) {
	return combineCiphertexts(a, b, false)
}

// SubtractCiphertexts homomorphically subtracts b from a, both encrypted under the same public
// key, returning an encryption of the difference of their plaintexts.
func SubtractCiphertexts(a, b Ciphertext) (difference Ciphertext, err error) {
	return combineCiphertexts(a, b, true)
}

// combineCiphertexts parses both operands into the library's internal point form, applies the
// requested group operation, and serializes the result back to the wire form.
func combineCiphertexts(a, b Ciphertext, subtract bool) (result Ciphertext, err error) {
	var aC1, aC2, bC1, bC2, outC1, outC2 C.secp256k1_pubkey
	if !C.mpt_make_ec_pair(uint8Ptr(&a[0]), &aC1, &aC2) {
		return result, fmt.Errorf("%w: first operand", ErrInvalidCiphertext)
	}
	if !C.mpt_make_ec_pair(uint8Ptr(&b[0]), &bC1, &bC2) {
		return result, fmt.Errorf("%w: second operand", ErrInvalidCiphertext)
	}

	ctx := C.mpt_secp256k1_context()
	operation, ret := "secp256k1_elgamal_add", C.int(0)
	if subtract {
		operation = "secp256k1_elgamal_subtract"
		ret = C.secp256k1_elgamal_subtract(ctx, &outC1, &outC2, &aC1, &aC2, &bC1, &bC2)
	} else {
		ret = C.secp256k1_elgamal_add(ctx, &outC1, &outC2, &aC1, &aC2, &bC1, &bC2)
	}
	if ret != 1 {
		return result, fmt.Errorf("%s failed with code %d", operation, ret)
	}

	if !C.mpt_serialize_ec_pair(&outC1, &outC2, uint8Ptr(&result[0])) {
		return result, fmt.Errorf("%w: %s produced a point that cannot be serialized", ErrInvalidCiphertext, operation)
	}
	return result, nil
}

// endregion

// region Canonical encrypted zero

// CanonicalEncryptedZero returns the deterministic encryption of zero that xrpld writes when a
// confidential transactor initializes or resets a balance: a first-time convert's spending
// balance, a merge's inbox, and every balance a clawback clears. It is derived from the key the
// balance is encrypted under, the holder's AccountID, and the issuance, so the ciphertext the
// ledger stores can be reproduced client-side.
func CanonicalEncryptedZero(pubkey PublicKey, account [mptsizes.AccountIDSize]byte, iss [mptsizes.IssuanceIDSize]byte) (ct Ciphertext, err error) {
	ctx := C.mpt_secp256k1_context()

	var parsed, c1, c2 C.secp256k1_pubkey
	if C.secp256k1_ec_pubkey_parse(ctx, &parsed, (*C.uchar)(unsafe.Pointer(&pubkey[0])), C.size_t(len(pubkey))) != 1 {
		return ct, ErrInvalidPublicKey
	}
	ret := C.generate_canonical_encrypted_zero(
		ctx,
		&c1,
		&c2,
		&parsed,
		(*C.uchar)(unsafe.Pointer(&account[0])),
		(*C.uchar)(unsafe.Pointer(&iss[0])),
	)
	if ret != 1 {
		return ct, fmt.Errorf("generate_canonical_encrypted_zero failed with code %d", ret)
	}
	if !C.mpt_serialize_ec_pair(&c1, &c2, uint8Ptr(&ct[0])) {
		return ct, fmt.Errorf("%w: canonical encrypted zero cannot be serialized", ErrInvalidCiphertext)
	}
	return ct, nil
}

// endregion
