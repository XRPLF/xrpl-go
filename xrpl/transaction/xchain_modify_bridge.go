package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TfClearAccountCreateAmount if enabled, indicates that the MinAccountCreateAmount field should be cleared from the bridge.
	TfClearAccountCreateAmount uint32 = 0x00010000
)

// XChainModifyBridge modifies an existing Bridge ledger object, updating its flags, minimum account create amount, and signature reward.
// (Requires the XChainBridge amendment)
//
// Example:
// ```json
//
//	{
//	  "TransactionType": "XChainModifyBridge",
//	  "Account": "rhWQzvdmhf5vFS35vtKUSUwNZHGT53qQsg",
//	  "XChainBridge": {
//	    "LockingChainDoor": "rhWQzvdmhf5vFS35vtKUSUwNZHGT53qQsg",
//	    "LockingChainIssue": {
//	      "currency": "XRP"
//	    },
//	    "IssuingChainDoor": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
//	    "IssuingChainIssue": {
//	      "currency": "XRP"
//	    }
//	  },
//	  "SignatureReward": 200,
//	  "MinAccountCreateAmount": 1000000
//	}
//
// ```
type XChainModifyBridge struct {
	BaseTx

	// Specifies the flags for this transaction.
	Flags uint32
	// The minimum amount, in XRP, required for a XChainAccountCreateCommit transaction.
	// If this is not present, the XChainAccountCreateCommit transaction will fail.
	// This field can only be present on XRP-XRP bridges.
	MinAccountCreateAmount types.CurrencyAmount `json:",omitempty"`
	// The signature reward split between the witnesses for submitting attestations.
	SignatureReward types.CurrencyAmount `json:",omitempty"`
	// The bridge to modify.
	XChainBridge types.XChainBridge
}

// TxType returns the transaction type identifier for XChainModifyBridge.
func (x *XChainModifyBridge) TxType() TxType {
	return XChainModifyBridgeTx
}

// SetClearAccountCreateAmount enables the flag to clear the minimum account create amount.
func (x *XChainModifyBridge) SetClearAccountCreateAmount() {
	x.Flags |= TfClearAccountCreateAmount
}

// Flatten returns a flat map representation of the XChainModifyBridge transaction.
func (x *XChainModifyBridge) Flatten() FlatTransaction {
	flatTx := x.BaseTx.Flatten()

	flatTx["TransactionType"] = x.TxType().String()

	if x.Flags != 0 {
		flatTx["Flags"] = x.Flags
	}

	if x.MinAccountCreateAmount != nil {
		flatTx["MinAccountCreateAmount"] = x.MinAccountCreateAmount.Flatten()
	}

	if x.SignatureReward != nil {
		flatTx["SignatureReward"] = x.SignatureReward.Flatten()
	}

	if x.XChainBridge != (types.XChainBridge{}) {
		flatTx["XChainBridge"] = x.XChainBridge.Flatten()
	}

	return flatTx
}

// Validate checks the XChainModifyBridge fields for correctness and returns an error if invalid.
func (x *XChainModifyBridge) Validate() (bool, error) {
	_, err := x.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if !flag.Contains(x.Flags, TfClearAccountCreateAmount) {
		return false, ErrInvalidFlags
	}

	if ok, err := IsAmount(x.MinAccountCreateAmount, "MinAccountCreateAmount", false); !ok {
		return false, err
	}

	if ok, err := IsAmount(x.SignatureReward, "SignatureReward", false); !ok {
		return false, err
	}

	return x.XChainBridge.Validate()
}
