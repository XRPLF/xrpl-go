//go:build !cgo

package mptcrypto

import "errors"

// ErrCgoRequired is returned by all crypto functions when built without CGo.
var ErrCgoRequired = errors.New(
	"mptcrypto: CGo is required for confidential MPT operations; " +
		"rebuild with CGO_ENABLED=1 and vendored mpt-crypto libraries",
)

//region ElGamal

// GenerateKeypair creates a new secp256k1 ElGamal keypair.
// Returns a 32-byte private key and a 33-byte compressed public key.
func GenerateKeypair() (privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, err error) {
	return privkey, pubkey, ErrCgoRequired
}

// GenerateBlindingFactor returns a random 32-byte scalar suitable for ElGamal encryption.
func GenerateBlindingFactor() (bf [BlindingFactorSize]byte, err error) {
	return bf, ErrCgoRequired
}

// EncryptAmount encrypts a uint64 amount under a compressed public key using a blinding factor.
// Returns a 66-byte ciphertext (two compressed EC points: C1 || C2).
func EncryptAmount(amount uint64, pubkey [PubKeySize]byte, bf [BlindingFactorSize]byte) (ct [CiphertextSize]byte, err error) {
	return ct, ErrCgoRequired
}

// DecryptAmount decrypts a 66-byte ElGamal ciphertext using a private key.
// Returns the plaintext uint64 amount.
func DecryptAmount(ciphertext [CiphertextSize]byte, privkey [PrivKeySize]byte) (uint64, error) {
	return 0, ErrCgoRequired
}

//endregion

//region Context hashes

// ConvertContextHash computes the context hash for a ConfidentialMPTConvert transaction.
func ConvertContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq uint32) (hash [HashOutputSize]byte, err error) {
	return hash, ErrCgoRequired
}

// ConvertBackContextHash computes the context hash for a ConfidentialMPTConvertBack transaction.
func ConvertBackContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq, ver uint32) (hash [HashOutputSize]byte, err error) {
	return hash, ErrCgoRequired
}

// SendContextHash computes the context hash for a ConfidentialMPTSend transaction.
func SendContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq uint32, dest [AccountIDSize]byte, ver uint32) (hash [HashOutputSize]byte, err error) {
	return hash, ErrCgoRequired
}

// ClawbackContextHash computes the context hash for a ConfidentialMPTClawback transaction.
func ClawbackContextHash(account [AccountIDSize]byte, iss [IssuanceIDSize]byte, seq uint32, holder [AccountIDSize]byte) (hash [HashOutputSize]byte, err error) {
	return hash, ErrCgoRequired
}

//endregion

//region Pedersen commitment

// PedersenCommitment computes a Pedersen commitment for the given amount and blinding factor.
func PedersenCommitment(amount uint64, bf [BlindingFactorSize]byte) (commitment [CommitmentSize]byte, err error) {
	return commitment, ErrCgoRequired
}

//endregion

//region Proof generation

// GenerateConvertProof generates a Schnorr proof of knowledge for a ConfidentialMPTConvert transaction.
func GenerateConvertProof(pubkey [PubKeySize]byte, privkey [PrivKeySize]byte, ctxHash [HashOutputSize]byte) (proof [SchnorrProofSize]byte, err error) {
	return proof, ErrCgoRequired
}

// GenerateConvertBackProof generates a linkage + range proof for a ConfidentialMPTConvertBack transaction.
func GenerateConvertBackProof(privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, ctxHash [HashOutputSize]byte, amount uint64, params PedersenProofParams) (proof [ConvertBackProofSize]byte, err error) {
	return proof, ErrCgoRequired
}

// GenerateClawbackProof generates an equality proof for a ConfidentialMPTClawback transaction.
func GenerateClawbackProof(privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, ctxHash [HashOutputSize]byte, amount uint64, ciphertext [CiphertextSize]byte) (proof [EqualityProofSize]byte, err error) {
	return proof, ErrCgoRequired
}

// GenerateSendProof generates a full proof for a ConfidentialMPTSend transaction.
func GenerateSendProof(privkey [PrivKeySize]byte, amount uint64, participants []Participant, txBF [BlindingFactorSize]byte, ctxHash [HashOutputSize]byte, amountParams, balanceParams PedersenProofParams) ([]byte, error) {
	return nil, ErrCgoRequired
}

