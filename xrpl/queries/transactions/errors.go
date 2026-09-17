package transactions

import "errors"

var (
	// ErrNoTxBlob is returned when no TxBlob is defined in the SubmitRequest.
	ErrNoTxBlob = errors.New("no TxBlob defined")
	// ErrInvalidSimulateRequest is returned when a simulation request is nil.
	ErrInvalidSimulateRequest = errors.New("simulate: request must not be nil")
	// ErrInvalidSimulateTxJSON was returned for malformed simulation JSON input.
	//
	// Deprecated: simulation request validation is delegated to the server.
	ErrInvalidSimulateTxJSON = errors.New("simulate: tx_json is malformed")
	// ErrInvalidSimulateTxBlob was returned for malformed simulation blob input.
	//
	// Deprecated: simulation request validation is delegated to the server.
	ErrInvalidSimulateTxBlob = errors.New("simulate: tx_blob must be an even-length hexadecimal string")
	// ErrSignedSimulateTransaction was returned for signed simulation input.
	//
	// Deprecated: simulation request validation is delegated to the server.
	ErrSignedSimulateTransaction = errors.New("simulate: transaction must be unsigned")
	// ErrInvalidSimulateNetworkID was returned for malformed simulation NetworkID values.
	//
	// Deprecated: simulation request validation is delegated to the server.
	ErrInvalidSimulateNetworkID = errors.New("simulate: NetworkID must be an unsigned 32-bit integer")
	// ErrMismatchedSimulateNetworkID was returned for simulation NetworkID mismatches.
	//
	// Deprecated: simulation request validation is delegated to the server.
	ErrMismatchedSimulateNetworkID = errors.New("simulate: NetworkID does not match the target network")
	// ErrInvalidSimulateResponse is returned when a simulate response does not match its JSON or binary wire variant.
	ErrInvalidSimulateResponse = errors.New("simulate: response must contain exactly one valid JSON or binary transaction payload")
)
