package client

import (
	"errors"

	binarycodectypes "github.com/Peersyst/xrpl-go/binary-codec/types"
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
	// ErrAccountIDTagNotAllowed indicates that a tagless AccountID field received
	// an X-address with an embedded tag.
	ErrAccountIDTagNotAllowed = binarycodectypes.ErrAccountIDTagNotAllowed

	// network

	// ErrInvalidBuildVersion indicates that a restricted network returned a
	// build version that cannot be compared with rippled 1.11.0.
	ErrInvalidBuildVersion = errors.New("invalid server build version")
	// errInvalidRippledVersionFormat indicates that a rippled version does not
	// contain its required major, minor, and patch components.
	errInvalidRippledVersionFormat = errors.New("version must have major, minor, and patch components")
	// ErrNetworkIDOverrideMismatch indicates that a client override does not
	// match the identity discovered from server_info.
	ErrNetworkIDOverrideMismatch = errors.New("configured network ID does not match server network ID")
	// ErrNetworkIDOverrideUnverified indicates that server_info did not include
	// the network ID needed to verify a client override.
	ErrNetworkIDOverrideUnverified = errors.New("configured network ID cannot be verified because server network ID is missing")
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
)
