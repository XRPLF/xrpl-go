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
	"unsafe"
)

func uint8Ptr(p *byte) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(p))
}

// region C struct helpers
func toAccountID(id [AccountIDSize]byte) C.account_id {
	var c C.account_id
	for i, b := range id {
		c.bytes[i] = C.uint8_t(b)
	}
	return c
}

func toIssuanceID(id [IssuanceIDSize]byte) C.mpt_issuance_id {
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
func GenerateKeypair() (privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, err error) {
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
func GenerateBlindingFactor() (bf [BlindingFactorSize]byte, err error) {
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
func EncryptAmount(amount uint64, pubkey [PubKeySize]byte, bf [BlindingFactorSize]byte) (ct [CiphertextSize]byte, err error) {
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

// endregion

// region Context hashes

// ConvertContextHash computes the context hash for a ConfidentialMPTConvert transaction.
func ConvertContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq uint32) (hash [HashOutputSize]byte, err error) {
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
func ConvertBackContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq, ver uint32) (hash [HashOutputSize]byte, err error) {
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
func SendContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq uint32, dest [AccountIDSize]byte, ver uint32) (hash [HashOutputSize]byte, err error) {
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
func ClawbackContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq uint32, holder [AccountIDSize]byte) (hash [HashOutputSize]byte, err error) {
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
func PedersenCommitment(amount uint64, bf [BlindingFactorSize]byte) (commitment [CommitmentSize]byte, err error) {
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
func GenerateConvertProof(pubkey [PubKeySize]byte, privkey [PrivKeySize]byte, ctxHash [HashOutputSize]byte) (proof [SchnorrProofSize]byte, err error) {
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
func GenerateConvertBackProof(privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, ctxHash [HashOutputSize]byte, amount uint64, params PedersenProofParams) (proof [ConvertBackProofSize]byte, err error) {
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
func GenerateClawbackProof(privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, ctxHash [HashOutputSize]byte, amount uint64, ciphertext [CiphertextSize]byte) (proof [CompactClawbackProofSize]byte, err error) {
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
func GenerateSendProof(privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, amount uint64, participants []Participant, txBF [BlindingFactorSize]byte, ctxHash [HashOutputSize]byte, amountCommitment [CommitmentSize]byte, balanceParams PedersenProofParams) ([]byte, error) {
	n := len(participants)
	if n == 0 {
		return nil, fmt.Errorf("mptcrypto: at least one participant is required")
	}
	if n > MaxParticipants {
		return nil, fmt.Errorf("mptcrypto: too many participants: %d (max %d)", n, MaxParticipants)
	}
	proof := make([]byte, SendProofSize)
	outLen := C.size_t(SendProofSize)

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
func VerifyConvertProof(proof [SchnorrProofSize]byte, pubkey [PubKeySize]byte, ctxHash [HashOutputSize]byte) error {
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
func VerifyConvertBackProof(proof [ConvertBackProofSize]byte, pubkey [PubKeySize]byte, ciphertext [CiphertextSize]byte, balanceCommit [CommitmentSize]byte, amount uint64, ctxHash [HashOutputSize]byte) error {
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
func VerifySendProof(proof []byte, participants []Participant, senderCt [CiphertextSize]byte, amountCommit, balanceCommit [CommitmentSize]byte, ctxHash [HashOutputSize]byte) error {
	if len(proof) != SendProofSize {
		return fmt.Errorf("mptcrypto: proof must be %d bytes, got %d", SendProofSize, len(proof))
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
func VerifyClawbackProof(proof [CompactClawbackProofSize]byte, amount uint64, pubkey [PubKeySize]byte, ciphertext [CiphertextSize]byte, ctxHash [HashOutputSize]byte) error {
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
func VerifyRevealedAmount(amount uint64, bf [BlindingFactorSize]byte, holder, issuer Participant, auditor *Participant) error {
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
func VerifySendRangeProof(proof [DoubleBulletproofSize]byte, amountCommit, balanceCommitment [CommitmentSize]byte, ctxHash [HashOutputSize]byte) error {
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
func ComputeConvertBackRemainder(commitmentIn [CommitmentSize]byte, amount uint64) (commitmentOut [CommitmentSize]byte, err error) {
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
