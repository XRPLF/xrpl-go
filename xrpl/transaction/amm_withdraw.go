package transaction

import (
	"math/bits"

	"github.com/Peersyst/xrpl-go/xrpl/flag"
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
	LPTokenIn types.IssuedCurrencyAmount `json:",omitzero"`
}

// ****************************
// AMMWithdraw Flags
// ****************************

const (
	// TfWithdrawAll performs a double-asset withdrawal returning all your LP Tokens.
	TfWithdrawAll uint32 = 131072
	// TfOneAssetWithdrawAll performs a single-asset withdrawal returning all of your LP Tokens.
	TfOneAssetWithdrawAll uint32 = 262144
)

const ammWithdrawModeFlags = TfLPToken | TfWithdrawAll | TfOneAssetWithdrawAll |
	TfSingleAsset | TfTwoAsset | TfOneAssetLPToken | TfLimitLPToken

// SetLPTokentFlag performs a double-asset withdrawal and receives the specified amount of LP Tokens.
func (a *AMMWithdraw) SetLPTokentFlag() {
	a.Flags |= TfLPToken
}

// SetWithdrawAllFlag performs a double-asset withdrawal returning all your LP Tokens.
func (a *AMMWithdraw) SetWithdrawAllFlag() {
	a.Flags |= TfWithdrawAll
}

// SetOneAssetWithdrawAllFlag performs a single-asset withdrawal returning all of your LP Tokens.
func (a *AMMWithdraw) SetOneAssetWithdrawAllFlag() {
	a.Flags |= TfOneAssetWithdrawAll
}

// SetSingleAssetFlag performs a single-asset withdrawal with a specified amount of the asset to withdraw.
func (a *AMMWithdraw) SetSingleAssetFlag() {
	a.Flags |= TfSingleAsset
}

// SetTwoAssetFlag performs a double-asset withdrawal with specified amounts of both assets.
func (a *AMMWithdraw) SetTwoAssetFlag() {
	a.Flags |= TfTwoAsset
}

// SetOneAssetLPTokenFlag performs a single-asset withdrawal and receives the specified amount of LP Tokens.
func (a *AMMWithdraw) SetOneAssetLPTokenFlag() {
	a.Flags |= TfOneAssetLPToken
}

// SetLimitLPTokenFlag performs a single-asset withdrawal with a specified effective price.
func (a *AMMWithdraw) SetLimitLPTokenFlag() {
	a.Flags |= TfLimitLPToken
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
	flattened["TransactionType"] = a.TxType().String()

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
	if !a.LPTokenIn.IsEmpty() {
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

	if !flag.ContainsOnly(a.Flags, ammWithdrawModeFlags|types.TfUniversal) ||
		bits.OnesCount32(a.Flags&ammWithdrawModeFlags) != 1 {
		return false, ErrInvalidFlags
	}

	if a.Amount2 != nil && a.Amount == nil {
		return false, ErrAMMMustSetAmountWithAmount2
	} else if a.EPrice != nil && a.Amount == nil {
		return false, ErrAMMMustSetAmountWithEPrice
	}

	mode := a.Flags & ammWithdrawModeFlags
	for _, field := range []struct {
		name     string
		present  bool
		required bool
	}{
		{"Amount", a.Amount != nil, mode != TfLPToken && mode != TfWithdrawAll},
		{"Amount2", a.Amount2 != nil, mode == TfTwoAsset},
		{"EPrice", a.EPrice != nil, mode == TfLimitLPToken},
		{"LPTokenIn", !a.LPTokenIn.IsEmpty(), mode == TfLPToken || mode == TfOneAssetLPToken},
	} {
		if field.required && !field.present {
			return false, ErrMissingField{Field: field.name}
		}
		if field.present && !field.required {
			return false, ErrTransactionInvalidField{Type: a.TxType().String(), Field: field.name}
		}
	}

	if sameAssetIssue(a.Asset, a.Asset2) {
		return false, ErrAMMInvalidTokens
	}

	if !a.LPTokenIn.IsEmpty() {
		if ok, err := IsIssuedCurrency(a.LPTokenIn); !ok {
			return false, err
		}
		if a.LPTokenIn.IsZero() {
			return false, ErrInvalidTokenValue
		}
	}
	if a.EPrice != nil && a.EPrice.Kind() == types.MPT {
		return false, ErrTransactionInvalidField{Type: a.TxType().String(), Field: "EPrice"}
	}

	for _, field := range []struct {
		amount    types.CurrencyAmount
		name      string
		validZero bool
		inPool    bool
	}{
		{a.Amount, "Amount", mode == TfOneAssetWithdrawAll || mode == TfOneAssetLPToken || a.EPrice != nil, true},
		{a.Amount2, "Amount2", false, true},
		{a.EPrice, "EPrice", false, false},
	} {
		if ok, err := IsAmount(field.amount, field.name, false); !ok {
			return false, err
		}
		if field.amount == nil {
			continue
		}
		if !field.validZero && field.amount.IsZero() {
			return false, ErrInvalidTokenValue
		}
		if field.inPool && !sameIssue(field.amount, a.Asset) && !sameIssue(field.amount, a.Asset2) {
			return false, ErrAMMInvalidTokens
		}
	}

	if a.Amount != nil && a.Amount2 != nil && sameIssue(a.Amount, a.Asset) == sameIssue(a.Amount2, a.Asset) {
		return false, ErrAMMInvalidTokens
	}

	if a.EPrice != nil && (a.EPrice.Kind() == types.XRP || sameIssue(a.EPrice, a.Asset) || sameIssue(a.EPrice, a.Asset2)) {
		return false, ErrAMMInvalidTokens
	}

	return true, nil
}
