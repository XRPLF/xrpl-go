package transaction

import (
	"encoding/json"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// EscrowCreate sequesters XRP until the escrow process either finishes or is canceled.
//
// Example:
//
// ```json
//
//	{
//	    "Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "TransactionType": "EscrowCreate",
//	    "Amount": "10000",
//	    "Destination": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
//	    "CancelAfter": 533257958,
//	    "FinishAfter": 533171558,
//	    "Condition": "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
//	    "DestinationTag": 23480,
//	    "SourceTag": 11747
//	}
//
// ```
type EscrowCreate struct {
	BaseTx
	// Amount of XRP or fungible tokens to deduct from the sender's balance and escrow.
	// Once escrowed, the payment can either go to the Destination address (after the FinishAfter time) or be returned to the sender (after the CancelAfter time).
	Amount types.CurrencyAmount
	// Address to receive escrowed XRP.
	Destination types.Address
	// (Optional) The time, in seconds since the Ripple Epoch, when this escrow expires. This value is immutable; the funds can only be returned to the sender after this time.
	CancelAfter uint32 `json:",omitempty"`
	// (Optional) The time, in seconds since the Ripple Epoch, when the escrowed XRP can be released to the recipient. This value is immutable, and the funds can't be accessed until this time.
	FinishAfter uint32 `json:",omitempty"`
	// (Optional) Hex value representing a PREIMAGE-SHA-256 crypto-condition. The funds can only be delivered to the recipient if this condition is fulfilled. If the condition is not fulfilled before the expiration time specified in the CancelAfter field, the XRP can only revert to the sender.
	Condition string `json:",omitempty"`
	// (Optional) Arbitrary tag to further specify the destination for this escrowed payment, such as a hosted recipient at the destination address.
	DestinationTag *uint32 `json:",omitempty"`
}

// TxType returns the transaction type for this transaction (EscrowCreate).
func (*EscrowCreate) TxType() TxType {
	return EscrowCreateTx
}

// Flatten returns the flattened map of the EscrowCreate transaction.
func (e *EscrowCreate) Flatten() FlatTransaction {
	flattened := e.BaseTx.Flatten()

	flattened["TransactionType"] = "EscrowCreate"

	flattened["Amount"] = e.Amount.Flatten()

	if e.Destination != "" {
		flattened["Destination"] = e.Destination.String()
	}
	if e.CancelAfter != 0 {
		flattened["CancelAfter"] = e.CancelAfter
	}
	if e.FinishAfter != 0 {
		flattened["FinishAfter"] = e.FinishAfter
	}
	if e.Condition != "" {
		flattened["Condition"] = e.Condition
	}
	if e.DestinationTag != nil {
		flattened["DestinationTag"] = *e.DestinationTag
	}

	return flattened
}

// UnmarshalJSON implements custom JSON unmarshalling for EscrowCreate.
func (e *EscrowCreate) UnmarshalJSON(data []byte) error {
	type escrowCreateHelper struct {
		BaseTx
		Amount         json.RawMessage
		Destination    types.Address
		CancelAfter    uint32  `json:",omitempty"`
		FinishAfter    uint32  `json:",omitempty"`
		Condition      string  `json:",omitempty"`
		DestinationTag *uint32 `json:",omitempty"`
	}
	var h escrowCreateHelper
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}
	*e = EscrowCreate{
		BaseTx:         h.BaseTx,
		Destination:    h.Destination,
		CancelAfter:    h.CancelAfter,
		FinishAfter:    h.FinishAfter,
		Condition:      h.Condition,
		DestinationTag: h.DestinationTag,
	}
	amount, err := types.UnmarshalCurrencyAmount(h.Amount)
	if err != nil {
		return err
	}
	e.Amount = amount
	return nil
}

// Validate checks the EscrowCreate transaction fields for correctness.
func (e *EscrowCreate) Validate() (bool, error) {
	ok, err := e.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if !addresscodec.IsValidAddress(e.Destination.String()) {
		return false, ErrEscrowCreateInvalidDestinationAddress
	}

	if (e.FinishAfter == 0 && e.CancelAfter == 0) || (e.Condition == "" && e.FinishAfter == 0) {
		return false, ErrEscrowCreateNoConditionOrFinishAfterSet
	}

	return true, nil
}
