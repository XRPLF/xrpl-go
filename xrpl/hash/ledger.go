package hash

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/pkg/hexutil"
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

	ledgerSpaceHex := "006C"
	addressHex := hex.EncodeToString(accountID)
	sequenceHex := fmt.Sprintf("%08x", sequence)

	payload := ledgerSpaceHex + addressHex + sequenceHex
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
	ledgerSpaceHex := "004C"
	sequenceHex := fmt.Sprintf("%08x", loanSequence)

	payload := ledgerSpaceHex + loanBrokerID + sequenceHex
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

// PaymentChannel computes the hash (channel ID) of a PaymentChannel ledger entry.
// The hash is computed as SHA-512Half(0x0078 + sourceAccountID + destAccountID + sequence as 8-char hex).
//
// source is the source address of the payment channel.
// destination is the destination address of the payment channel.
// sequence is the sequence number of the PaymentChannelCreate transaction.
// Returns the computed channel ID.
func PaymentChannel(source, destination string, sequence uint32) (string, error) {
	_, sourceID, err := addresscodec.DecodeClassicAddressToAccountID(source)
	if err != nil {
		return "", fmt.Errorf("failed to decode source classic address: %w", err)
	}

	_, destID, err := addresscodec.DecodeClassicAddressToAccountID(destination)
	if err != nil {
		return "", fmt.Errorf("failed to decode destination classic address: %w", err)
	}

	ledgerSpaceHex := "0078"
	sourceHex := hex.EncodeToString(sourceID)
	destHex := hex.EncodeToString(destID)
	sequenceHex := fmt.Sprintf("%08x", sequence)

	payload := ledgerSpaceHex + sourceHex + destHex + sequenceHex
	payloadBytes, err := hex.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	return EncodeToHashString(payloadBytes), nil
}

// MPTID computes the unique identifier for a Multi-Purpose Token (MPT).
// The ID is computed by concatenating the sequence number (as 8-char hex)
// and the account ID of the issuer.
//
// sequence is the sequence number of the MPTokenIssuance transaction.
// issuer is the account address that issued the token.
// Returns the computed MPT ID as an uppercase hexadecimal string.
func MPTID(sequence uint32, issuer string) (string, error) {
	_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(issuer)
	if err != nil {
		return "", fmt.Errorf("failed to decode issuer classic address: %w", err)
	}

	seqBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(seqBytes, sequence)
	return strings.ToUpper(hex.EncodeToString(seqBytes) + hex.EncodeToString(accountID)), nil
}
