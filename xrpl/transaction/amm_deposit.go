package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// AMMDeposit deposits funds into an Automated Market Maker (AMM) and receives liquidity provider tokens (LP Tokens) in exchange.
// You can deposit one or both assets in the AMM's pool. A trust line to hold LP Tokens is created upon success.
//
// Example:
//
// ```json
//
//	{
//	    "Account" : "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "Amount" : {
//	        "currency" : "TST",
//	        "issuer" : "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
//	        "value" : "2.5"
//	    },
//	    "Amount2" : "30000000",
//	    "Asset" : {
//	        "currency" : "TST",
//	        "issuer" : "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
//	    },
//	    "Asset2" : {
//	        "currency" : "XRP"
//	    },
//	    "Fee" : "10",
//	    "Flags" : 1048576,
//	    "Sequence" : 7,
//	    "TransactionType" : "AMMDeposit"
//	}
//
// ```
type AMMDeposit struct {
	BaseTx
	// The definition for one of the assets in the AMM's pool. In JSON, this is an object with currency and issuer fields (omit issuer for XRP).
	Asset ledger.Asset
	// The definition for the other asset in the AMM's pool. In JSON, this is an object with currency and issuer fields (omit issuer for XRP).
	Asset2 ledger.Asset
	// The amount of one asset to deposit to the AMM. If present, this must match the type of one of the assets (tokens or XRP) in the AMM's pool.
	Amount types.CurrencyAmount `json:",omitempty"`
	// The amount of another asset to add to the AMM. If present, this must match the type of the other asset in the AMM's pool and cannot be the same asset as Amount.
	Amount2 types.CurrencyAmount `json:",omitempty"`
	// The maximum effective price, in the deposit asset, to pay for each LP Token received.
	EPrice types.CurrencyAmount `json:",omitempty"`
	// How many of the AMM's LP Tokens to buy.
	LPTokenOut types.CurrencyAmount `json:",omitempty"`
	// Submit a vote for the AMM's trading fee, in units of 1/100,000; a value of 1 is equivalent to 0.001%. The maximum value is 1000, indicating a 1% fee.
	TradingFee uint16 `json:",omitempty"`
}

// ****************************
// AMMDeposit Flags
// ****************************

// You must specify exactly one of these flags, plus any global flags.
const (
	// TfTwoAssetIfEmpty performs a special double-asset deposit to an AMM with an empty pool.
	TfTwoAssetIfEmpty uint32 = 8388608
)

// SetLPTokentFlag performs a double-asset deposit and receives the specified amount of LP Tokens.
func (a *AMMDeposit) SetLPTokentFlag() {
	a.Flags |= TfLPToken
}

// SetSingleAssetFlag performs a single-asset deposit with a specified asset amount.
func (a *AMMDeposit) SetSingleAssetFlag() {
	a.Flags |= TfSingleAsset
}

// SetTwoAssetFlag performs a double-asset deposit with specified amounts of both assets.
func (a *AMMDeposit) SetTwoAssetFlag() {
	a.Flags |= TfTwoAsset
}

// SetOneAssetLPTokenFlag performs a single-asset deposit and receives the specified amount of LP Tokens.
func (a *AMMDeposit) SetOneAssetLPTokenFlag() {
	a.Flags |= TfOneAssetLPToken
}

// SetLimitLPTokenFlag performs a single-asset deposit with a specified effective price.
func (a *AMMDeposit) SetLimitLPTokenFlag() {
	a.Flags |= TfLimitLPToken
}

// SetTwoAssetIfEmptyFlag performs a special double-asset deposit into an empty AMM pool.
func (a *AMMDeposit) SetTwoAssetIfEmptyFlag() {
	a.Flags |= TfTwoAssetIfEmpty
}

// TxType implements the TxType method for the AMMDeposit struct.
func (*AMMDeposit) TxType() TxType {
	return AMMDepositTx
}

// Flatten implements the Flatten method for the AMMDeposit struct.
func (a *AMMDeposit) Flatten() FlatTransaction {

	// Add BaseTx fields
	flattened := a.BaseTx.Flatten()

	// Add AMMDeposit-specific fields
	flattened["TransactionType"] = AMMDepositTx.String()

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

	if a.LPTokenOut != nil {
		flattened["LPTokenOut"] = a.LPTokenOut.Flatten()
	}

	if a.TradingFee != 0 {
		flattened["TradingFee"] = a.TradingFee
	}

	return flattened
}

// Validate implements the Validate method for the AMMDeposit struct.
func (a *AMMDeposit) Validate() (bool, error) {
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

	switch {
	case a.Amount2 != nil && a.Amount == nil:
		return false, ErrAMMMustSetAmountWithAmount2
	case a.EPrice != nil && a.Amount == nil:
		return false, ErrAMMMustSetAmountWithEPrice
	case a.LPTokenOut == nil && a.Amount == nil:
		return false, ErrAMMAtLeastOneAssetMustBeSet
	}

	if ok, err := IsAmount(a.Amount, "Amount", false); !ok {
		return false, err
	}

	if ok, err := IsAmount(a.Amount2, "Amount", false); !ok {
		return false, err
	}

	if ok, err := IsAmount(a.EPrice, "EPrice", false); !ok {
		return false, err
	}

	return true, nil
}
