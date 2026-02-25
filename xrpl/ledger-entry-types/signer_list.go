package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

const (
	// LsfOneOwnerCount if this flag is enabled, this SignerList counts as one item for purposes of the owner reserve
	// Otherwise, this list counts as N+2 items, where N is the number of signers it contains. This
	// flag is automatically enabled if you add or update a signer list after the MultiSignReserve amendment is enabled.
	LsfOneOwnerCount uint32 = 0x00010000
)

// SignerList entry represents a list of parties that, as a group, are authorized to sign a transaction in place of an individual account.
// You can create, replace, or remove a signer list using a SignerListSet transaction.
//
// Example:
//
// ```json
//
//	{
//	    "Flags": 0,
//	    "LedgerEntryType": "SignerList",
//	    "OwnerNode": "0000000000000000",
//	    "PreviousTxnID": "5904C0DC72C58A83AEFED2FFC5386356AA83FCA6A88C89D00646E51E687CDBE4",
//	    "PreviousTxnLgrSeq": 16061435,
//	    "SignerEntries": [
//	        {
//	            "SignerEntry": {
//	                "Account": "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
//	                "SignerWeight": 2
//	            }
//	        },
//	        {
//	            "SignerEntry": {
//	                "Account": "raKEEVSGnKSD9Zyvxu4z6Pqpm4ABH8FS6n",
//	                "SignerWeight": 1
//	            }
//	        },
//	        {
//	            "SignerEntry": {
//	                "Account": "rUpy3eEg8rqjqfUoLeBnZkscbKbFsKXC3v",
//	                "SignerWeight": 1
//	            }
//	        }
//	    ],
//	    "SignerListID": 0,
//	    "SignerQuorum": 3,
//	    "index": "A9C28A28B85CD533217F5C0A0C7767666B093FA58A0F2D80026FCC4CD932DDC7"
//	}
//
// ```
type SignerList struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The value 0x0053, mapped to the string SignerList, indicates that this is a SignerList ledger entry.
	LedgerEntryType EntryType
	// The identifying hash of the transaction that most recently modified this object.
	PreviousTxnID string
	// The index of the ledger that contains the transaction that most recently modified this object.
	PreviousTxnLgrSeq uint32
	// A hint indicating which page of the owner directory links to this object, in case the directory consists of multiple pages.
	OwnerNode string
	// An array of Signer Entry objects representing the parties who are part of this signer list.
	SignerEntries []SignerEntryWrapper
	// An ID for this signer list. Currently always set to 0. If a future amendment allows multiple signer lists for an account, this may change.
	SignerListID uint32
	// A target number for signer weights. To produce a valid signature for the owner of this SignerList, the signers must provide valid signatures whose weights sum to this value or more.
	SignerQuorum uint32
	Flags        uint32
}

// SignerEntryWrapper wraps a SignerEntry for inclusion in a SignerList ledger entry.
type SignerEntryWrapper struct {
	SignerEntry SignerEntry
}

// Flatten returns a map of the SignerEntryWrapper object
func (s *SignerEntryWrapper) Flatten() FlatLedgerObject {
	flattened := make(FlatLedgerObject)
	flattened["SignerEntry"] = s.SignerEntry.Flatten()
	return flattened
}

// SignerEntry describes a signer in a SignerList ledger entry, including account, weight, and optional locator.
type SignerEntry struct {
	// An XRP Ledger address whose signature contributes to the multi-signature. It does not need to be a funded address in the ledger.
	Account types.Address
	// The weight of a signature from this signer. A multi-signature is only valid if the sum weight of the signatures provided meets or exceeds the signer list's SignerQuorum value.
	SignerWeight uint16
	// (Optional) Arbitrary hexadecimal data. This can be used to identify the signer or for other, related purposes. (Added by the ExpandedSignerList amendment.)
	WalletLocator types.Hash256 `json:",omitempty"`
}

// Flatten returns a map of the SignerEntry object
func (s *SignerEntry) Flatten() map[string]interface{} {
	flattened := make(map[string]interface{})

	if s.Account != "" {
		flattened["Account"] = s.Account.String()
	}

	if s.SignerWeight != 0 {
		flattened["SignerWeight"] = int(s.SignerWeight)
	}

	if s.WalletLocator != "" {
		flattened["WalletLocator"] = s.WalletLocator.String()
	}

	return flattened
}

// EntryType returns the type of the ledger entry (SignerList)
func (*SignerList) EntryType() EntryType {
	return SignerListEntry
}

// SetLsfOneOwnerCount sets the one owner count flag.
func (s *SignerList) SetLsfOneOwnerCount() {
	s.Flags |= LsfOneOwnerCount
}
