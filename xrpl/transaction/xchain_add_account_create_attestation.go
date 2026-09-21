package transaction

import (
	"strconv"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// XChainAddAccountCreateAttestation provides an attestation that an XChainAccountCreateCommit transaction occurred on the other chain. (Requires the XChainBridge amendment)
//
// The signature must be from one of the keys on the door's signer list at the time the signature was provided.
// If the signature list changes between the time the signature was submitted and the quorum is reached,
// the new signature set is used and some of the currently collected signatures may be removed.
//
// Any account can submit signatures.
//
// ```json
//
//	{
//	  "Account": "rDr5okqGKmMpn44Bbhe5WAfDQx8e9XquEv",
//	  "TransactionType": "XChainAddAccountCreateAttestation",
//	  "OtherChainSource": "rUzB7yg1LcFa7m3q1hfrjr5w53vcWzNh3U",
//	  "Destination": "rJMfWNVbyjcCtds8kpoEjEbYQ41J5B6MUd",
//	  "Amount": "2000000000",
//	  "PublicKey": "EDF7C3F9C80C102AF6D241752B37356E91ED454F26A35C567CF6F8477960F66614",
//	  "Signature": "F95675BA8FDA21030DE1B687937A79E8491CE51832D6BEEBC071484FA5AF5B8A0E9AFF11A4AA46F09ECFFB04C6A8DAE8284AF3ED8128C7D0046D842448478500",
//	  "WasLockingChainSend": 1,
//	  "AttestationRewardAccount": "rpFp36UHW6FpEcZjZqq5jSJWY6UCj3k4Es",
//	  "AttestationSignerAccount": "rpWLegmW9WrFBzHUj7brhQNZzrxgLj9oxw",
//	  "XChainAccountCreateCount": "2",
//	  "SignatureReward": "204",
//	  "XChainBridge": {
//	    "LockingChainDoor": "r3nCVTbZGGYoWvZ58BcxDmiMUU7ChMa1eC",
//	    "LockingChainIssue": {
//	      "currency": "XRP"
//	    },
//	    "IssuingChainDoor": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
//	    "IssuingChainIssue": {
//	      "currency": "XRP"
//	    }
//	  },
//	  "Fee": "20"
//	}
//
// ```
type XChainAddAccountCreateAttestation struct {
	BaseTx

	// The amount committed by the XChainAccountCreateCommit transaction on the source chain.
	Amount types.CurrencyAmount
	// The account that should receive this signer's share of the SignatureReward.
	AttestationRewardAccount types.Address
	// The account on the door account's signer list that is signing the transaction.
	AttestationSignerAccount types.Address
	// The destination account for the funds on the destination chain.
	Destination types.Address
	// The account on the source chain that submitted the XChainAccountCreateCommit transaction
	// that triggered the event associated with the attestation.
	OtherChainSource types.Address
	// The public key used to verify the signature.
	PublicKey string
	// The signature attesting to the event on the other chain.
	Signature string
	// The signature reward paid in the XChainAccountCreateCommit transaction.
	SignatureReward types.CurrencyAmount
	// A boolean representing the chain where the event occurred.
	WasLockingChainSend uint8
	// The counter that represents the order that the claims must be processed in.
	XChainAccountCreateCount string
	// The bridge associated with the attestation.
	XChainBridge types.XChainBridge
}

// TxType returns the TxType for XChainAddAccountCreateAttestation transactions.
func (x *XChainAddAccountCreateAttestation) TxType() TxType {
	return XChainAddAccountCreateAttestationTx
}

// Flatten returns a FlatTransaction map representing the XChainAddAccountCreateAttestation transaction.
func (x *XChainAddAccountCreateAttestation) Flatten() FlatTransaction {
	flatTx := x.BaseTx.Flatten()

	flatTx["TransactionType"] = x.TxType().String()

	if x.Amount != nil {
		flatTx["Amount"] = x.Amount.Flatten()
	}

	if x.AttestationRewardAccount != "" {
		flatTx["AttestationRewardAccount"] = x.AttestationRewardAccount.String()
	}

	if x.AttestationSignerAccount != "" {
		flatTx["AttestationSignerAccount"] = x.AttestationSignerAccount.String()
	}

	if x.Destination != "" {
		flatTx["Destination"] = x.Destination.String()
	}

	if x.OtherChainSource != "" {
		flatTx["OtherChainSource"] = x.OtherChainSource.String()
	}

	if x.PublicKey != "" {
		flatTx["PublicKey"] = x.PublicKey
	}

	if x.Signature != "" {
		flatTx["Signature"] = x.Signature
	}

	if x.SignatureReward != nil {
		flatTx["SignatureReward"] = x.SignatureReward.Flatten()
	}

	flatTx["WasLockingChainSend"] = x.WasLockingChainSend

	if x.XChainAccountCreateCount != "" {
		flatTx["XChainAccountCreateCount"] = x.XChainAccountCreateCount
	}

	if x.XChainBridge != (types.XChainBridge{}) {
		flatTx["XChainBridge"] = x.XChainBridge.Flatten()
	}

	return flatTx
}

// Validate checks XChainAddAccountCreateAttestation fields and returns false and an error if invalid.
func (x *XChainAddAccountCreateAttestation) Validate() (bool, error) {
	_, err := x.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if ok, err := IsAmount(x.Amount, "Amount", true); !ok {
		return false, err
	}

	if !addresscodec.IsValidAddress(x.AttestationRewardAccount.String()) {
		return false, ErrInvalidAttestationRewardAccount
	}

	if !addresscodec.IsValidAddress(x.AttestationSignerAccount.String()) {
		return false, ErrInvalidAttestationSignerAccount
	}

	if !addresscodec.IsValidAddress(x.Destination.String()) {
		return false, ErrInvalidDestination
	}

	if !addresscodec.IsValidAddress(x.OtherChainSource.String()) {
		return false, ErrInvalidOtherChainSource
	}

	if !isPublicKey(x.PublicKey) {
		return false, ErrInvalidPublicKey
	}

	if !typecheck.IsHexBlob(x.Signature) {
		return false, ErrInvalidSignature
	}

	if ok, err := IsAmount(x.SignatureReward, "SignatureReward", true); !ok {
		return false, err
	}

	if x.WasLockingChainSend != 0 && x.WasLockingChainSend != 1 {
		return false, ErrInvalidWasLockingChainSend
	}

	if _, err := strconv.ParseUint(x.XChainAccountCreateCount, 10, 64); err != nil {
		return false, ErrInvalidXChainAccountCreateCount
	}

	return x.XChainBridge.Validate()
}
