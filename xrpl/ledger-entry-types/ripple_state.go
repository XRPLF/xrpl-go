package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

const (
	// LsfAMMNode this entry consumed AMM liquidity to complete a Payment transaction.
	LsfAMMNode uint32 = 0x01000000
	// LsfLowReserve this entry contributes to the low account's owner reserve.
	LsfLowReserve uint32 = 0x00010000
	// LsfHighReserve this entry contributes to the high account's owner reserve.
	LsfHighReserve uint32 = 0x00020000
	// LsfLowAuth the low account has authorized the high account to hold tokens issued by the low account.
	LsfLowAuth uint32 = 0x00040000
	// LsfHighAuth the high account has authorized the low account to hold tokens issued by the high account.
	LsfHighAuth uint32 = 0x00080000
	// LsfLowNoRipple the low account has disabled rippling from this trust line.
	LsfLowNoRipple uint32 = 0x00100000
	// LsfHighNoRipple the high account has disabled rippling from this trust line.
	LsfHighNoRipple uint32 = 0x00200000
	// LsfLowFreeze the low account has frozen the trust line, preventing the high account from transferring the asset.
	LsfLowFreeze uint32 = 0x00400000
	// LsfHighFreeze the high account has frozen the trust line, preventing the low account from transferring the asset.
	LsfHighFreeze uint32 = 0x00800000

	// LsfLowDeepFreeze (XLS-77d Deep freeze) the low account has deep-frozen the trust line, preventing the high account from sending and
	// receiving the asset.
	LsfLowDeepFreeze uint32 = 0x02000000
	// LsfHighDeepFreeze the high account has deep-frozen the trust line, preventing the low account from sending and
	// receiving the asset.
	LsfHighDeepFreeze uint32 = 0x04000000
)

// RippleState ledger entry represents a trust line between two accounts.
// Each account can change its own limit and other settings, but the balance
// is a single shared value. A trust line that is entirely in its default state
// is considered the same as a trust line that does not exist and is automatically deleted.
//
// ```json
//
//	{
//	    "Balance": {
//	        "currency": "USD",
//	        "issuer": "rrrrrrrrrrrrrrrrrrrrBZbvji",
//	        "value": "-10"
//	    },
//	    "Flags": 393216,
//	    "HighLimit": {
//	        "currency": "USD",
//	        "issuer": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	        "value": "110"
//	    },
//	    "HighNode": "0000000000000000",
//	    "LedgerEntryType": "RippleState",
//	    "LowLimit": {
//	        "currency": "USD",
//	        "issuer": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
//	        "value": "0"
//	    },
//	    "LowNode": "0000000000000000",
//	    "PreviousTxnID": "E3FE6EA3D48F0C2B639448020EA4F03D4F4F8FFDB243A852A0F59177921B4879",
//	    "PreviousTxnLgrSeq": 14090896,
//	    "index": "9CA88CDEDFF9252B3DE183CE35B038F57282BC9503CDFA1923EF9A95DF0D6F7B"
//	}
//
//	{
type RippleState struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The balance of the trust line, from the perspective of the low account.
	// A negative balance indicates that the high account holds tokens issued by the low account.
	// The issuer in this is always set to the neutral value ACCOUNT_ONE.
	Balance types.IssuedCurrencyAmount
	// A bit-map of boolean options enabled for this entry.
	Flags uint32
	// The limit that the high account has set on the trust line. The issuer is the address of
	// the high account that set this limit.
	HighLimit types.IssuedCurrencyAmount
	// (Omitted in some historical ledgers) A hint indicating which page of the high account's
	// owner directory links to this entry, in case the directory consists of multiple pages.
	HighNode string
	// The inbound quality set by the high account, as an integer in the implied ratio
	// HighQualityIn:1,000,000,000. As a special case, the value 0 is equivalent to 1 billion, or face value.
	HighQualityIn uint32 `json:",omitempty"`
	// The outbound quality set by the high account, as an integer in the implied ratio
	// HighQualityOut:1,000,000,000. As a special case, the value 0 is equivalent to 1 billion, or face value.
	HighQualityOut uint32 `json:",omitempty"`
	// The value 0x0072, mapped to the string RippleState, indicates that this is a RippleState entry.
	LedgerEntryType EntryType
	// The limit that the low account has set on the trust line. The issuer is the address of the low account that set this limit.
	LowLimit types.IssuedCurrencyAmount
	// (Omitted in some historical ledgers) A hint indicating which page of the low account's owner directory links to this entry,
	// in case the directory consists of multiple pages.
	LowNode string
	// The inbound quality set by the low account, as an integer in the implied ratio LowQualityIn:1,000,000,000.
	// As a special case, the value 0 is equivalent to 1 billion, or face value.
	LowQualityIn uint32 `json:",omitempty"`
	// The outbound quality set by the low account, as an integer in the implied ratio LowQualityOut:1,000,000,000.
	// As a special case, the value 0 is equivalent to 1 billion, or face value.
	LowQualityOut uint32 `json:",omitempty"`
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this entry.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the type of the ledger entry.
func (*RippleState) EntryType() EntryType {
	return RippleStateEntry
}

// SetLsfAMMNode sets the AMM node flag.
func (r *RippleState) SetLsfAMMNode() {
	r.Flags |= LsfAMMNode
}

// SetLsfLowReserve sets the low reserve flag.
func (r *RippleState) SetLsfLowReserve() {
	r.Flags |= LsfLowReserve
}

// SetLsfHighReserve sets the high reserve flag.
func (r *RippleState) SetLsfHighReserve() {
	r.Flags |= LsfHighReserve
}

// SetLsfLowAuth sets the low auth flag.
func (r *RippleState) SetLsfLowAuth() {
	r.Flags |= LsfLowAuth
}

// SetLsfHighAuth sets the high auth flag.
func (r *RippleState) SetLsfHighAuth() {
	r.Flags |= LsfHighAuth
}

// SetLsfLowNoRipple sets the low no ripple flag.
func (r *RippleState) SetLsfLowNoRipple() {
	r.Flags |= LsfLowNoRipple
}

// SetLsfHighNoRipple sets the high no ripple flag.
func (r *RippleState) SetLsfHighNoRipple() {
	r.Flags |= LsfHighNoRipple
}

// SetLsfLowFreeze sets the low freeze flag.
func (r *RippleState) SetLsfLowFreeze() {
	r.Flags |= LsfLowFreeze
}

// SetLsfHighFreeze sets the high freeze flag.
func (r *RippleState) SetLsfHighFreeze() {
	r.Flags |= LsfHighFreeze
}

// SetLsfLowDeepFreeze sets the low deep freeze flag.
func (r *RippleState) SetLsfLowDeepFreeze() {
	r.Flags |= LsfLowDeepFreeze
}

// SetLsfHighDeepFreeze sets the high deep freeze flag.
func (r *RippleState) SetLsfHighDeepFreeze() {
	r.Flags |= LsfHighDeepFreeze
}
