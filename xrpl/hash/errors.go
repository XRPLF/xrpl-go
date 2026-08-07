package hash

import (
	"errors"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
)

var (
	// transaction signature

	// ErrNonSignedTransaction indicates that a transaction lacks the required signature fields.
	ErrNonSignedTransaction = errors.New("transaction must have at least one of TxnSignature, Signers, or SigningPubKey")
	// ErrMissingSignature is no longer returned.
	//
	// Deprecated: SignTx and SignTxBlob now return ErrNonSignedTransaction or
	// ErrInvalidSignedTransaction.
	ErrMissingSignature = ErrNonSignedTransaction
	// ErrInvalidSignedTransaction is returned when signing fields are incomplete, empty,
	// malformed, or mixed between single-sign and multisign forms.
	ErrInvalidSignedTransaction = clientinternal.ErrInvalidSignedTransaction
)
