// Package types contains data structures for subscription stream types.
//
//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
)

// The ValidationStream sends messages whenever it receives validation messages,
// also called validation votes, regardless of whether or not the validation message
// is from a trusted validator.
type ValidationStream struct {
	// The value validationReceived indicates this is from the validations stream.
	Type Type `json:"type"`
	// (May be omitted) The amendments this server wants to be added to the protocol.
	Amendments []string `json:"amendments,omitempty"`
	// (May be omitted) The unscaled transaction cost (reference_fee value) this server
	// wants to set by Fee Voting.
	BaseFee int `json:"base_fee,omitempty"`
	// (May be omitted) An arbitrary value chosen by the server at startup. If the same
	// validation key pair signs validations with different cookies concurrently, that
	// usually indicates that multiple servers are incorrectly configured to use the same
	// validation key pair.
	Cookie any `json:"cookie,omitempty"`
	// Bit-mask of flags added to this validation message. The flag 0x80000000 indicates
	// that the validation signature is fully-canonical. The flag 0x00000001 indicates
	// that this is a full validation; otherwise it's a partial validation. Partial
	// validations are not meant to vote for any particular ledger. A partial validation
	// indicates that the validator is still online but not keeping up with consensus.
	Flags uint32 `json:"flags"`
	// If true, this is a full validation. Otherwise, this is a partial validation.
	// Partial validations are not meant to vote for any particular ledger. A partial validation
	// indicates that the validator is still online but not keeping up with consensus.
	Full bool `json:"full"`
	// The identifying hash of the proposed ledger is being validated.
	LedgerHash common.LedgerHash `json:"ledger_hash"`
	// The Ledger Index of the proposed ledger.
	LedgerIndex common.LedgerIndex `json:"ledger_index"`
	// (May be omitted) The local load-scaled transaction cost this validator is currently enforcing,
	// in fee units.
	LoadFee int `json:"load_fee,omitempty"`
	// (May be omitted) The validator's master public key, if the validator is using a validator
	// token, in the XRP Ledger's base58 format. (See also: Enable Validation on your rippled Server.)
	MasterKey string `json:"master_key,omitempty"`
	// (May be omitted) The minimum reserve requirement (account_reserve value) this validator wants
	// to set by Fee Voting.
	ReserveBase int `json:"reserve_base,omitempty"`
	// (May be omitted) The increment in the reserve requirement (owner_reserve value) this validator
	// wants to set by Fee Voting.
	ReserveInc int `json:"reserve_inc,omitempty"`
	// (May be omitted) An 64-bit integer that encodes the version number of the validating server.
	// For example, "1745990410175512576". Only provided once every 256 ledgers.
	ServerVersion any `json:"server_version,omitempty"`
	// The signature that the validator used to sign its vote for this ledger.
	Signature string `json:"signature"`
	// When this validation vote was signed, in seconds since the Ripple Epoch.
	SigningTime uint64 `json:"signing_time"`
	// The unique hash of the proposed ledger this validation applies to.
	ValidatedHash string `json:"validated_hash"`
	// The public key from the key-pair that the validator used to sign the message, in the XRP
	// Ledger's base58 format. This identifies the validator sending the message and can also be
	// used to verify the signature. If the validator is using a token, this is an ephemeral
	// public key.
	ValidationPublicKey string `json:"validation_public_key"`
}
