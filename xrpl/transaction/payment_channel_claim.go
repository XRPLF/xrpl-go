package transaction

import (
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TfRenew clears the channel's Expiration time. (Expiration is different from the
	// channel's immutable CancelAfter time.) Only the source address of the
	// payment channel can use this flag.
	TfRenew uint32 = 65536 // 0x00010000
	// TfClose requests to close the channel. Only the channel source and destination
	// addresses can use this flag. This flag closes the channel immediately if it
	// has no more XRP allocated to it after processing the current claim, or if
	// the destination address uses it. If the source address uses this flag when
	// the channel still holds XRP, this schedules the channel to close after
	// SettleDelay seconds have passed. (Specifically, this sets the Expiration of
	// the channel to the close time of the previous ledger plus the channel's
	// SettleDelay time, unless the channel already has an earlier Expiration
	// time.) If the destination address uses this flag when the channel still
	// holds XRP, any XRP that remains after processing the claim is returned to
	// the source address.
	TfClose uint32 = 131072 // 0x00020000
)

// PaymentChannelClaim claims XRP from a payment channel, adjusts the payment channel's expiration, or both.
// This transaction can be used differently depending on the transaction sender's role in the specified channel:
//
// The source address of a channel can:
//
// - Send XRP from the channel to the destination with or without a signed Claim.
// - Set the channel to expire as soon as the channel's SettleDelay has passed.
// - Clear a pending Expiration time.
// - Close a channel immediately, with or without processing a claim first. The source address cannot close the channel immediately if the channel has XRP remaining.
//
// The destination address of a channel can:
//
// - Receive XRP from the channel using a signed Claim.
// - Close the channel immediately after processing a Claim, refunding any unclaimed XRP to the channel's source.
//
// Any address sending this transaction can:
//
// - Cause a channel to be closed if its Expiration or CancelAfter time is older than the previous ledger's close time. Any validly-formed PaymentChannelClaim transaction has this effect regardless of the contents of the transaction.
//
// Example:
//
// ```json
//
//	{
//		"Channel": "C1AE6DDDEEC05CF2978C0BAD6FE302948E9533691DC749DCDD3B9E5992CA6198",
//		"Balance": "1000000",
//		"Amount": "1000000",
//		"Signature": "30440220718D264EF05CAED7C781FF6DE298DCAC68D002562C9BF3A07C1E721B420C0DAB02203A5A4779EF4D2CCC7BC3EF886676D803A9981B928D3B8ACA483B80ECA3CD7B9B",
//		"PublicKey": "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A"
//	  }
//
// ```
type PaymentChannelClaim struct {
	BaseTx
	// The unique ID of the channel, as a 64-character hexadecimal string.
	Channel types.Hash256
	// Set of Credentials to authorize a deposit made by this transaction.
	// Each member of the array must be the ledger entry ID of a Credential entry in the ledger.
	// For details see https://xrpl.org/docs/references/protocol/transactions/types/payment#credential-ids
	CredentialIDs types.CredentialIDs `json:",omitempty"`
	// (Optional) Total amount of XRP, in drops, delivered by this channel after processing this claim. Required to deliver XRP.
	// Must be more than the total amount delivered by the channel so far, but not greater than the Amount of the signed claim. Must be provided except when closing the channel.
	Balance types.XRPCurrencyAmount `json:",omitempty"`
	// (Optional) The amount of XRP, in drops, authorized by the Signature. This must match the amount in the signed message.
	// This is the cumulative amount of XRP that can be dispensed by the channel, including XRP previously redeemed. Must be provided except when closing the channel.
	Amount types.XRPCurrencyAmount `json:",omitempty"`
	// (Optional) The signature of this claim, as hexadecimal. The signed message contains the channel ID and the amount of the claim.
	// Required unless the sender of the transaction is the source address of the channel.
	Signature string `json:",omitempty"`
	// (Optional) The public key used for the signature, as hexadecimal. This must match the PublicKey stored in the ledger for the channel.
	// Required unless the sender of the transaction is the source address of the channel and the Signature field is omitted.
	// (The transaction includes the public key so that rippled can check the validity of the signature before trying to apply the transaction to the ledger.)
	PublicKey string `json:",omitempty"`
}

// TxType returns the type of the transaction (PaymentChannelClaim).
func (*PaymentChannelClaim) TxType() TxType {
	return PaymentChannelClaimTx
}

// Flatten returns a flattened map of the PaymentChannelClaim transaction.
func (p *PaymentChannelClaim) Flatten() FlatTransaction {
	flattened := p.BaseTx.Flatten()

	flattened["TransactionType"] = "PaymentChannelClaim"

	if p.Channel != "" {
		flattened["Channel"] = p.Channel.String()
	}
	if p.Balance != 0 {
		flattened["Balance"] = p.Balance.Flatten()
	}
	if p.Amount != 0 {
		flattened["Amount"] = p.Amount.Flatten()
	}
	if p.Signature != "" {
		flattened["Signature"] = p.Signature
	}
	if p.PublicKey != "" {
		flattened["PublicKey"] = p.PublicKey
	}
	if len(p.CredentialIDs) > 0 {
		flattened["CredentialIDs"] = p.CredentialIDs.Flatten()
	}
	return flattened
}

// SetRenewFlag sets the Renew flag.
//
// Renew: Clear the channel's Expiration time. (Expiration is different from the
// channel's immutable CancelAfter time.) Only the source address of the
// payment channel can use this flag.
func (p *PaymentChannelClaim) SetRenewFlag() {
	p.Flags |= TfRenew
}

// SetCloseFlag sets the Close flag.
//
// Close: Request to close the channel. Only the channel source and destination
// addresses can use this flag. This flag closes the channel immediately if it
// has no more XRP allocated to it after processing the current claim, or if
// the destination address uses it. If the source address uses this flag when
// the channel still holds XRP, this schedules the channel to close after
// SettleDelay seconds have passed. (Specifically, this sets the Expiration of
// the channel to the close time of the previous ledger plus the channel's
// SettleDelay time, unless the channel already has an earlier Expiration
// time.) If the destination address uses this flag when the channel still
// holds XRP, any XRP that remains after processing the claim is returned to
// the source address.
func (p *PaymentChannelClaim) SetCloseFlag() {
	p.Flags |= TfClose
}

// Validate validates the PaymentChannelFund fields.
func (p *PaymentChannelClaim) Validate() (bool, error) {
	ok, err := p.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if p.Channel == "" {
		return false, ErrInvalidChannel
	}

	if p.Signature != "" && !typecheck.IsHex(p.Signature) {
		return false, ErrInvalidSignature
	}

	if p.PublicKey != "" && !typecheck.IsHex(p.PublicKey) {
		return false, ErrInvalidHexPublicKey
	}

	if p.CredentialIDs != nil && !p.CredentialIDs.IsValid() {
		return false, ErrInvalidCredentialIDs
	}

	return true, nil
}
