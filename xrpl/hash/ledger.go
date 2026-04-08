package hash

import (
	"encoding/hex"
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
)

const (
	ledgerSpaceVault           = "0056"
	ledgerSpaceLoanBroker      = "006C"
	ledgerSpaceLoan            = "004C"
	ledgerSpaceMPToken         = "0074"
	ledgerSpaceMPTokenIssuance = "007E"
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

	addressHex := hex.EncodeToString(accountID)
	sequenceHex := fmt.Sprintf("%08x", sequence)

	payload := ledgerSpaceVault + addressHex + sequenceHex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	return EncodeToHashString(payloadBytes), nil
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

	addressHex := hex.EncodeToString(accountID)
	sequenceHex := fmt.Sprintf("%08x", sequence)

	payload := ledgerSpaceLoanBroker + addressHex + sequenceHex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	return EncodeToHashString(payloadBytes), nil
}

// Loan computes the hash of a Loan ledger entry.
// The hash is computed as SHA-512Half(ledgerSpaceHex('loan') + loanBrokerID + loanSequence as 8-char hex).
//
// loanBrokerID is the LoanBrokerID of the associated LoanBroker object.
// loanSequence is the sequence number of the Loan.
// Returns the computed hash of the Loan object.
func Loan(loanBrokerID string, loanSequence uint32) (string, error) {
	sequenceHex := fmt.Sprintf("%08x", loanSequence)

	payload := ledgerSpaceLoan + loanBrokerID + sequenceHex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	return EncodeToHashString(payloadBytes), nil
}

// EncodeToHashString computes SHA-512Half of the given bytes and returns it as an uppercase hex string.
func EncodeToHashString(bytes []byte) string {
	return hexutil.EncodeToUpperHex(crypto.Sha512Half(bytes))
}

// MPToken computes the hash of an MPToken ledger entry.
// The hash is computed as SHA-512Half(ledgerSpaceHex('mptoken') + MPTokenIssuance(issuanceIDHex) + addressToHex(holder)).
//
// issuanceIDHex is the 48-character hex-encoded MPTokenIssuanceID (Hash192, 24 bytes).
// holder is the classic address of the MPToken holder.
// Returns the computed hash of the MPToken object.
func MPToken(issuanceIDHex string, holder string) (string, error) {
	issuanceKey, err := MPTokenIssuance(issuanceIDHex)
	if err != nil {
		return "", fmt.Errorf("failed to compute issuance key: %w", err)
	}

	_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(holder)
	if err != nil {
		return "", fmt.Errorf("failed to decode holder address: %w", err)
	}

	holderHex := hex.EncodeToString(accountID)

	payload := ledgerSpaceMPToken + issuanceKey + holderHex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	return EncodeToHashString(payloadBytes), nil
}

// MPTokenIssuance computes the hash of an MPTokenIssuance ledger entry.
// The hash is computed as SHA512Half(ledgerSpaceHex('mptIssuance') + issuanceIDHex).
//
// issuanceIDHex is the 48-character hex-encoded MPTokenIssuanceID (Hash192, 24 bytes).
// Returns the computed hash of the MPTokenIssuance object.
func MPTokenIssuance(issuanceIDHex string) (string, error) {
	if len(issuanceIDHex) != 48 {
		return "", fmt.Errorf("issuance ID must be 48 hex chars (24 bytes), got %d", len(issuanceIDHex))
	}

	payload := ledgerSpaceMPTokenIssuance + issuanceIDHex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	return EncodeToHashString(payloadBytes), nil
}
