package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// AMMCreate creates a new Automated Market Maker (AMM) instance for trading a pair of assets (fungible tokens or XRP).
//
// Creates both an AMM entry and a special AccountRoot entry to represent the AMM.
// Also transfers ownership of the starting balance of both assets from the sender to the created AccountRoot and issues an initial balance of liquidity provider tokens (LP Tokens) from the AMM account to the sender.
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
//	        "value" : "25"
//	    },
//	    "Amount2" : "250000000",
//	    "Fee" : "2000000",
//	    "Flags" : 2147483648,
//	    "Sequence" : 6,
//	    "TradingFee" : 500,
//	    "TransactionType" : "AMMCreate"
//	}
//
// ```
type AMMCreate struct {
	BaseTx
	// The first of the two assets to fund this AMM with. This must be a positive amount.
	Amount types.CurrencyAmount
	// The second of the two assets to fund this AMM with. This must be a positive amount.
	Amount2 types.CurrencyAmount
	// The fee to charge for trades against this AMM instance, in units of 1/100,000; a value of 1 is equivalent to 0.001%. The maximum value is 1000, indicating a 1% fee. The minimum value is 0.
	TradingFee uint16
}

// TxType returns the type of the transaction (AMMCreate).
func (*AMMCreate) TxType() TxType {
	return AMMCreateTx
}

// Flatten returns a map of the AMMCreate struct
func (a *AMMCreate) Flatten() FlatTransaction {
	// Add BaseTx fields
	flattened := a.BaseTx.Flatten()

	// Add AMMCreate-specific fields
	flattened["TransactionType"] = AMMCreateTx.String()

	if a.Amount != nil {
		flattened["Amount"] = a.Amount.Flatten()
	}

	if a.Amount2 != nil {
		flattened["Amount2"] = a.Amount2.Flatten()
	}

	flattened["TradingFee"] = a.TradingFee

	return flattened
}

// Validate validates the AMMCreate struct and ensures all fields are correct.
func (a *AMMCreate) Validate() (bool, error) {
	_, err := a.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if ok, err := IsAmount(a.Amount, "Amount", true); !ok {
		return false, err
	}

	if ok, err := IsAmount(a.Amount2, "Amount2", true); !ok {
		return false, err
	}

	if a.TradingFee > AmmMaxTradingFee {
		return false, ErrAMMTradingFeeTooHigh{
			Value: a.TradingFee,
			Limit: AmmMaxTradingFee,
		}
	}

	return true, nil
}
