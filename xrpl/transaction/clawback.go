package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// Clawback reclaims tokens issued by the account. Requires the Clawback amendment.
// Before using, enable Allow Trust Line Clawback via AccountSet with an empty owner directory. Once enabled, clawback cannot be disabled.
type Clawback struct {
	// Base transaction fields
	BaseTx

	// Indicates the amount being clawed back, as well as the counterparty from which the amount is being clawed back.
	// The quantity to claw back, in the value sub-field, must not be zero. If this is more than the current balance,
	// the transaction claws back the entire balance. The sub-field issuer within Amount represents the token holder's
	// account ID, rather than the issuer's.
	Amount types.CurrencyAmount
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

	return flattened
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

	// check if the Amount is a valid currency amount
	if ok, _ := IsIssuedCurrency(c.Amount); !ok {
		return false, ErrClawbackInvalidAmount
	}

	// check if Account is not the same as the issuer
	if c.Account.String() == c.Amount.Flatten().(map[string]any)["issuer"] {
		return false, ErrClawbackSameAccount
	}

	return true, nil
}
