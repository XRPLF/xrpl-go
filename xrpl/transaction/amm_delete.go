package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
)

// AMMDelete deletes an empty Automated Market Maker (AMM) instance that could not be fully deleted automatically.
// Normally, an AMMWithdraw transaction automatically deletes an AMM and all associated ledger entries when it withdraws all the assets from the AMM's pool.
// However, if there are too many trust lines to the AMM account to remove in one transaction, it may stop before fully removing the AMM.
// Similarly, an AMMDelete transaction removes up to a maximum of 512 trust lines; it may take several AMMDelete transactions to delete all the trust lines and the associated AMM.
// In all cases, only the last such transaction deletes the AMM and AccountRoot ledger entries.
//
// Example:
//
// ```json
//
//	{
//	    "Account" : "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "Asset" : {
//	        "currency" : "XRP"
//	    },
//	    "Asset2" : {
//	        "currency" : "TST",
//	        "issuer" : "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
//	    },
//	    "Fee" : "10",
//	    "Flags" : 0,
//	    "Sequence" : 9,
//	    "TransactionType" : "AMMDelete"
//	}
//
// ```
type AMMDelete struct {
	BaseTx
	// The definition for one of the assets in the AMM's pool. In JSON, this is an object with currency and issuer fields (omit issuer for XRP).
	Asset ledger.Asset
	// The definition for the other asset in the AMM's pool. In JSON, this is an object with currency and issuer fields (omit issuer for XRP).
	Asset2 ledger.Asset
}

// TxType returns the type of the transaction (AMMDelete).
func (*AMMDelete) TxType() TxType {
	return AMMDeleteTx
}

// Flatten returns a map of the AMMDelete struct
func (a *AMMDelete) Flatten() FlatTransaction {
	// Add BaseTx fields
	flattened := a.BaseTx.Flatten()

	// Add AMMDelete-specific fields
	flattened["TransactionType"] = "AMMDelete"

	flattened["Asset"] = a.Asset.Flatten()

	flattened["Asset2"] = a.Asset2.Flatten()
	return flattened
}

// Validate checks the AMMDelete transaction fields for correctness.
func (a *AMMDelete) Validate() (bool, error) {
	_, err := a.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if ok, err := IsAsset(a.Asset); !ok {
		return false, err
	}

	if ok, err := IsAsset(a.Asset2); !ok {
		return false, err
	}

	return true, nil
}
