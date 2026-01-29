package transaction

import (
	"time"

	rippletime "github.com/Peersyst/xrpl-go/xrpl/time"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// PaymentChannelFund adds additional XRP to an open payment channel, and optionally updates the expiration time of the channel. Only the source address of the channel can use this transaction.
//
// Example:
//
// ```json
//
//	{
//		"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//		"TransactionType": "PaymentChannelFund",
//		"Channel": "C1AE6DDDEEC05CF2978C0BAD6FE302948E9533691DC749DCDD3B9E5992CA6198",
//		"Amount": "200000",
//		"Expiration": 543171558
//	}
//
// ```
type PaymentChannelFund struct {
	BaseTx
	// The unique ID of the channel to fund, as a 64-character hexadecimal string.
	Channel types.Hash256
	// Amount of XRP, in drops to add to the channel. Must be a positive amount of XRP.
	Amount types.XRPCurrencyAmount
	// (Optional) New Expiration time to set for the channel, in seconds since the Ripple Epoch.
	// This must be later than either the current time plus the SettleDelay of the channel, or the existing Expiration of the channel.
	// After the Expiration time, any transaction that would access the channel closes the channel without taking its normal action.
	// Any unspent XRP is returned to the source address when the channel closes. (Expiration is separate from the channel's immutable CancelAfter time.) For more information, see the PayChannel ledger object type.
	Expiration uint32 `json:",omitempty"`
}

// TxType returns the type of the transaction (PaymentChannelFund).
func (*PaymentChannelFund) TxType() TxType {
	return PaymentChannelFundTx
}

// Flatten returns a map of the PaymentChannelFund transaction fields.
func (p *PaymentChannelFund) Flatten() FlatTransaction {
	flattened := p.BaseTx.Flatten()

	flattened["Channel"] = p.Channel.String()
	flattened["Amount"] = p.Amount.String()

	if p.Expiration != 0 {
		flattened["Expiration"] = p.Expiration
	}

	return flattened
}

// Validate validates the PaymentChannelFund fields.
func (p *PaymentChannelFund) Validate() (bool, error) {
	ok, err := p.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// check the expiration time is in the future. /!\ Incomplete as the channel SettleDelay is not taken into account but it's already a good check.
	currentRippleTime := rippletime.UnixTimeToRippleTime(time.Now().Unix())
	if (p.Expiration != 0) && (int64(p.Expiration) < currentRippleTime) {
		return false, ErrInvalidExpiration
	}

	return true, nil
}
