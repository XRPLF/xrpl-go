package builder

import "errors"

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
	ErrMissingSenderState  = errors.New("builder: current balance and ciphertext are required")
	ErrMissingCiphertext   = errors.New("builder: issuer ciphertext is required")
	ErrSelfSend            = errors.New("builder: sender and destination cannot be the same")
	ErrSelfClawback        = errors.New("builder: issuer and holder cannot be the same")
	ErrZeroAmount          = errors.New("builder: amount must be greater than zero")
	ErrInsufficientBalance = errors.New("builder: amount exceeds current balance")
	ErrLedgerQuery         = errors.New("builder: ledger query failed")
	ErrEncryptionKeyNotSet = errors.New("builder: encryption key not registered on issuance")
	ErrReceiverNotOptedIn  = errors.New("builder: receiver has no encryption key registered")
	ErrMPTokenNotFound     = errors.New("builder: MPToken ledger entry not found")
	ErrCryptoFailed        = errors.New("builder: cryptographic operation failed")
)
