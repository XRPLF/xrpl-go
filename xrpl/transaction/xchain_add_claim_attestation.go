package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// XChainAddClaimAttestation transaction provides proof from a witness server,
// attesting to an XChainCommit transaction.
//
// The signature must be from one of the keys on the door's signer list at the time the signature was provided.
// However, if the signature list changes between the time the signature was submitted and the quorum is reached,
// the new signature set is used and some of the currently collected signatures may be removed.
//
// Any account can submit signatures.
// (Requires the XChainBridge amendment)
//
// ```json
//
//	{
//	  "TransactionType": "XChainAddClaimAttestation",
//	  "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
//	  "XChainAttestationBatch": {
//	    "XChainBridge": {
//	      "IssuingChainDoor": "rKeSSvHvaMZJp9ykaxutVwkhZgWuWMLnQt",
//	      "IssuingChainIssue": {
//	        "currency": "XRP"
//	      },
//	      "LockingChainDoor": "rJvExveLEL4jNDEeLKCVdxaSCN9cEBnEQC",
//	      "LockingChainIssue": {
//	        "currency": "XRP"
//	      }
//	    },
//	    "XChainClaimAttestationBatch" : [
//	      {
//	        "XChainClaimAttestationBatchElement" : {
//	          "Account" : "rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx",
//	          "Amount" : "100000000",
//	          "AttestationSignerAccount" : "rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx",
//	          "Destination" : "r9A8UyNpW3X46FUc6P7JZqgn6WgAPjBwPg",
//	          "PublicKey" : "03DAB289CA36FF377F3F4304C7A7203FDE5EDCBFC209F430F6A4355361425526D0",
//	          "Signature" : "616263",
//	          "WasLockingChainSend" : 1,
//	          "XChainClaimID" : "0000000000000000"
//	        }
//	      }
//	    ],
//	    "XChainCreateAccountAttestationBatch": [
//	      {
//	        "XChainCreateAccountAttestationBatchElement": {
//	          "Account": "rnJmYAiqEVngtnb5ckRroXLtCbWC7CRUBx",
//	          "Amount": "1000000000",
//	          "AttestationSignerAccount": "rEziJZmeZzsJvGVUmpUTey7qxQLKYxaK9f",
//	          "Destination": "rKT9gDkaedAosiHyHZTjyZs2HvXpzuiGmC",
//	          "PublicKey": "03ADB44CA8E56F78A0096825E5667C450ABD5C24C34E027BC1AAF7E5BD114CB5B5",
//	          "Signature": "3044022036C8B90F85E8073C465F00625248A72D4714600F98EBBADBAD3B7ED226109A3A02204C5A0AE12D169CF790F66541F3DB59C289E0D9CA7511FDFE352BB601F667A26",
//	          "SignatureReward": "1000000",
//	          "WasLockingChainSend": 1,
//	          "XChainAccountCreateCount": "0000000000000001"
//	        }
//	      }
//	    ]
//	  }
//	}
//
// ```
type XChainAddClaimAttestation struct {
	BaseTx

	// The amount committed by the XChainCommit transaction on the source chain.
	Amount types.CurrencyAmount
	// The account that should receive this signer's share of the SignatureReward.
	AttestationRewardAccount types.Address
	// The account on the door account's signer list that is signing the transaction.
	AttestationSignerAccount types.Address
	// The destination account for the funds on the destination chain (taken from the XChainCommit transaction).
	Destination types.Address `json:",omitempty"`
	// The account on the source chain that submitted the XChainCommit transaction
	// that triggered the event associated with the attestation.
	OtherChainSource types.Address
	// The public key used to verify the attestation signature.
	PublicKey string
	// The signature attesting to the event on the other chain.
	Signature string
	// A boolean representing the chain where the event occurred.
	WasLockingChainSend uint8
	// The bridge to use to transfer funds.
	XChainBridge types.XChainBridge
	// The XChainClaimID associated with the transfer, which was included in the XChainCommit transaction.
	XChainClaimID string
}

// TxType returns the type of the transaction.
func (x *XChainAddClaimAttestation) TxType() TxType {
	return XChainAddClaimAttestationTx
}

// Flatten returns a flattened version of the transaction.
func (x *XChainAddClaimAttestation) Flatten() FlatTransaction {
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

	flatTx["WasLockingChainSend"] = x.WasLockingChainSend

	if x.XChainBridge != (types.XChainBridge{}) {
		flatTx["XChainBridge"] = x.XChainBridge.Flatten()
	}

	if x.XChainClaimID != "" {
		flatTx["XChainClaimID"] = x.XChainClaimID
	}

	return flatTx
}

// Validate validates the transaction.
func (x *XChainAddClaimAttestation) Validate() (bool, error) {
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

	if x.Destination != "" && !addresscodec.IsValidAddress(x.Destination.String()) {
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

	if x.WasLockingChainSend != 0 && x.WasLockingChainSend != 1 {
		return false, ErrInvalidWasLockingChainSend
	}

	if x.XChainClaimID == "" || !typecheck.IsHex(x.XChainClaimID) {
		return false, ErrInvalidXChainClaimID
	}

	return x.XChainBridge.Validate()
}
