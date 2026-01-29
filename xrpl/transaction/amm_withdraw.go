package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// AMMWithdraw withdraws assets from an Automated Market Maker (AMM) by redeeming liquidity provider tokens (LP Tokens).
//
// # Example
//
// ```json
//
//	{
//	    "Account" : "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "Amount" : {
//	        "currency" : "TST",
//	        "issuer" : "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
//	        "value" : "5"
//	    },
//	    "Amount2" : "50000000",
//	    "Asset" : {
//	        "currency" : "TST",
//	        "issuer" : "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
//	    },
//	    "Asset2" : {
//	        "currency" : "XRP"
//	    },
//	    "Fee" : "10",
//	    "Flags" : 1048576,
//	    "Sequence" : 10,
//	    "TransactionType" : "AMMWithdraw"
//	}
//
// / ```
type AMMWithdraw struct {
	BaseTx
	// The definition for one of the assets in the AMM's pool. In JSON, this is an object with currency and issuer fields (omit issuer for XRP).
	Asset ledger.Asset
	// The definition for the other asset in the AMM's pool. In JSON, this is an object with currency and issuer fields (omit issuer for XRP).
	Asset2 ledger.Asset
	// The amount of one asset to withdraw from the AMM. This must match the type of one of the assets (tokens or XRP) in the AMM's pool.
	Amount types.CurrencyAmount `json:",omitempty"`
	// The amount of another asset to withdraw from the AMM. If present, this must match the type of the other asset in the AMM's pool and cannot be the same type as Amount.
	Amount2 types.CurrencyAmount `json:",omitempty"`
	// The minimum effective price, in LP Token returned, to pay per unit of the asset to withdraw.
	EPrice types.CurrencyAmount `json:",omitempty"`
	// How many of the AMM's LP Tokens to redeem.
	LPTokenIn types.IssuedCurrencyAmount `json:",omitempty"`
}

// ****************************
// AMMWithdraw Flags
// ****************************

const (
	// Perform a double-asset withdrawal returning all your LP Tokens.
	tfWithdrawAll uint32 = 131072
	// Perform a single-asset withdrawal returning all of your LP Tokens.
	tfOneAssetWithdrawAll uint32 = 262144
)

// SetLPTokentFlag performs a double-asset withdrawal and receives the specified amount of LP Tokens.
func (a *AMMWithdraw) SetLPTokentFlag() {
	a.Flags |= tfLPToken
}

// SetWithdrawAllFlag performs a double-asset withdrawal returning all your LP Tokens.
func (a *AMMWithdraw) SetWithdrawAllFlag() {
	a.Flags |= tfWithdrawAll
}

// SetOneAssetWithdrawAllFlag performs a single-asset withdrawal returning all of your LP Tokens.
func (a *AMMWithdraw) SetOneAssetWithdrawAllFlag() {
	a.Flags |= tfOneAssetWithdrawAll
}

// SetSingleAssetFlag performs a single-asset withdrawal with a specified amount of the asset to withdraw.
func (a *AMMWithdraw) SetSingleAssetFlag() {
	a.Flags |= tfSingleAsset
}

// SetTwoAssetFlag performs a double-asset withdrawal with specified amounts of both assets.
func (a *AMMWithdraw) SetTwoAssetFlag() {
	a.Flags |= tfTwoAsset
}

// SetOneAssetLPTokenFlag performs a single-asset withdrawal and receives the specified amount of LP Tokens.
func (a *AMMWithdraw) SetOneAssetLPTokenFlag() {
	a.Flags |= tfOneAssetLPToken
}

// SetLimitLPTokenFlag performs a single-asset withdrawal with a specified effective price.
func (a *AMMWithdraw) SetLimitLPTokenFlag() {
	a.Flags |= tfLimitLPToken
}

// TxType returns the type of the transaction (AMMWithdraw).
func (*AMMWithdraw) TxType() TxType {
	return AMMWithdrawTx
}

// Flatten returns the flattened map of the AMMWithdraw transaction.
func (a *AMMWithdraw) Flatten() FlatTransaction {
	// Add BaseTx fields
	flattened := a.BaseTx.Flatten()

	// Add AMMWithdraw-specific fields
	flattened["TransactionType"] = "AMMWithdraw"

	flattened["Asset"] = a.Asset.Flatten()
	flattened["Asset2"] = a.Asset2.Flatten()

	if a.Amount != nil {
		flattened["Amount"] = a.Amount.Flatten()
	}
	if a.Amount2 != nil {
		flattened["Amount2"] = a.Amount2.Flatten()
	}
	if a.EPrice != nil {
		flattened["EPrice"] = a.EPrice.Flatten()
	}
	if !a.LPTokenIn.IsZero() {
		flattened["LPTokenIn"] = a.LPTokenIn.Flatten()
	}

	return flattened
}

// Validate validates the AMMWithdraw struct and make sure all the fields are correct.
func (a *AMMWithdraw) Validate() (bool, error) {
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

	if a.Amount2 != nil && a.Amount == nil {
		return false, ErrAMMMustSetAmountWithAmount2
	} else if a.EPrice != nil && a.Amount == nil {
		return false, ErrAMMMustSetAmountWithEPrice
	}

	if ok, err := IsAmount(a.Amount, "Amount", false); !ok {
		return false, err
	}

	if ok, err := IsAmount(a.Amount2, "Amount2", false); !ok {
		return false, err
	}

	if ok, err := IsAmount(a.EPrice, "EPrice", false); !ok {
		return false, err
	}

	if ok, err := IsIssuedCurrency(a.LPTokenIn); !ok {
		return false, err
	}

	return true, nil
}
