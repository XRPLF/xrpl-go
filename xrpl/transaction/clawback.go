package transaction

import (
	"bytes"
	"encoding/json"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

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

	flattened["TransactionType"] = c.TxType().String()

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
	type clawbackFields Clawback
	var decoded struct {
		clawbackFields
		Amount json.RawMessage
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	amount, err := types.UnmarshalCurrencyAmount(decoded.Amount)
	if err != nil {
		return err
	}

	decoded.clawbackFields.Amount = amount
	*c = Clawback(decoded.clawbackFields)
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

	accountID, accountHasTag, err := decodeAddressAccountID(c.Account)
	if err != nil {
		return false, ErrInvalidAccount
	}
	if accountHasTag && c.SourceTag != 0 {
		return false, ErrDuplicateXAddressTag
	}

	// XRP and any other non-token amount cannot be clawed back.
	if !isPositiveTokenAmount(c.Amount) {
		return false, ErrClawbackInvalidAmount
	}

	switch amount := c.Amount.(type) {
	case types.IssuedCurrencyAmount:
		if c.Holder != "" {
			return false, ErrClawbackHolderNotAllowed
		}

		issuerID, hasTag, err := decodeAddressAccountID(amount.Issuer)
		if err != nil || hasTag {
			return false, ErrClawbackInvalidAmount
		}
		if bytes.Equal(accountID, issuerID) {
			return false, ErrClawbackSameAccount
		}
	case types.MPTCurrencyAmount:
		issuanceIssuerID, ok := mptIssuerAccountID(amount.MPTIssuanceID)
		if !ok {
			return false, ErrClawbackInvalidAmount
		}
		if !bytes.Equal(issuanceIssuerID, accountID) {
			return false, ErrClawbackMPTIssuerMismatch
		}

		if c.Holder == "" {
			return false, ErrClawbackMissingHolder
		}
		holderID, hasTag, err := decodeAddressAccountID(c.Holder)
		if err != nil {
			return false, ErrClawbackInvalidHolder
		}
		if hasTag {
			return false, ErrClawbackHolderTagNotAllowed
		}
		if bytes.Equal(accountID, holderID) {
			return false, ErrClawbackSameHolder
		}
	}

	return true, nil
}
