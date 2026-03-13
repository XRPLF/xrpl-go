package xrpl

import "errors"

var (
	// ErrNoTxToMultisign is returned when no transaction blobs are provided to Multisign.
	ErrNoTxToMultisign = errors.New("no transaction to multisign")
	// ErrMultisignNonEmptySigningPubKey is returned when SigningPubKey is not empty
	// on one or more transactions passed to Multisign, it must be an empty string for all.
	ErrMultisignNonEmptySigningPubKey = errors.New("SigningPubKey must be an empty string for all transactions when multisigning")
)
