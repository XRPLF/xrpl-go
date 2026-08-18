package builder

import (
	"errors"
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// Builder validation errors.
var (
	ErrMissingAccount      = errors.New("builder: account is required")
	ErrMissingIssuanceID   = errors.New("builder: issuance ID is required")
	ErrMissingHolderKey    = errors.New("builder: holder private/public key is required")
	ErrMissingIssuerKey    = errors.New("builder: issuer private/public key is required")
	ErrMissingSenderKey    = errors.New("builder: sender private/public key is required")
	ErrMissingReceiverKey  = errors.New("builder: receiver public key is required")
	ErrMissingDestination  = errors.New("builder: destination is required")
	ErrMissingHolder       = errors.New("builder: holder is required")
	ErrMissingSenderState  = errors.New("builder: current confidential balance state is required")
	ErrMissingCiphertext   = errors.New("builder: issuer ciphertext is required")
	ErrMissingSequence     = errors.New("builder: a final transaction sequence is required")
	ErrSelfSend            = errors.New("builder: sender and destination cannot be the same")
	ErrSelfClawback        = errors.New("builder: issuer and holder cannot be the same")
	ErrZeroAmount          = errors.New("builder: amount must be greater than zero")
	ErrAmountTooLarge      = errors.New("builder: amount exceeds the maximum MPT amount")
	ErrIssuerNotAllowed    = errors.New("builder: account cannot be the issuance issuer")
	ErrDestinationIsIssuer = errors.New("builder: destination cannot be the issuance issuer")
	ErrNotIssuer           = errors.New("builder: account must be the issuance issuer")
	ErrInsufficientBalance = errors.New("builder: amount exceeds current balance")
	ErrLedgerQuery         = errors.New("builder: ledger query failed")
	ErrEncryptionKeyNotSet = errors.New("builder: encryption key not registered on issuance")
	ErrReceiverNotOptedIn  = errors.New("builder: receiver has no encryption key registered")
	ErrMPTokenNotFound     = errors.New("builder: MPToken ledger entry not found")
	ErrIssuanceNotFound    = errors.New("builder: MPTokenIssuance ledger entry not found")
	ErrCryptoFailed        = errors.New("builder: cryptographic operation failed")
	ErrKeyMismatch         = errors.New("builder: public key does not match the key registered on the ledger")

	// Issuance capability errors mirror the preclaim conditions the confidential MPT
	// transactors enforce, so a doomed transaction never costs a fee and a sequence.
	ErrConfidentialDisabled     = errors.New("builder: issuance does not allow confidential balances")
	ErrTransferDisabled         = errors.New("builder: issuance does not allow transfers")
	ErrTransferFeeSet           = errors.New("builder: issuance with a transfer fee cannot send confidentially")
	ErrAmountExceedsOutstanding = errors.New("builder: amount exceeds the issuance confidential outstanding amount")

	// ErrInvalidAddress names an address that failed to decode where the field it came
	// from is not known. The query helpers use it because they resolve an MPToken for an
	// Account, a Destination, or a Holder depending on the caller, so naming any one of
	// those would be wrong for the others.
	ErrInvalidAddress = errors.New("builder: invalid address")

	ErrInvalidAccount     = errors.New("builder: invalid account address")
	ErrInvalidDestination = errors.New("builder: invalid destination address")
	ErrInvalidHolder      = errors.New("builder: invalid holder address")
	ErrInvalidIssuanceID  = errors.New("builder: issuance ID must be exactly 48 hex characters")
	ErrInvalidPrivKey     = errors.New("builder: private key must be a non-zero secp256k1 scalar")
	ErrInvalidPubKey      = errors.New("builder: public key must be a valid 33-byte compressed secp256k1 point")
	ErrInvalidCiphertext  = errors.New("builder: ciphertext must contain two valid compressed secp256k1 points")

	// ErrInvalidCredentialIDs wraps the transaction sentinel the shared CredentialIDs
	// validator raises, so a caller matching the builder error set does not have to import
	// xrpl/transaction for this one case while errors.Is still matches either sentinel.
	ErrInvalidCredentialIDs = fmt.Errorf("builder: %w", transaction.ErrInvalidCredentialIDs)
)
