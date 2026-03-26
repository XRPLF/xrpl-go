package proof

import (
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

// decodeAddress decodes a classic XRPL address to a 20-byte account ID.
func decodeAddress(address string) ([mptcrypto.AccountIDSize]byte, error) {
	var id [mptcrypto.AccountIDSize]byte
	_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(address)
	if err != nil {
		return id, fmt.Errorf("%w: %w", ErrInvalidAddress, err)
	}
	copy(id[:], accountID)
	return id, nil
}

// decodeIssuanceID decodes a 48-char hex issuance ID to a 24-byte array.
func decodeIssuanceID(issHex string) ([mptcrypto.IssuanceIDSize]byte, error) {
	var id [mptcrypto.IssuanceIDSize]byte
	b, err := hexutil.DecodeFixedHex(issHex, mptcrypto.IssuanceIDSize)
	if err != nil {
		return id, fmt.Errorf("%w: %w", ErrInvalidIssuanceIDLength, err)
	}
	copy(id[:], b)
	return id, nil
}

// decodeParticipant converts a HexParticipant to a mptcrypto.Participant.
func decodeParticipant(hp HexParticipant) (mptcrypto.Participant, error) {
	var p mptcrypto.Participant
	pubBytes, err := hexutil.DecodeFixedHex(hp.PubKeyHex, mptcrypto.PubKeySize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidPubKeyLength, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(hp.CiphertextHex, mptcrypto.CiphertextSize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidCiphertextLength, err)
	}
	copy(p.PubKey[:], pubBytes)
	copy(p.Ciphertext[:], ctBytes)
	return p, nil
}

// decodeParticipants converts a slice of HexParticipant to mptcrypto.Participant.
func decodeParticipants(hps []HexParticipant) ([]mptcrypto.Participant, error) {
	parts := make([]mptcrypto.Participant, len(hps))
	for i, hp := range hps {
		p, err := decodeParticipant(hp)
		if err != nil {
			return nil, err
		}
		parts[i] = p
	}
	return parts, nil
}

// decodeProofParams converts a HexProofParams to a mptcrypto.PedersenProofParams.
func decodeProofParams(hp HexProofParams) (mptcrypto.PedersenProofParams, error) {
	var p mptcrypto.PedersenProofParams
	commitBytes, err := hexutil.DecodeFixedHex(hp.CommitmentHex, mptcrypto.CommitmentSize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidCommitmentLength, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(hp.CiphertextHex, mptcrypto.CiphertextSize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidCiphertextLength, err)
	}
	bfBytes, err := hexutil.DecodeFixedHex(hp.BlindingFactorHex, mptcrypto.BlindingFactorSize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidBlindingFactor, err)
	}
	copy(p.Commitment[:], commitBytes)
	p.Amount = hp.Amount
	copy(p.Ciphertext[:], ctBytes)
	copy(p.BlindingFactor[:], bfBytes)
	return p, nil
}
