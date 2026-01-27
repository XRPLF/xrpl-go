package hash

import (
	"encoding/hex"
	"fmt"
	"strings"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
)

// Vault computes the hash of a Vault ledger entry.
// The hash is computed as SHA-512Half(ledgerSpaceHex('vault') + addressToHex(address) + sequence as 8-char hex).
//
// address is the account of the Vault Owner (Account submitting VaultCreate transaction).
// sequence is the sequence number of the Transaction that created the Vault object.
// Returns the computed hash of the Vault object.
func Vault(address string, sequence uint32) (string, error) {
	_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(address)
	if err != nil {
		return "", fmt.Errorf("failed to decode address: %w", err)
	}

	ledgerSpaceHex := "0056"
	addressHex := hex.EncodeToString(accountID)
	sequenceHex := fmt.Sprintf("%08x", sequence)

	payload := ledgerSpaceHex + addressHex + sequenceHex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	hash := crypto.Sha512Half(payloadBytes)
	return strings.ToUpper(hex.EncodeToString(hash)), nil
}

// LoanBroker computes the hash of a LoanBroker ledger entry.
// The hash is computed as SHA-512Half(ledgerSpaceHex('loanBroker') + addressToHex(address) + sequence as 8-char hex).
//
// address is the account of the Lender (Account submitting LoanBrokerSet transaction, i.e. Lender).
// sequence is the sequence number of the Transaction that created the LoanBroker object.
// Returns the computed hash of the LoanBroker object.
func LoanBroker(address string, sequence uint32) (string, error) {
	_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(address)
	if err != nil {
		return "", fmt.Errorf("failed to decode address: %w", err)
	}

	ledgerSpaceHex := "006C"
	addressHex := hex.EncodeToString(accountID)
	sequenceHex := fmt.Sprintf("%08x", sequence)

	payload := ledgerSpaceHex + addressHex + sequenceHex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	hash := crypto.Sha512Half(payloadBytes)
	return strings.ToUpper(hex.EncodeToString(hash)), nil
}

// Loan computes the hash of a Loan ledger entry.
// The hash is computed as SHA-512Half(ledgerSpaceHex('loan') + loanBrokerID + loanSequence as 8-char hex).
//
// loanBrokerID is the LoanBrokerID of the associated LoanBroker object.
// loanSequence is the sequence number of the Loan.
// Returns the computed hash of the Loan object.
func Loan(loanBrokerID string, loanSequence uint32) (string, error) {
	ledgerSpaceHex := "004C"
	sequenceHex := fmt.Sprintf("%08x", loanSequence)

	payload := ledgerSpaceHex + loanBrokerID + sequenceHex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	hash := crypto.Sha512Half(payloadBytes)
	return strings.ToUpper(hex.EncodeToString(hash)), nil
}
