package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// XChainClaim completes a cross-chain transfer of value on the destination chain. It allows a user to claim the value on the
// destination chain equivalent to the value locked on the source chain. The user can send the funds to any account via the
// Destination field. This transaction is only needed if OtherChainDestination isn't specified in the XChainCommit transaction,
// or if the automatic transfer fails. (Requires the XChainBridge amendment)
//
// If the transaction succeeds in moving funds, the referenced XChainOwnedClaimID ledger object will be destroyed.
// This prevents transaction replay. If the transaction fails, the XChainOwnedClaimID won't be
// destroyed and the transaction can be re-run with different parameters.
//
// ```json
//
//	{
//	  "Account": "rahDmoXrtPdh7sUdrPjini3gcnTVYjbjjw",
//	  "Amount": "10000",
//	  "TransactionType": "XChainClaim",
//	  "XChainClaimID": "13f",
//	  "Destination": "rahDmoXrtPdh7sUdrPjini3gcnTVYjbjjw",
//	  "XChainBridge": {
//	    "LockingChainDoor": "rMAXACCrp3Y8PpswXcg3bKggHX76V3F8M4",
//	    "LockingChainIssue": {
//	      "currency": "XRP"
//	    },
//	    "IssuingChainDoor": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
//	    "IssuingChainIssue": {
//	      "currency": "XRP"
//	    }
//	  }
//	}
//
// ```
type XChainClaim struct {
	BaseTx

	// The amount to claim on the destination chain. This must match the amount attested to on the
	// attestations associated with this XChainClaimID.
	Amount types.CurrencyAmount
	// The destination account on the destination chain. It must exist or the transaction will fail.
	// However, if the transaction fails in this case, the sequence number and collected signatures
	// won't be destroyed, and the transaction can be rerun with a different destination.
	Destination types.Address
	// An integer destination tag.
	DestinationTag *uint32 `json:",omitempty"`
	// The bridge to use for the transfer.
	XChainBridge types.XChainBridge
	// The unique integer ID for the cross-chain transfer that was referenced in the corresponding XChainCommit transaction.
	XChainClaimID string
}

// TxType returns the type of the transaction.
func (x *XChainClaim) TxType() TxType {
	return XChainClaimTx
}

// Flatten returns a flattened version of the transaction.
func (x *XChainClaim) Flatten() FlatTransaction {
	flatTx := x.BaseTx.Flatten()

	flatTx["TransactionType"] = x.TxType().String()

	if x.Amount != nil {
		flatTx["Amount"] = x.Amount.Flatten()
	}

	if x.Destination != "" {
		flatTx["Destination"] = x.Destination.String()
	}

	if x.DestinationTag != nil {
		flatTx["DestinationTag"] = *x.DestinationTag
	}

	if x.XChainBridge != (types.XChainBridge{}) {
		flatTx["XChainBridge"] = x.XChainBridge.Flatten()
	}

	if x.XChainClaimID != "" {
		flatTx["XChainClaimID"] = x.XChainClaimID
	}

	return flatTx
}

// Validate validates the transaction.
func (x *XChainClaim) Validate() (bool, error) {
	_, err := x.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if ok, err := IsAmount(x.Amount, "Amount", true); !ok {
		return false, err
	}

	if !addresscodec.IsValidAddress(x.Destination.String()) {
		return false, ErrInvalidDestinationAddress
	}

	if x.XChainClaimID == "" || !typecheck.IsHex(x.XChainClaimID) {
		return false, ErrMissingXChainClaimID
	}

	if ok, err := x.XChainBridge.Validate(); !ok {
		return false, err
	}

	return true, nil
}
