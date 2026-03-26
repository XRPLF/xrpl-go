package proof

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// VerifyRevealedAmount verifies that a revealed amount and blinding factor are consistent
// with the participants' ciphertexts. auditor may be nil.
func VerifyRevealedAmount(amount uint64, bfHex string, holder, issuer HexParticipant, auditor *HexParticipant) error {
	bfBytes, err := hexutil.DecodeFixedHex(bfHex, mptcrypto.BlindingFactorSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidBlindingFactor, err)
	}
	holderP, err := decodeParticipant(holder)
	if err != nil {
		return err
	}
	issuerP, err := decodeParticipant(issuer)
	if err != nil {
		return err
	}

	var bf [mptcrypto.BlindingFactorSize]byte
	copy(bf[:], bfBytes)

	var auditorP *mptcrypto.Participant
	if auditor != nil {
		a, err := decodeParticipant(*auditor)
		if err != nil {
			return err
		}
		auditorP = &a
	}

	if err := mptcrypto.VerifyRevealedAmount(amount, bf, holderP, issuerP, auditorP); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}

// linkageVerifyFn is the signature shared by mptcrypto.VerifyAmountLinkage and VerifyBalanceLinkage.
type linkageVerifyFn func([mptcrypto.PedersenLinkSize]byte, [mptcrypto.CiphertextSize]byte, [mptcrypto.PubKeySize]byte, [mptcrypto.CommitmentSize]byte, [mptcrypto.HashOutputSize]byte) error

func verifyLinkage(proofHex, ciphertextHex, pubkeyHex, commitmentHex, ctxHashHex string, fn linkageVerifyFn) error {
	proofBytes, err := hexutil.DecodeFixedHex(proofHex, mptcrypto.PedersenLinkSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidProofLength, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(ciphertextHex, mptcrypto.CiphertextSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCiphertextLength, err)
	}
	pubBytes, err := hexutil.DecodeFixedHex(pubkeyHex, mptcrypto.PubKeySize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPubKeyLength, err)
	}
	commitBytes, err := hexutil.DecodeFixedHex(commitmentHex, mptcrypto.CommitmentSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitmentLength, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var proof [mptcrypto.PedersenLinkSize]byte
	var ct [mptcrypto.CiphertextSize]byte
	var pub [mptcrypto.PubKeySize]byte
	var commit [mptcrypto.CommitmentSize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(proof[:], proofBytes)
	copy(ct[:], ctBytes)
	copy(pub[:], pubBytes)
	copy(commit[:], commitBytes)
	copy(hash[:], hashBytes)

	if err := fn(proof, ct, pub, commit, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}

// VerifyAmountLinkage verifies a Pedersen linkage proof for the transaction amount.
func VerifyAmountLinkage(proofHex, ciphertextHex, pubkeyHex, commitmentHex, ctxHashHex string) error {
	return verifyLinkage(proofHex, ciphertextHex, pubkeyHex, commitmentHex, ctxHashHex, mptcrypto.VerifyAmountLinkage)
}

// VerifyBalanceLinkage verifies a Pedersen linkage proof for the sender's balance.
func VerifyBalanceLinkage(proofHex, ciphertextHex, pubkeyHex, commitmentHex, ctxHashHex string) error {
	return verifyLinkage(proofHex, ciphertextHex, pubkeyHex, commitmentHex, ctxHashHex, mptcrypto.VerifyBalanceLinkage)
}

// VerifyEqualityProof verifies that all participants' ciphertexts encrypt the same value.
func VerifyEqualityProof(proofHex string, participants []HexParticipant, ctxHashHex string) error {
	proofBytes, err := hex.DecodeString(proofHex)
	if err != nil {
		return fmt.Errorf("%w: invalid hex: %w", ErrInvalidProofLength, err)
	}
	parts, err := decodeParticipants(participants)
	if err != nil {
		return err
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var hash [mptcrypto.HashOutputSize]byte
	copy(hash[:], hashBytes)

	if err := mptcrypto.VerifyEqualityProof(proofBytes, parts, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}

// VerifySendRangeProof verifies that the transfer amount and remaining balance are within [0, 2^64-1].
func VerifySendRangeProof(proofHex, amountCommitHex, remainderCommitHex, ctxHashHex string) error {
	proofBytes, err := hexutil.DecodeFixedHex(proofHex, mptcrypto.DoubleBulletproofSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidProofLength, err)
	}
	amountCommitBytes, err := hexutil.DecodeFixedHex(amountCommitHex, mptcrypto.CommitmentSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitmentLength, err)
	}
	remainderCommitBytes, err := hexutil.DecodeFixedHex(remainderCommitHex, mptcrypto.CommitmentSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitmentLength, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var proof [mptcrypto.DoubleBulletproofSize]byte
	var amountCommit [mptcrypto.CommitmentSize]byte
	var remainderCommit [mptcrypto.CommitmentSize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(proof[:], proofBytes)
	copy(amountCommit[:], amountCommitBytes)
	copy(remainderCommit[:], remainderCommitBytes)
	copy(hash[:], hashBytes)

	if err := mptcrypto.VerifySendRangeProof(proof, amountCommit, remainderCommit, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}
