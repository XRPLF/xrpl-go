package proof

import (
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
	"github.com/Peersyst/xrpl-go/pkg/mptsizes"
)

// decodeAddress decodes a classic address or X-address to a 20-byte account ID.
func decodeAddress(address string) ([mptsizes.AccountIDSize]byte, error) {
	var id [mptsizes.AccountIDSize]byte
	decoded, err := addresscodec.DecodeAddress(address)
	if err != nil {
		return id, fmt.Errorf("%w: %w", ErrInvalidAddress, err)
	}
	copy(id[:], decoded.AccountID[:])
	return id, nil
}

// decodeIssuanceID decodes a 48-char hex issuance ID to a 24-byte array.
func decodeIssuanceID(issHex string) ([mptsizes.IssuanceIDSize]byte, error) {
	var id [mptsizes.IssuanceIDSize]byte
	b, err := hexutil.DecodeFixedHex(issHex, mptsizes.IssuanceIDSize)
	if err != nil {
		return id, fmt.Errorf("%w: %w", ErrInvalidIssuanceID, err)
	}
	copy(id[:], b)
	return id, nil
}

// decodeParticipant converts a Participant to a mptcrypto.Participant.
func decodeParticipant(hp Participant) (mptcrypto.Participant, error) {
	var p mptcrypto.Participant
	pubBytes, err := hexutil.DecodeFixedHex(hp.PubKeyHex, mptsizes.PubKeySize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidPubKey, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(hp.CiphertextHex, mptsizes.CiphertextSize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}
	p.PubKey = mptcrypto.PublicKey(pubBytes)
	p.Ciphertext = mptcrypto.Ciphertext(ctBytes)
	return p, nil
}

// decodeProofParams converts a Params to a mptcrypto.PedersenProofParams.
func decodeProofParams(hp Params) (mptcrypto.PedersenProofParams, error) {
	var p mptcrypto.PedersenProofParams
	commitBytes, err := hexutil.DecodeFixedHex(hp.CommitmentHex, mptsizes.CommitmentSize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidCommitment, err)
	}
	ctBytes, err := hexutil.DecodeFixedHex(hp.CiphertextHex, mptsizes.CiphertextSize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}
	bfBytes, err := hexutil.DecodeFixedHex(hp.BlindingFactorHex, mptsizes.BlindingFactorSize)
	if err != nil {
		return p, fmt.Errorf("%w: %w", ErrInvalidBlindingFactor, err)
	}
	p.Commitment = mptcrypto.Commitment(commitBytes)
	p.Amount = hp.Amount
	p.Ciphertext = mptcrypto.Ciphertext(ctBytes)
	p.BlindingFactor = mptcrypto.BlindingFactor(bfBytes)
	return p, nil
}
