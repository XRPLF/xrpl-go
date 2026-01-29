package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// PaymentChannelCreate creates a payment channel and funds it with XRP. The sender becomes the "source address" of the channel.
//
// Example:
//
// ```json
//
//	{
//	    "Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	    "TransactionType": "PaymentChannelCreate",
//	    "Amount": "10000",
//	    "Destination": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
//	    "SettleDelay": 86400,
//	    "PublicKey": "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
//	    "CancelAfter": 533171558,
//	    "DestinationTag": 23480,
//	    "SourceTag": 11747
//	}
//
// / ```
type PaymentChannelCreate struct {
	BaseTx
	// Amount of XRP, in drops, to deduct from the sender's balance and set aside in this channel.
	// While the channel is open, the XRP can only go to the Destination address. When the channel closes, any unclaimed XRP is returned to the source address's balance.
	Amount types.XRPCurrencyAmount
	// Address to receive XRP claims against this channel. This is also known as the "destination address" for the channel. Cannot be the same as the sender (Account).
	Destination types.Address
	// Amount of time the source address must wait before closing the channel if it has unclaimed XRP.
	SettleDelay uint32
	// The 33-byte public key of the key pair the source will use to sign claims against this channel, in hexadecimal.
	// This can be any secp256k1 or Ed25519 public key. For more information on key pairs, see Key Derivation
	PublicKey string
	// (Optional) The time, in seconds since the Ripple Epoch, when this channel expires.
	// Any transaction that would modify the channel after this time closes the channel without otherwise affecting it.
	// This value is immutable; the channel can be closed earlier than this time but cannot remain open after this time.
	CancelAfter uint32 `json:",omitempty"`
	// (Optional) Arbitrary tag to further specify the destination for this payment channel, such as a hosted recipient at the destination address.
	DestinationTag *uint32 `json:",omitempty"`
}

// TxType returns the type of the transaction (PaymentChannelCreate).
func (*PaymentChannelCreate) TxType() TxType {
	return PaymentChannelCreateTx
}

// Flatten returns a map of the PaymentChannelCreate transaction fields.
func (p *PaymentChannelCreate) Flatten() FlatTransaction {
	flattened := p.BaseTx.Flatten()

	flattened["Amount"] = p.Amount.String()
	flattened["Destination"] = p.Destination.String()
	flattened["SettleDelay"] = p.SettleDelay
	flattened["PublicKey"] = p.PublicKey

	if p.CancelAfter != 0 {
		flattened["CancelAfter"] = p.CancelAfter
	}

	if p.DestinationTag != nil {
		flattened["DestinationTag"] = *p.DestinationTag
	}

	return flattened
}

// Validate validates the PaymentChannelCreate fields.
func (p *PaymentChannelCreate) Validate() (bool, error) {
	ok, err := p.BaseTx.Validate()
	if (err != nil) || !ok {
		return false, err
	}

	// check valid xrpl address for Destination
	if !addresscodec.IsValidAddress(p.Destination.String()) {
		return false, ErrInvalidDestination
	}

	// check PublicKey is valid hexademical string
	if p.PublicKey == "" || !typecheck.IsHex(p.PublicKey) {
		return false, ErrInvalidHexPublicKey
	}

	return true, nil
}
