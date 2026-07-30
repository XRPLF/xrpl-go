package transactions

import "errors"

var (
	// ErrNoTxBlob is returned when no TxBlob is defined in the SubmitRequest.
	ErrNoTxBlob = errors.New("no TxBlob defined")
	// ErrInvalidSimulateRequest is returned unless exactly one simulation input is provided.
	ErrInvalidSimulateRequest = errors.New("simulate: exactly one of tx_json or tx_blob must be specified")
	// ErrInvalidSimulateTxJSON is returned when tx_json is not a structurally valid unsigned transaction object.
	ErrInvalidSimulateTxJSON = errors.New("simulate: tx_json is malformed")
	// ErrInvalidSimulateTxBlob is returned when tx_blob is not a valid serialized hexadecimal transaction.
	ErrInvalidSimulateTxBlob = errors.New("simulate: tx_blob must be a valid serialized hexadecimal transaction")
	// ErrSignedSimulateTransaction is returned when a transaction contains non-empty signature data.
	ErrSignedSimulateTransaction = errors.New("simulate: transaction must be unsigned")
	// ErrInvalidSimulateNetworkID is returned when NetworkID is not a uint32-compatible number.
	ErrInvalidSimulateNetworkID = errors.New("simulate: NetworkID must be an unsigned 32-bit integer")
	// ErrMissingSimulateNetworkID is returned when a restricted target network requires an explicit NetworkID.
	ErrMissingSimulateNetworkID = errors.New("simulate: explicit NetworkID is required for the target network")
	// ErrMismatchedSimulateNetworkID is returned when NetworkID differs from the client's target network.
	ErrMismatchedSimulateNetworkID = errors.New("simulate: NetworkID does not match the target network")
	// ErrInvalidSimulateResponse is returned when a simulate response does not match its JSON or binary wire variant.
	ErrInvalidSimulateResponse = errors.New("simulate: response must contain exactly one valid JSON or binary transaction payload")
)
