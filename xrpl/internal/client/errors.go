package client

import (
	"errors"

	codectypes "github.com/Peersyst/xrpl-go/binary-codec/types"
)

var (
	// address

	// ErrAddressFieldIsNotAString indicates that an address-bearing transaction
	// field has the wrong Go type.
	ErrAddressFieldIsNotAString = errors.New("transaction address field must be a string")
	// ErrTagFieldIsNotAUint32 indicates that an explicit source or destination
	// tag has the wrong Go type.
	ErrTagFieldIsNotAUint32 = errors.New("transaction tag field must be a uint32")
	// ErrInvalidAddress indicates that an in-scope transaction address is neither
	// a valid classic address nor a valid X-address.
	ErrInvalidAddress = errors.New("invalid transaction address")
	// ErrMismatchedTag indicates that an explicit transaction tag conflicts with
	// the tag embedded in an X-address.
	ErrMismatchedTag = errors.New("transaction tag mismatch")
	// ErrAccountIDTagNotAllowed indicates that an X-address contains a tag for
	// a transaction field that cannot represent one.
	ErrAccountIDTagNotAllowed = codectypes.ErrAccountIDTagNotAllowed

	// network

	// ErrNetworkIDUnavailable indicates that a client cannot safely determine
	// the server's network identity.
	ErrNetworkIDUnavailable = errors.New("server network ID is unavailable")
	// ErrBuildVersionUnavailable indicates that a restricted network's rippled
	// version is unavailable, so NetworkID requiredness cannot be determined.
	ErrBuildVersionUnavailable = errors.New("server build version is unavailable")
	// ErrInvalidBuildVersion indicates that a restricted network returned a
	// build version that cannot be compared with rippled 1.11.0.
	ErrInvalidBuildVersion = errors.New("invalid server build version")
	// errInvalidRippledVersionFormat indicates that a rippled version does not
	// contain its required major, minor, and patch components.
	errInvalidRippledVersionFormat = errors.New("version must have major, minor, and patch components")
	// ErrNetworkIDOverrideMismatch indicates that a trusted client override does
	// not match the identity discovered from server_info.
	ErrNetworkIDOverrideMismatch = errors.New("configured network ID does not match server network ID")
	// ErrNetworkIDFieldIsNotAUint32 indicates that a transaction NetworkID value
	// has the wrong Go type.
	ErrNetworkIDFieldIsNotAUint32 = errors.New("field NetworkID must be a uint32")
	// ErrNetworkIDFieldMismatch indicates that a transaction NetworkID value
	// does not match the client identity.
	ErrNetworkIDFieldMismatch = errors.New("field NetworkID must match expected NetworkID")
	// ErrNetworkIDFieldUnexpected indicates that NetworkID was supplied on a
	// network or rippled version where the field must be omitted.
	ErrNetworkIDFieldUnexpected = errors.New("field NetworkID must be omitted for this network")

	// batch

	// ErrRawTransactionsFieldIsNotAnArray indicates a malformed Batch wrapper.
	ErrRawTransactionsFieldIsNotAnArray = errors.New("field RawTransactions must be an array")
	// ErrRawTransactionFieldIsNotAnObject indicates a malformed inner Batch wrapper.
	ErrRawTransactionFieldIsNotAnObject = errors.New("field RawTransaction must be an object")
	// ErrBatchRawTransactionsCount indicates that a Batch has fewer than two or more than eight inner transactions.
	ErrBatchRawTransactionsCount = errors.New("Batch RawTransactions must contain between 2 and 8 transactions")
	// ErrSigningPubKeyFieldMustBeEmpty indicates that an inner Batch SigningPubKey is not an empty string.
	ErrSigningPubKeyFieldMustBeEmpty = errors.New("field SigningPubKey must be empty")
	// ErrTxnSignatureFieldMustBeEmpty indicates that an inner Batch contains a forbidden TxnSignature field.
	ErrTxnSignatureFieldMustBeEmpty = errors.New("field TxnSignature must be empty")
	// ErrSignersFieldMustBeEmpty indicates that an inner Batch contains a forbidden Signers field.
	ErrSignersFieldMustBeEmpty = errors.New("field Signers must be empty")
	// ErrLastLedgerSequenceFieldMustBeAbsent indicates that an inner Batch contains a forbidden LastLedgerSequence field.
	ErrLastLedgerSequenceFieldMustBeAbsent = errors.New("field LastLedgerSequence must be absent")
	// ErrAccountFieldIsNotAString indicates that an inner Batch Account field is not a string.
	ErrAccountFieldIsNotAString = errors.New("field Account must be a string")

	// transaction

	// ErrInvalidSignedTransaction indicates that transaction signing fields do not
	// form a complete single-sign, multisign, or permitted inner-Batch structure.
	ErrInvalidSignedTransaction = errors.New("transaction has an invalid signed form")
	// ErrTransactionNotMultisigned indicates that SubmitMultisigned received a transaction in another signing form.
	ErrTransactionNotMultisigned = errors.New("transaction is not multisigned")
	// ErrAmountAndDeliverMaxMustBeIdentical indicates that a Payment has conflicting Amount and DeliverMax values.
	ErrAmountAndDeliverMaxMustBeIdentical = errors.New("payment transaction: Amount and DeliverMax fields must be identical when both are provided")
)