// GenerateAmountLinkageProof generates a Pedersen linkage proof between an ElGamal ciphertext and a commitment.
func GenerateAmountLinkageProof(pubkey [PubKeySize]byte, bf [BlindingFactorSize]byte, ctxHash [HashOutputSize]byte, params PedersenProofParams) (proof [PedersenLinkSize]byte, err error) {
	return proof, ErrCgoRequired
}

// GenerateBalanceLinkageProof generates a Pedersen linkage proof for the sender's balance.
func GenerateBalanceLinkageProof(privkey [PrivKeySize]byte, pubkey [PubKeySize]byte, ctxHash [HashOutputSize]byte, params PedersenProofParams) (proof [PedersenLinkSize]byte, err error) {
	return proof, ErrCgoRequired
}

//endregion

//region Proof verification (top-level)

// VerifyConvertProof verifies a Schnorr proof for a ConfidentialMPTConvert transaction.
func VerifyConvertProof(proof [SchnorrProofSize]byte, pubkey [PubKeySize]byte, ctxHash [HashOutputSize]byte) error {
	return ErrCgoRequired
}

// VerifyConvertBackProof verifies a linkage + range proof for a ConfidentialMPTConvertBack transaction.
// balanceCommit must be the original balance commitment, not the remainder after subtraction;
// the C library internally subtracts the transparent amount before checking the range proof.
func VerifyConvertBackProof(proof [ConvertBackProofSize]byte, pubkey [PubKeySize]byte, ciphertext [CiphertextSize]byte, balanceCommit [CommitmentSize]byte, amount uint64, ctxHash [HashOutputSize]byte) error {
	return ErrCgoRequired
}

// VerifySendProof verifies the full proof for a ConfidentialMPTSend transaction.
func VerifySendProof(proof []byte, participants []Participant, senderCt [CiphertextSize]byte, amountCommit, balanceCommit [CommitmentSize]byte, ctxHash [HashOutputSize]byte) error {
	return ErrCgoRequired
}

// VerifyClawbackProof verifies an equality proof for a ConfidentialMPTClawback transaction.
func VerifyClawbackProof(proof [EqualityProofSize]byte, amount uint64, pubkey [PubKeySize]byte, ciphertext [CiphertextSize]byte, ctxHash [HashOutputSize]byte) error {
	return ErrCgoRequired
}

//endregion

//region Internal component verifiers

// VerifyRevealedAmount verifies that a revealed amount and blinding factor are consistent
// with the participants' ciphertexts.
func VerifyRevealedAmount(amount uint64, bf [BlindingFactorSize]byte, holder, issuer Participant, auditor *Participant) error {
	return ErrCgoRequired
}

// VerifyAmountLinkage verifies a Pedersen linkage proof for the transaction amount.
func VerifyAmountLinkage(proof [PedersenLinkSize]byte, ciphertext [CiphertextSize]byte, pubkey [PubKeySize]byte, commitment [CommitmentSize]byte, ctxHash [HashOutputSize]byte) error {
	return ErrCgoRequired
}

// VerifyBalanceLinkage verifies a Pedersen linkage proof for the sender's balance.
func VerifyBalanceLinkage(proof [PedersenLinkSize]byte, ciphertext [CiphertextSize]byte, pubkey [PubKeySize]byte, commitment [CommitmentSize]byte, ctxHash [HashOutputSize]byte) error {
	return ErrCgoRequired
}

// VerifyEqualityProof verifies that all participants' ciphertexts encrypt the same value.
func VerifyEqualityProof(proof []byte, participants []Participant, ctxHash [HashOutputSize]byte) error {
	return ErrCgoRequired
}

// VerifySendRangeProof verifies that the transfer amount and remaining balance are within [0, 2^64-1].
func VerifySendRangeProof(proof [DoubleBulletproofSize]byte, amountCommit, remainderCommit [CommitmentSize]byte, ctxHash [HashOutputSize]byte) error {
	return ErrCgoRequired
}

//endregion

//region Utilities

// GetSendProofSize returns the total proof size in bytes for a ConfidentialMPTSend with nRecipients.
// Without CGo this always returns 0; callers should rely on GenerateSendProof returning ErrCgoRequired.
func GetSendProofSize(nRecipients int) int {
	return 0
}

// ComputeConvertBackRemainder subtracts a transparent amount from a hidden Pedersen commitment.
func ComputeConvertBackRemainder(commitmentIn [CommitmentSize]byte, amount uint64) (commitmentOut [CommitmentSize]byte, err error) {
	return commitmentOut, ErrCgoRequired
}

//endregion
