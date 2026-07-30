package transaction

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	bctypes "github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// mptIssuanceIDHexLength is the length of an MPT issuance ID in hex characters (48).
const mptIssuanceIDHexLength = 2 * bctypes.MPTIssuanceIDByteLength

// Clawback reclaims tokens issued by the account. Requires the Clawback amendment.
// Before using, enable Allow Trust Line Clawback via AccountSet with an empty owner directory. Once enabled, clawback cannot be disabled.
type Clawback struct {
	// Base transaction fields
	BaseTx

	// Indicates the amount being clawed back. For issued currencies, the issuer sub-field identifies the token holder.
	// For MPTs, the mpt_issuance_id sub-field identifies the issuance and Holder identifies the token holder.
	// The quantity to claw back must be greater than zero. If it exceeds the current balance, the entire balance is clawed back.
	Amount types.CurrencyAmount
	// Holder is the token holder to claw back from when Amount is an MPT amount.
	// It is required for MPT clawbacks and must be omitted for issued-currency clawbacks.
	Holder types.Address `json:",omitempty"`
}

// TxType implements the TxType method for the Clawback struct.
func (*Clawback) TxType() TxType {
	return ClawbackTx
}

// Flatten implements the Flatten method for the Clawback struct.
func (c *Clawback) Flatten() FlatTransaction {
	flattened := c.BaseTx.Flatten()

	flattened["TransactionType"] = "Clawback"

	if c.Amount != nil {
		flattened["Amount"] = c.Amount.Flatten()
	}

	if c.Holder != "" {
		flattened["Holder"] = c.Holder.String()
	}

	return flattened
}

// UnmarshalJSON implements custom JSON unmarshalling for Clawback currency amounts.
func (c *Clawback) UnmarshalJSON(data []byte) error {
	type clawbackJSON struct {
		BaseTx
		Amount json.RawMessage
		Holder types.Address
	}

	var decoded clawbackJSON
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	amount, err := types.UnmarshalCurrencyAmount(decoded.Amount)
	if err != nil {
		return err
	}

	*c = Clawback{
		BaseTx: decoded.BaseTx,
		Amount: amount,
		Holder: decoded.Holder,
	}
	return nil
}

// Validate implements the Validate method for the Clawback struct.
func (c *Clawback) Validate() (bool, error) {
	// validate the base transaction
	_, err := c.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	// check if the field Amount is set
	if c.Amount == nil {
		return false, ErrClawbackMissingAmount
	}

	switch amount := c.Amount.(type) {
	case types.IssuedCurrencyAmount:
		if ok, _ := IsIssuedCurrency(amount); !ok || amount.IsZero() {
			return false, ErrClawbackInvalidAmount
		}
		if c.Holder != "" {
			return false, ErrClawbackHolderNotAllowed
		}
		if c.Account == amount.Issuer {
			return false, ErrClawbackSameAccount
		}
	case types.MPTCurrencyAmount:
		if len(amount.MPTIssuanceID) != mptIssuanceIDHexLength {
			return false, ErrClawbackInvalidAmount
		}
		if ok, _ := IsMPTCurrency(amount); !ok || amount.IsZero() {
			return false, ErrClawbackInvalidAmount
		}

		issuanceIDBytes, err := hex.DecodeString(amount.MPTIssuanceID)
		if err != nil {
			return false, ErrClawbackInvalidAmount
		}
		// The issuer AccountID occupies the trailing AccountAddressLength bytes of the issuance ID.
		issuanceIssuerID := issuanceIDBytes[len(issuanceIDBytes)-addresscodec.AccountAddressLength:]
		_, accountID, err := addresscodec.DecodeClassicAddressToAccountID(c.Account.String())
		if err != nil {
			var hasTag bool
			accountID, _, hasTag, _, err = addresscodec.DecodeXAddress(c.Account.String())
			if err != nil || hasTag {
				return false, ErrInvalidAccount
			}
		}
		if !bytes.Equal(issuanceIssuerID, accountID) {
			return false, ErrClawbackMPTIssuerMismatch
		}

		if c.Holder == "" {
			return false, ErrClawbackMissingHolder
		}
		if !addresscodec.IsValidAddress(c.Holder.String()) {
			return false, ErrClawbackInvalidHolder
		}
		if _, _, hasTag, _, err := addresscodec.DecodeXAddress(c.Holder.String()); err == nil && hasTag {
			return false, ErrClawbackHolderTagNotAllowed
		}
		if c.Account == c.Holder {
			return false, ErrClawbackSameHolder
		}
	default:
		// XRP and any other amount kind cannot be clawed back.
		return false, ErrClawbackInvalidAmount
	}

	return true, nil
}
