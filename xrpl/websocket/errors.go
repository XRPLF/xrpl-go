package websocket

import (
	"errors"
	"fmt"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
)

const (
	// txnNotFound is the error message returned by the xrpl node when requesting for a not found transaction.
	txnNotFound = "txnNotFound"
	// actNotFound is the error message returned by the xrpl node when requesting for a not found account.
	actNotFound = "actNotFound"
)

var (
	// transaction

	// ErrMissingTxSignatureOrSigningPubKey is returned when a transaction has no complete signing form.
	ErrMissingTxSignatureOrSigningPubKey = errors.New("transaction must include a complete TxnSignature/SigningPubKey or Signers form")
	// ErrInvalidSignedTransaction is returned when signing fields are malformed, incomplete, empty, or mixed.
	ErrInvalidSignedTransaction = clientinternal.ErrInvalidSignedTransaction
	// ErrNilTransaction is returned when a nil transaction is submitted or autofilled.
	ErrNilTransaction = errors.New("transaction must not be nil")
	// ErrMissingLastLedgerSequenceInTransaction is returned when a reliable-submission transaction does not contain LastLedgerSequence.
	ErrMissingLastLedgerSequenceInTransaction = errors.New("missing LastLedgerSequence in transaction")
	// ErrMissingWallet is returned when a wallet is required but not provided for an unsigned transaction.
	ErrMissingWallet = errors.New("wallet must be provided when submitting an unsigned transaction")
	// ErrTransactionNotFound is retained for compatibility.
	//
	// Deprecated: no client operation returns this error.
	ErrTransactionNotFound = errors.New("transaction not found")
	// ErrPreliminaryResult indicates a malformed preliminary submit result.
	ErrPreliminaryResult = clientinternal.ErrPreliminaryResult
	// ErrTransactionExpired indicates ledger-driven expiry after LastLedgerSequence.
	ErrTransactionExpired = clientinternal.ErrTransactionExpired
	// ErrFinalityTransport indicates repeated transport failures during monitoring.
	ErrFinalityTransport = clientinternal.ErrFinalityTransport
	// ErrMissingAccountInTransaction is returned when the Account field is missing from a transaction.
	ErrMissingAccountInTransaction = errors.New("missing Account in transaction")
	// ErrInvalidFulfillmentLength is returned when the fulfillment length is invalid.
	ErrInvalidFulfillmentLength = errors.New("invalid fulfillment length")

	// fields

	// ErrAddressFieldIsNotAString is returned when an address field has the wrong Go type.
	ErrAddressFieldIsNotAString = clientinternal.ErrAddressFieldIsNotAString
	// ErrTagFieldIsNotAUint32 is returned when an explicit transaction tag has the wrong Go type.
	ErrTagFieldIsNotAUint32 = clientinternal.ErrTagFieldIsNotAUint32
	// ErrInvalidAddress is returned when an autofilled address is neither classic nor an X-address.
	ErrInvalidAddress = clientinternal.ErrInvalidAddress
	// ErrMismatchedTag is returned when an explicit transaction tag conflicts with an X-address tag.
	ErrMismatchedTag = clientinternal.ErrMismatchedTag
	// ErrAccountIDTagNotAllowed is returned when an X-address contains a tag for
	// a transaction field that cannot represent one.
	ErrAccountIDTagNotAllowed = clientinternal.ErrAccountIDTagNotAllowed
	// ErrRawTransactionsFieldIsNotAnArray is returned when the RawTransactions field is not an array type.
	ErrRawTransactionsFieldIsNotAnArray = clientinternal.ErrRawTransactionsFieldIsNotAnArray
	// ErrRawTransactionFieldIsNotAnObject is returned when the RawTransaction field is not an object type.
	ErrRawTransactionFieldIsNotAnObject = clientinternal.ErrRawTransactionFieldIsNotAnObject
	// ErrSigningPubKeyFieldMustBeEmpty is returned when the SigningPubKey field should be empty but isn't.
	ErrSigningPubKeyFieldMustBeEmpty = errors.New("field SigningPubKey must be empty")
	// ErrTxnSignatureFieldMustBeEmpty is returned when the TxnSignature field should be empty but isn't.
	ErrTxnSignatureFieldMustBeEmpty = errors.New("field TxnSignature must be empty")
	// ErrSignersFieldMustBeEmpty is returned when the Signers field should be empty but isn't.
	ErrSignersFieldMustBeEmpty = errors.New("field Signers must be empty")
	// ErrAccountFieldIsNotAString is returned when the Account field is not a string type.
	ErrAccountFieldIsNotAString = errors.New("field Account must be a string")
	// ErrNetworkIDFieldIsNotAUint32 is returned when the NetworkID field is set but not a uint32.
	ErrNetworkIDFieldIsNotAUint32 = clientinternal.ErrNetworkIDFieldIsNotAUint32
	// ErrNetworkIDFieldMismatch is returned when the NetworkID field does not match the expected NetworkID.
	ErrNetworkIDFieldMismatch = clientinternal.ErrNetworkIDFieldMismatch
	// ErrNetworkIDFieldUnexpected is returned when NetworkID must be omitted for the target network.
	ErrNetworkIDFieldUnexpected = clientinternal.ErrNetworkIDFieldUnexpected
	// ErrNetworkIDUnavailable is returned when server identity discovery did not produce a network ID.
	ErrNetworkIDUnavailable = clientinternal.ErrNetworkIDUnavailable
	// ErrBuildVersionUnavailable is returned when restricted-network policy cannot be determined without a build version.
	ErrBuildVersionUnavailable = clientinternal.ErrBuildVersionUnavailable
	// ErrInvalidBuildVersion is returned when the discovered rippled version cannot be compared.
	ErrInvalidBuildVersion = clientinternal.ErrInvalidBuildVersion
	// ErrNetworkIDOverrideMismatch is returned when a trusted override differs from server_info.
	ErrNetworkIDOverrideMismatch = clientinternal.ErrNetworkIDOverrideMismatch
	// ErrRawTransactionsFieldMissing is returned when the RawTransactions field is missing from a Batch transaction.
	ErrRawTransactionsFieldMissing = errors.New("RawTransactions field missing from Batch transaction")
	// ErrRawTransactionFieldMissing is returned when the RawTransaction field is missing from a wrapper.
	ErrRawTransactionFieldMissing = errors.New("RawTransaction field missing from wrapper")
	// ErrFeeFieldMissing is returned when the fee field is missing after calculation.
	ErrFeeFieldMissing = errors.New("fee field missing after calculation")

	// client

	// ErrIncorrectID indicates that a response contains an incorrect request ID.
	ErrIncorrectID = errors.New("incorrect id")
	// ErrNotConnectedToServer indicates that the client is not connected to a WebSocket server.
	ErrNotConnectedToServer = errors.New("not connected to server")
	// ErrRequestTimedOut indicates that a request to the server timed out.
	ErrRequestTimedOut = errors.New("request timed out")
	// ErrDisconnected indicates that a pending request could not receive a response because the connection closed.
	ErrDisconnected = errors.New("websocket disconnected before response")
	// ErrSignerDataIsEmpty is returned when signer data is empty or missing.
	ErrSignerDataIsEmpty = errors.New("signer data is empty")

	// wallet

	// ErrCannotFundWalletWithoutClassicAddress is returned when attempting to fund a wallet without a classic address.
	ErrCannotFundWalletWithoutClassicAddress = errors.New("cannot fund a wallet without a classic address")
	// ErrFundWalletBalanceNotUpdated is returned when the wallet balance does not update on the validated ledger after polling.
	ErrFundWalletBalanceNotUpdated = errors.New("fund wallet: balance did not update on validated ledger after polling")

	// fees

	// ErrInvalidFeeValue is returned when fee configuration is not a finite, non-negative decimal value.
	ErrInvalidFeeValue = clientinternal.ErrInvalidFeeValue
	// ErrFeeHasTooManyDecimals is returned when an XRP fee cannot be represented as whole drops.
	ErrFeeHasTooManyDecimals = clientinternal.ErrFeeHasTooManyDecimals
	// ErrCouldNotGetBaseFeeXrp is returned when BaseFeeXrp cannot be retrieved from ServerInfo.
	ErrCouldNotGetBaseFeeXrp = errors.New("get fee xrp: could not get BaseFeeXrp from ServerInfo")
	// ErrCouldNotFetchOwnerReserve is returned when the owner reserve fee cannot be fetched.
	ErrCouldNotFetchOwnerReserve = errors.New("could not fetch Owner Reserve")
	// ErrLoanBrokerIDRequired is returned when LoanBrokerID is required but not provided.
	ErrLoanBrokerIDRequired = errors.New("LoanBrokerID is required for LoanSet transaction")
	// ErrCouldNotFetchLoanBroker is returned when the LoanBroker cannot be fetched.
	ErrCouldNotFetchLoanBroker = errors.New("could not fetch LoanBroker")
	// ErrCouldNotFetchLoanBrokerOwner is returned when the Owner field cannot be extracted from LoanBroker.
	ErrCouldNotFetchLoanBrokerOwner = errors.New("could not fetch LoanBroker Owner")
	// ErrCounterpartyRequired is returned when Counterparty is required but not provided.
	ErrCounterpartyRequired = errors.New("field Counterparty is required")

	// account

	// ErrAccountCannotBeDeleted is returned when an account cannot be deleted due to associated objects.
	ErrAccountCannotBeDeleted = errors.New("account cannot be deleted; there are Escrows, PayChannels, RippleStates, or Checks associated with the account")

	// payment

	// ErrAmountAndDeliverMaxMustBeIdentical is returned when Amount and DeliverMax fields are not identical.
	ErrAmountAndDeliverMaxMustBeIdentical = errors.New("payment transaction: Amount and DeliverMax fields must be identical when both are provided")

	// connection

	// ErrNotConnected is returned when attempting to perform operations on a connection that is not established.
	ErrNotConnected = errors.New("connection is not connected")
	// ErrAlreadyConnected is returned when attempting to replace a live connection.
	ErrAlreadyConnected = errors.New("connection is already connected")
)

