package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// PayChannel represents a ledger entry for a payment channel (added by the PayChan amendment).
type PayChannel struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The source address that owns this payment channel. This comes from the sending
	// address of the transaction that created the channel.
	Account types.Address
	// Total XRP, in drops, that has been allocated to this channel. This includes
	// XRP that has been paid to the destination address. This is initially set by the
	// transaction that created the channel and can be increased if the source address
	// sends a PaymentChannelFund transaction.
	Amount types.XRPCurrencyAmount
	// Total XRP, in drops, already paid out by the channel. The difference between
	// this value and the Amount field is how much XRP can still be paid to the
	// destination address with PaymentChannelClaim transactions. If the channel closes,
	// the remaining difference is returned to the source address.
	Balance types.XRPCurrencyAmount
	// The immutable expiration time for this payment channel, in seconds since the Ripple Epoch.
	// This channel is expired if this value is present and smaller than the previous ledger's
	// close_time field. This is optionally set by the transaction that created the channel, and
	// cannot be changed.
	CancelAfter uint32 `json:",omitempty"`
	// The destination address for this payment channel. While the payment channel is open,
	// this address is the only one that can receive XRP from the channel. This comes from
	// the Destination field of the transaction that created the channel.
	Destination types.Address
	// An arbitrary tag to further specify the destination for this payment channel, such as a
	// hosted recipient at the destination address.
	DestinationTag uint32 `json:",omitempty"`
	// A hint indicating which page of the destination's owner directory links to this entry,
	// in case the directory consists of multiple pages. Omitted on payment channels created
	// before enabling the fixPayChanRecipientOwnerDir amendment.
	DestinationNode string `json:",omitempty"`
	// The mutable expiration time for this payment channel, in seconds since the Ripple Epoch.
	// The channel is expired if this value is present and smaller than the previous ledger's close_time field.
	// See Channel Expiration for more details.
	Expiration uint32 `json:",omitempty"`
	Flags      uint32
	// The value 0x0078, mapped to the string PayChannel, indicates that this is a payment channel entry.
	LedgerEntryType EntryType
	// A hint indicating which page of the source address's owner directory links to this entry,
	// in case the directory consists of multiple pages.
	OwnerNode string
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this entry.
	PreviousTxnLgrSeq uint32
	// Public key, in hexadecimal, of the key pair that can be used to sign claims against this channel.
	// This can be any valid secp256k1 or Ed25519 public key. This is set by the transaction that created
	// the channel and must match the public key used in claims against the channel.
	// The channel source address can also send XRP from this channel to the destination without signed claims.
	PublicKey string
	// 	Number of seconds the source address must wait to close the channel if it still has
	// any XRP in it. Smaller values mean that the destination address has less time to redeem any
	// outstanding claims after the source address requests to close the channel. Can be any value
	// that fits in a 32-bit unsigned integer (0 to 2^32-1). This is set by the transaction that creates the channel.
	SettleDelay uint32
	// An arbitrary tag to further specify the source for this payment channel, such as a hosted
	// recipient at the owner's address.
	SourceTag uint32 `json:",omitempty"`
	// The account paying this object's owner reserve. Requires the Sponsor amendment.
	Sponsor types.Address `json:",omitempty"`
}

// EntryType returns the type of the ledger entry.
func (*PayChannel) EntryType() EntryType {
	return PayChannelEntry
}
