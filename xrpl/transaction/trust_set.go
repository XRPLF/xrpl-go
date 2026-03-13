package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TfSetAuth authorizes the other party to hold currency issued by this account. (No
	// effect unless using the AsfRequireAuth AccountSet flag.) Cannot be unset.
	TfSetAuth uint32 = 0x00010000
	// TfSetNoRipple enables the No Ripple flag, which blocks rippling between two trust lines.
	// of the same currency if this flag is enabled on both.
	TfSetNoRipple uint32 = 0x00020000
	// TfClearNoRipple disables the No Ripple flag, allowing rippling on this trust line.
	TfClearNoRipple uint32 = 0x00040000
	// TfSetFreeze freezes the trust line.
	TfSetFreeze uint32 = 0x00100000
	// TfClearFreeze unfreezes the trust line.
	TfClearFreeze uint32 = 0x00200000

	// TfSetDeepFreeze (XLS-77d Deep freeze) freezes the trust line, preventing the high account from sending and
	// receiving the asset. Allowed only if the trustline is already regularly
	// frozen, or if TfSetFreeze is set in the same transaction.
	TfSetDeepFreeze uint32 = 0x00400000
	// TfClearDeepFreeze unfreezes the trust line, allowing the high account to send and
	// receive the asset.
	TfClearDeepFreeze uint32 = 0x00800000
)

// TrustSet creates or modifies a trust line linking two accounts.
type TrustSet struct {
	// Base transaction fields
	BaseTx
	// Object defining the trust line to create or modify, in the format of a Currency Amount.
	LimitAmount types.CurrencyAmount
	// (Optional) Value incoming balances on this trust line at the ratio of this number per 1,000,000,000 units.
	// A value of 0 is shorthand for treating balances at face value. For example, if you set the value to 10,000,000, 1% of incoming funds remain with the sender.
	// If an account sends 100 currency, the sender retains 1 currency unit and the destination receives 99 units. This option is included for parity: in practice, you are much more likely to set a QualityOut value.
	// Note that this fee is separate and independent from token transfer fees.
	QualityIn uint32 `json:",omitempty"`
	// (Optional) Value outgoing balances on this trust line at the ratio of this number per 1,000,000,000 units.
	// A value of 0 is shorthand for treating balances at face value. For example, if you set the value to 10,000,000, 1% of outgoing funds would remain with the issuer.
	// If the sender sends 100 currency units, the issuer retains 1 currency unit and the destination receives 99 units. Note that this fee is separate and independent from token transfer fees.
	QualityOut uint32 `json:",omitempty"`
}

// TxType returns the type of the transaction (TrustSet).
func (*TrustSet) TxType() TxType {
	return TrustSetTx
}

// Flatten returns a flattened map of the TrustSet transaction.
func (t *TrustSet) Flatten() FlatTransaction {
	flattened := t.BaseTx.Flatten()

	flattened["TransactionType"] = "TrustSet"

	if t.LimitAmount != nil {
		flattened["LimitAmount"] = t.LimitAmount.Flatten()
	}
	if t.QualityIn != 0 {
		flattened["QualityIn"] = t.QualityIn
	}
	if t.QualityOut != 0 {
		flattened["QualityOut"] = t.QualityOut
	}

	return flattened
}

// SetSetAuthFlag sets the TfSetAuth flag, authorizing the other party to hold currency issued by this account. Cannot be unset.
func (t *TrustSet) SetSetAuthFlag() {
	t.Flags |= TfSetAuth
}

// SetSetNoRippleFlag sets the TfSetNoRipple flag, enabling the No Ripple feature on the trust line.
func (t *TrustSet) SetSetNoRippleFlag() {
	t.Flags |= TfSetNoRipple
}

// SetClearNoRippleFlag sets the TfClearNoRipple flag, disabling the No Ripple feature on the trust line.
func (t *TrustSet) SetClearNoRippleFlag() {
	t.Flags |= TfClearNoRipple
}

// SetSetFreezeFlag sets the TfSetFreeze flag to freeze the trust line.
func (t *TrustSet) SetSetFreezeFlag() {
	t.Flags |= TfSetFreeze
}

// SetClearFreezeFlag sets the TfClearFreeze flag to unfreeze the trust line.
func (t *TrustSet) SetClearFreezeFlag() {
	t.Flags |= TfClearFreeze
}

// SetSetDeepFreezeFlag sets the TfSetDeepFreeze flag to deep freeze the trust line (XLS-77d).
func (t *TrustSet) SetSetDeepFreezeFlag() {
	t.Flags |= TfSetDeepFreeze
}

// SetClearDeepFreezeFlag sets the TfClearDeepFreeze flag to remove deep freeze on the trust line.
func (t *TrustSet) SetClearDeepFreezeFlag() {
	t.Flags |= TfClearDeepFreeze
}

// Validate checks that the TrustSet transaction has valid fields and flags.
func (t *TrustSet) Validate() (bool, error) {
	// Validate the base transaction
	_, err := t.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	// Check if the field LimitAmount is set
	if t.LimitAmount == nil {
		return false, ErrTrustSetMissingLimitAmount
	}

	if ok, err := IsAmount(t.LimitAmount, "LimitAmount", true); !ok {
		return false, err
	}

	return true, nil
}
