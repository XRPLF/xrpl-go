// Package proofs provides a hex-string API for ZK proof generation and verification.
// It wraps the byte-array functions in mptcrypto with hex encoding/decoding and
// classic XRPL address decoding for use with transaction fields.
package proofs

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
)

// ConvertContextHash computes the context hash for a ConfidentialMPTConvert transaction.
// account is a classic XRPL address, issuanceIDHex is 48 hex chars (24 bytes).
// Returns 64 hex chars (32 bytes).
func ConvertContextHash(account string, issuanceIDHex string, seq uint32) (string, error) {
	accID, err := decodeAddress(account)
	if err != nil {
		return "", err
	}
	issID, err := decodeIssuanceID(issuanceIDHex)
	if err != nil {
		return "", err
	}

	hash, err := mptcrypto.ConvertContextHash(accID, issID, seq)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrContextHashFailed, err)
	}
	return hex.EncodeToString(hash[:]), nil
}

// ConvertBackContextHash computes the context hash for a ConfidentialMPTConvertBack transaction.
// account is a classic XRPL address, issuanceIDHex is 48 hex chars (24 bytes).
// Returns 64 hex chars (32 bytes).
func ConvertBackContextHash(account string, issuanceIDHex string, seq, version uint32) (string, error) {
	accID, err := decodeAddress(account)
	if err != nil {
		return "", err
	}
	issID, err := decodeIssuanceID(issuanceIDHex)
	if err != nil {
		return "", err
	}

	hash, err := mptcrypto.ConvertBackContextHash(accID, issID, seq, version)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrContextHashFailed, err)
	}
	return hex.EncodeToString(hash[:]), nil
}

// SendContextHash computes the context hash for a ConfidentialMPTSend transaction.
// account and dest are classic XRPL addresses, issuanceIDHex is 48 hex chars (24 bytes).
// Returns 64 hex chars (32 bytes).
func SendContextHash(account string, issuanceIDHex string, seq uint32, dest string, version uint32) (string, error) {
	accID, err := decodeAddress(account)
	if err != nil {
		return "", err
	}
	issID, err := decodeIssuanceID(issuanceIDHex)
	if err != nil {
		return "", err
	}
	destID, err := decodeAddress(dest)
	if err != nil {
		return "", err
	}

	hash, err := mptcrypto.SendContextHash(accID, issID, seq, destID, version)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrContextHashFailed, err)
	}
	return hex.EncodeToString(hash[:]), nil
}

// ClawbackContextHash computes the context hash for a ConfidentialMPTClawback transaction.
// account and holder are classic XRPL addresses, issuanceIDHex is 48 hex chars (24 bytes).
// Returns 64 hex chars (32 bytes).
func ClawbackContextHash(account string, issuanceIDHex string, seq uint32, holder string) (string, error) {
	accID, err := decodeAddress(account)
	if err != nil {
		return "", err
	}
	issID, err := decodeIssuanceID(issuanceIDHex)
	if err != nil {
		return "", err
	}
	holderID, err := decodeAddress(holder)
	if err != nil {
		return "", err
	}

	hash, err := mptcrypto.ClawbackContextHash(accID, issID, seq, holderID)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrContextHashFailed, err)
	}
	return hex.EncodeToString(hash[:]), nil
}
