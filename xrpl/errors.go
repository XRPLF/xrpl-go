package xrpl

import "errors"

var (
	// ErrNoTxToMultisign is returned when no transaction blobs are provided to Multisign.
	ErrNoTxToMultisign = errors.New("no transaction to multisign")
	// ErrMultisignNonEmptySigningPubKey is returned when SigningPubKey is not empty
	// on one or more transactions passed to Multisign, it must be an empty string for all.
	ErrMultisignNonEmptySigningPubKey = errors.New("SigningPubKey must be an empty string for all transactions when multisigning")
	// ErrMultisignTxNotEqual is returned when transaction blobs passed to Multisign
	// do not represent the same transaction (ignoring the Signers field).
	ErrMultisignTxNotEqual = errors.New("all transactions to multisign must be equal except for the Signers field")
	// ErrMultisignInvalidSignature is returned when a signer signature is invalid
	// for one or more transactions passed to Multisign.
	ErrMultisignInvalidSignature = errors.New("invalid multisign signer signature")
	// ErrMultisignInvalidSigner is returned when signer data in a multisign blob
	// is missing or has an unexpected shape.
	ErrMultisignInvalidSigner = errors.New("invalid multisign signer data")
	// ErrInvalidSigner is returned when a signer entry is malformed.
	ErrInvalidSigner = errors.New("invalid signer")
)
