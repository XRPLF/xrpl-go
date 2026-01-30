package websocket

import (
	"errors"
	"fmt"
)

const (
	// txnNotFound is the error message returned by the xrpl node when requesting for a not found transaction.
	txnNotFound = "txnNotFound"
)

var (
	// transaction

	// ErrMissingTxSignatureOrSigningPubKey is returned when a transaction lacks both TxSignature and SigningPubKey.
	ErrMissingTxSignatureOrSigningPubKey = errors.New("transaction must include either TxSignature or SigningPubKey")
	// ErrMissingLastLedgerSequenceInTransaction is returned when LastLedgerSequence is missing from a transaction.
	ErrMissingLastLedgerSequenceInTransaction = errors.New("missing LastLedgerSequence in transaction")
	// ErrMissingWallet is returned when a wallet is required but not provided for an unsigned transaction.
	ErrMissingWallet = errors.New("wallet must be provided when submitting an unsigned transaction")
	// ErrTransactionNotFound is returned when a transaction cannot be found.
	ErrTransactionNotFound = errors.New("transaction not found")
	// ErrMissingAccountInTransaction is returned when the Account field is missing from a transaction.
	ErrMissingAccountInTransaction = errors.New("missing Account in transaction")
	// ErrTransactionTypeMissing is returned when the transaction type is missing from a transaction.
	ErrTransactionTypeMissing = errors.New("transaction type is missing in transaction")
	// ErrInvalidFulfillmentLength is returned when the fulfillment length is invalid.
	ErrInvalidFulfillmentLength = errors.New("invalid fulfillment length")

	// fields

	// ErrRawTransactionsFieldIsNotAnArray is returned when the RawTransactions field is not an array type.
	ErrRawTransactionsFieldIsNotAnArray = errors.New("field RawTransactions must be an array")
	// ErrRawTransactionFieldIsNotAnObject is returned when the RawTransaction field is not an object type.
	ErrRawTransactionFieldIsNotAnObject = errors.New("field RawTransaction must be an object")
	// ErrSigningPubKeyFieldMustBeEmpty is returned when the SigningPubKey field should be empty but isn't.
	ErrSigningPubKeyFieldMustBeEmpty = errors.New("field SigningPubKey must be empty")
	// ErrTxnSignatureFieldMustBeEmpty is returned when the TxnSignature field should be empty but isn't.
	ErrTxnSignatureFieldMustBeEmpty = errors.New("field TxnSignature must be empty")
	// ErrSignersFieldMustBeEmpty is returned when the Signers field should be empty but isn't.
	ErrSignersFieldMustBeEmpty = errors.New("field Signers must be empty")
	// ErrAccountFieldIsNotAString is returned when the Account field is not a string type.
	ErrAccountFieldIsNotAString = errors.New("field Account must be a string")
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
	// ErrSignerDataIsEmpty is returned when signer data is empty or missing.
	ErrSignerDataIsEmpty = errors.New("signer data is empty")

	// wallet

	// ErrCannotFundWalletWithoutClassicAddress is returned when attempting to fund a wallet without a classic address.
	ErrCannotFundWalletWithoutClassicAddress = errors.New("cannot fund a wallet without a classic address")

	// fees

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
)

// Dynamic errors

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
	Type interface{}
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
