package transaction

import (
	"bytes"
	"encoding/json"
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
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
	// The issued token or MPT that the issuer wants to claw back from the AMM pool.
	// Its issuer must match Account; XRP is not allowed here.
	Asset ledger.Asset
	// The other XRP, issued-token, or MPT asset in the AMM pool.
	Asset2 ledger.Asset
	// The optional maximum amount to claw back. It must be positive and identify Asset.
	// Omission claws back all of the holder's available Asset tokens in the AMM.
	Amount types.CurrencyAmount `json:",omitempty"`
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

	flattened["Asset"] = a.Asset.Flatten()
	flattened["Asset2"] = a.Asset2.Flatten()

	if a.Amount != nil {
		flattened["Amount"] = a.Amount.Flatten()
	}

	return flattened
}

// UnmarshalJSON implements custom JSON unmarshalling for the optional currency amount.
func (a *AMMClawback) UnmarshalJSON(data []byte) error {
	type ammClawbackFields AMMClawback
	var decoded struct {
		ammClawbackFields
		Amount json.RawMessage
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	amount, err := types.UnmarshalCurrencyAmount(decoded.Amount)
	if err != nil {
		return err
	}
	decoded.ammClawbackFields.Amount = amount
	*a = AMMClawback(decoded.ammClawbackFields)
	return nil
}

func validateAMMClawbackAmount(amount types.CurrencyAmount, asset ledger.Asset) error {
	amountKey, ok := amountIssueKey(amount)
	if !ok || !isPositiveTokenAmount(amount) {
		return ErrAMMClawbackInvalidAmount
	}
	if assetKey, _ := assetIssueKey(asset); !bytes.Equal(amountKey, assetKey) {
		return ErrAMMClawbackAmountAssetMismatch
	}

	return nil
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

	if ok, err := IsAsset(a.Asset); !ok {
		return false, fmt.Errorf("%w: %w", ErrAMMClawbackInvalidAsset, err)
	}
	if a.Asset.Kind() == ledger.AssetXRP {
		return false, ErrAMMClawbackAssetCannotBeXRP
	}

	accountID, _, err := decodeAddressAccountID(a.Account)
	if err != nil {
		return false, ErrInvalidAccount
	}
	assetIssuer, ok := assetIssuerAccountID(a.Asset)
	if !ok || !bytes.Equal(assetIssuer, accountID) {
		return false, ErrInvalidAssetIssuer
	}

	if ok, err := IsAsset(a.Asset2); !ok {
		return false, fmt.Errorf("%w: %w", ErrAMMClawbackInvalidAsset2, err)
	}

	if a.Flags&TfClawTwoAssets != 0 {
		asset2Issuer, ok := assetIssuerAccountID(a.Asset2)
		if !ok || !bytes.Equal(asset2Issuer, accountID) {
			return false, ErrAMMClawbackAsset2IssuerMismatch
		}
	}

	if a.Amount != nil {
		if err := validateAMMClawbackAmount(a.Amount, a.Asset); err != nil {
			return false, err
		}
	}

	// AMM existence, balances, clawback permissions, and amendment availability require ledger state.
	return true, nil
}

// SetClawTwoAssets sets the TfClawTwoAssets flag to claw back both assets based on the AMM pool proportions.
func (a *AMMClawback) SetClawTwoAssets() {
	a.Flags |= TfClawTwoAssets
}
