package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TfClawTwoAssets claws back the specified amount of Asset, and a corresponding amount of Asset2 based on
	// the AMM pool's asset proportion; both assets must be issued by the issuer in the Account
	// field. If this flag isn't enabled, the issuer claws back the specified amount of Asset,
	// while a corresponding proportion of Asset2 goes back to the Holder.
	TfClawTwoAssets uint32 = 0x00000001
)

// AMMClawback claws back tokens from a holder who has deposited issued tokens into an AMM pool.
// To enable clawback, send an AccountSet transaction to allow trust line clawback. This setting
// cannot be reverted once the owner directory has any entries. (Added by the AMMClawback amendment)
//
// ```json
//
//	{
//	  "TransactionType": "AMMClawback",
//	  "Account": "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
//	  "Holder": "rvYAfWj5gh67oV6fW32ZzP3Aw4Eubs59B",
//	  "Asset": {
//	      "currency" : "FOO",
//	      "issuer" : "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL"
//	  },
//	  "Asset2" : {
//	      "currency" : "BAR",
//	      "issuer" : "rHtptZx1yHf6Yv43s1RWffM3XnEYv3XhRg"
//	  },
//	  "Amount": {
//	      "currency" : "FOO",
//	      "issuer" : "rPdYxU9dNkbzC5Y2h4jLbVJ3rMRrk7WVRL",
//	      "value" : "1000"
//	  }
//	}
//
// ```
type AMMClawback struct {
	BaseTx
	// The account holding the asset to be clawed back.
	Holder string
	// Specifies the asset that the issuer wants to claw back from the AMM pool.
	// The asset can be XRP, a token, or an MPT (see: Specifying Without Amounts).
	// The issuer field must match with Account.
	Asset types.IssuedCurrency
	// Specifies the other asset in the AMM's pool. The asset can be XRP, a token,
	// or an MPT (see: Specifying Without Amounts).
	Asset2 types.CurrencyAmount `json:",omitempty"`
	// The maximum amount to claw back from the AMM account. The currency and issuer subfields
	// should match the Asset subfields. If this field isn't specified, or the value subfield
	// exceeds the holder's available tokens in the AMM, all of the holder's tokens are clawed back.
	Amount types.IssuedCurrencyAmount `json:",omitempty"`
}

// TxType returns the transaction type for AMMClawback.
func (a *AMMClawback) TxType() TxType {
	return AMMClawbackTx
}

// Flatten returns the flattened representation of the AMMClawback transaction.
func (a *AMMClawback) Flatten() FlatTransaction {
	flattened := a.BaseTx.Flatten()
	flattened["TransactionType"] = a.TxType().String()

	if a.Holder != "" {
		flattened["Holder"] = a.Holder
	}

	if a.Asset != (types.IssuedCurrency{}) {
		flattened["Asset"] = a.Asset.Flatten()
	}

	if a.Asset2 != nil {
		flattened["Asset2"] = a.Asset2.Flatten()
	}

	if a.Amount != (types.IssuedCurrencyAmount{}) {
		flattened["Amount"] = a.Amount.Flatten()
	}

	return flattened
}

// Validate validates the AMMClawback transaction.
func (a *AMMClawback) Validate() (bool, error) {
	_, err := a.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if !addresscodec.IsValidAddress(a.Holder) {
		return false, ErrInvalidHolder
	}

	// Enforce that the issuer for Asset matches the Account if that is truly required.
	if a.Asset != (types.IssuedCurrency{}) && a.Asset.Issuer != a.Account {
		return false, ErrInvalidAssetIssuer
	}

	if a.Amount != (types.IssuedCurrencyAmount{}) {
		if !addresscodec.IsValidAddress(a.Amount.Issuer.String()) {
			return false, ErrInvalidAmountIssuer
		}
	}

	return true, nil
}

// SetClawTwoAssets sets the TfClawTwoAssets flag to claw back both assets based on the AMM pool proportions.
func (a *AMMClawback) SetClawTwoAssets() {
	a.Flags |= TfClawTwoAssets
}
