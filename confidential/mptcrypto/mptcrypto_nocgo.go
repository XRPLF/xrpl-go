//go:build !cgo || js || wasip1 || tinygo || gofuzz || !(linux || darwin) || !(amd64 || arm64)

package mptcrypto

import "errors"

// ErrCgoRequired is returned by all crypto functions when built without CGo.
var ErrCgoRequired = errors.New(
	"mptcrypto: CGo is required for confidential MPT operations; " +
		"rebuild with CGO_ENABLED=1 and vendored mpt-crypto libraries",
)

// region ElGamal

// GenerateKeypair creates a new secp256k1 ElGamal keypair.
// Returns a 32-byte private key and a 33-byte compressed public key.
func GenerateKeypair() (privkey PrivateKey, pubkey PublicKey, err error) {
	return privkey, pubkey, ErrCgoRequired
}

// GenerateBlindingFactor returns a random 32-byte scalar suitable for ElGamal encryption.
func GenerateBlindingFactor() (bf BlindingFactor, err error) {
	return bf, ErrCgoRequired
}

// EncryptAmount encrypts a uint64 amount under a compressed public key using a blinding factor.
// Returns a 66-byte ciphertext (two compressed EC points: C1 || C2).
func EncryptAmount(amount uint64, pubkey PublicKey, bf BlindingFactor) (ct Ciphertext, err error) {
	return ct, ErrCgoRequired
}

// DecryptAmount decrypts a 66-byte ElGamal ciphertext using a private key.
// It searches the inclusive [rangeLow, rangeHigh] interval with linear cost.
func DecryptAmount(ciphertext Ciphertext, privateKey PrivateKey, rangeLow, rangeHigh uint64) (uint64, error) {
	if err := validateAmountRange(rangeLow, rangeHigh); err != nil {
		return 0, err
	}
	return 0, ErrCgoRequired
}

// endregion

// region Context hashes

// ConvertContextHash computes the context hash for a ConfidentialMPTConvert transaction.
func ConvertContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq uint32) (hash ContextHash, err error) {
	return hash, ErrCgoRequired
}

// ConvertBackContextHash computes the context hash for a ConfidentialMPTConvertBack transaction.
func ConvertBackContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq, ver uint32) (hash ContextHash, err error) {
	return hash, ErrCgoRequired
}

// SendContextHash computes the context hash for a ConfidentialMPTSend transaction.
func SendContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq uint32, dest [AccountIDSize]byte, ver uint32) (hash ContextHash, err error) {
	return hash, ErrCgoRequired
}

// ClawbackContextHash computes the context hash for a ConfidentialMPTClawback transaction.
func ClawbackContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq uint32, holder [AccountIDSize]byte) (hash ContextHash, err error) {
	return hash, ErrCgoRequired
}

// endregion

// region Pedersen commitment

// PedersenCommitment computes a Pedersen commitment for the given amount and blinding factor.
func PedersenCommitment(amount uint64, bf BlindingFactor) (commitment Commitment, err error) {
	return commitment, ErrCgoRequired
}

// endregion

// region Proof generation

// GenerateConvertProof generates a Schnorr proof of knowledge for a ConfidentialMPTConvert transaction.
func GenerateConvertProof(pubkey PublicKey, privkey PrivateKey, ctxHash ContextHash) (proof [SchnorrProofSize]byte, err error) {
	return proof, ErrCgoRequired
}

// GenerateConvertBackProof generates a compact AND-composed sigma proof over the balance
// witness, followed by a single Bulletproof range proof over the remainder commitment,
// for a ConfidentialMPTConvertBack transaction.
func GenerateConvertBackProof(privkey PrivateKey, pubkey PublicKey, ctxHash ContextHash, amount uint64, params PedersenProofParams) (proof [ConvertBackProofSize]byte, err error) {
	return proof, ErrCgoRequired
}

// GenerateClawbackProof generates an equality proof for a ConfidentialMPTClawback transaction.
func GenerateClawbackProof(privkey PrivateKey, pubkey PublicKey, ctxHash ContextHash, amount uint64, ciphertext Ciphertext) (proof [CompactClawbackProofSize]byte, err error) {
	return proof, ErrCgoRequired
}

// GenerateSendProof generates a compact AND-composed sigma proof + aggregated Bulletproof range proof
// for a ConfidentialMPTSend transaction.
func GenerateSendProof(privkey PrivateKey, pubkey PublicKey, amount uint64, participants []Participant, txBF BlindingFactor, ctxHash ContextHash, amountCommitment Commitment, balanceParams PedersenProofParams) ([]byte, error) {
	return nil, ErrCgoRequired
}

// endregion

// region Proof verification (top-level)

// VerifyConvertProof verifies a Schnorr proof for a ConfidentialMPTConvert transaction.
func VerifyConvertProof(proof [SchnorrProofSize]byte, pubkey PublicKey, ctxHash ContextHash) error {
	return ErrCgoRequired
}

// VerifyConvertBackProof verifies a linkage + range proof for a ConfidentialMPTConvertBack transaction.
// balanceCommit must be the original balance commitment, not the remainder after subtraction;
// the C library internally subtracts the transparent amount before checking the range proof.
func VerifyConvertBackProof(proof [ConvertBackProofSize]byte, pubkey PublicKey, ciphertext Ciphertext, balanceCommit Commitment, amount uint64, ctxHash ContextHash) error {
	return ErrCgoRequired
}

// VerifySendProof verifies the full proof for a ConfidentialMPTSend transaction.
func VerifySendProof(proof []byte, participants []Participant, senderCt Ciphertext, amountCommit, balanceCommit Commitment, ctxHash ContextHash) error {
	return ErrCgoRequired
}

// VerifyClawbackProof verifies an equality proof for a ConfidentialMPTClawback transaction.
func VerifyClawbackProof(proof [CompactClawbackProofSize]byte, amount uint64, pubkey PublicKey, ciphertext Ciphertext, ctxHash ContextHash) error {
	return ErrCgoRequired
}

// endregion

// region Internal component verifiers

// VerifyRevealedAmount verifies that a revealed amount and blinding factor are consistent
// with the participants' ciphertexts.
func VerifyRevealedAmount(amount uint64, bf BlindingFactor, holder, issuer Participant, auditor *Participant) error {
	return ErrCgoRequired
}

// VerifySendRangeProof verifies that the transfer amount and remaining balance are within [0, 2^64-1].
func VerifySendRangeProof(proof [DoubleBulletproofSize]byte, amountCommit, balanceCommitment Commitment, ctxHash ContextHash) error {
	return ErrCgoRequired
}

// endregion

// region Utilities

// ComputeConvertBackRemainder subtracts a transparent amount from a hidden Pedersen commitment.
func ComputeConvertBackRemainder(commitmentIn Commitment, amount uint64) (commitmentOut Commitment, err error) {
	return commitmentOut, ErrCgoRequired
}

// endregion
