package proof

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// VerifyRevealedAmount verifies that a revealed amount and blinding factor are consistent
// with the participants' ciphertexts. auditor may be nil.
func VerifyRevealedAmount(amount uint64, bfHex string, holder, issuer Participant, auditor *Participant) error {
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

// VerifySendRangeProof verifies that the transfer amount and remaining balance are within [0, 2^64-1].
// balanceCommitHex must be the sender's original balance commitment, the C library derives the remainder internally.
func VerifySendRangeProof(proofHex, amountCommitHex, balanceCommitHex, ctxHashHex string) error {
	proofBytes, err := hexutil.DecodeFixedHex(proofHex, mptcrypto.DoubleBulletproofSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidProof, err)
	}
	amountCommitBytes, err := hexutil.DecodeFixedHex(amountCommitHex, mptcrypto.CommitmentSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitment, err)
	}
	balanceCommitBytes, err := hexutil.DecodeFixedHex(balanceCommitHex, mptcrypto.CommitmentSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommitment, err)
	}
	hashBytes, err := hexutil.DecodeFixedHex(ctxHashHex, mptcrypto.HashOutputSize)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidContextHash, err)
	}

	var proof [mptcrypto.DoubleBulletproofSize]byte
	var amountCommit [mptcrypto.CommitmentSize]byte
	var balanceCommit [mptcrypto.CommitmentSize]byte
	var hash [mptcrypto.HashOutputSize]byte
	copy(proof[:], proofBytes)
	copy(amountCommit[:], amountCommitBytes)
	copy(balanceCommit[:], balanceCommitBytes)
	copy(hash[:], hashBytes)

	if err := mptcrypto.VerifySendRangeProof(proof, amountCommit, balanceCommit, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrProofVerificationFailed, err)
	}
	return nil
}