// Dynamic errors

// PreliminaryResultError reports a malformed preliminary engine result.
type PreliminaryResultError = clientinternal.PreliminaryResultError

// TransactionExpiredError reports expiry with ledger and preliminary-result diagnostics.
type TransactionExpiredError = clientinternal.TransactionExpiredError

// FinalityTransportError reports repeated monitoring query failures and unwraps their cause.
type FinalityTransportError = clientinternal.FinalityTransportError

// ClientError represents a dynamic error with a custom error message string.
type ClientError struct {
	ErrorString string
}

// Error returns the error message string for ClientError.
func (e *ClientError) Error() string {
	return e.ErrorString
}

// ErrUnknownStreamType is returned when an unknown stream type is encountered.
type ErrUnknownStreamType struct {
	Type any
}

// Error implements the error interface for ErrUnknownStreamType
func (e ErrUnknownStreamType) Error() string {
	return fmt.Sprintf("unknown stream type: %v", e.Type)
}

// ErrMaxReconnectionAttemptsReached is returned when maximum reconnection attempts are reached.
type ErrMaxReconnectionAttemptsReached struct {
	Attempts int
}

// Error implements the error interface for ErrMaxReconnectionAttemptsReached
func (e ErrMaxReconnectionAttemptsReached) Error() string {
	return fmt.Sprintf("max reconnection attempts reached: %d", e.Attempts)
}

// ErrFailedToParseFee is returned when fee parsing fails.
type ErrFailedToParseFee struct {
	Fee string
	Err error
}

// Error implements the error interface for ErrFailedToParseFee
func (e ErrFailedToParseFee) Error() string {
	return fmt.Sprintf("failed to parse fee %s: %v", e.Fee, e.Err)
}
