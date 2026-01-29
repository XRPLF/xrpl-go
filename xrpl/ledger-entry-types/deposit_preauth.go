package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// DepositPreauthObj tracks a preauthorization entry from one account to another.
// DepositPreauth transactions create these entries.
//
// This has no effect on processing of transactions unless the account that provided
// the preauthorization requires Deposit Authorization. In that case, the account
// that was preauthorized can send payments and other transactions directly to the
// account that provided the preauthorization. Preauthorizations are one-directional,
// and have no effect on payments going the opposite direction.
//
// ```json
//
//	{
//	  "LedgerEntryType": "DepositPreauth",
//	  "Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//	  "Authorize": "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
//	  "Flags": 0,
//	  "OwnerNode": "0000000000000000",
//	  "PreviousTxnID": "3E8964D5A86B3CD6B9ECB33310D4E073D64C865A5B866200AD2B7E29F8326702",
//	  "PreviousTxnLgrSeq": 7,
//	  "index": "4A255038CC3ADCC1A9C91509279B59908251728D0DAADB248FFE297D0F7E068C"
//	}
//
// ```
type DepositPreauthObj struct {
	// The unique ID for this ledger entry.
	// In JSON, this field is represented with different names depending on the context and API method.
	// (Note, even though this is specified as "optional" in the code, every ledger entry should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// The value 0x0070, mapped to the string DepositPreauth, indicates that this is a DepositPreauth object.
	LedgerEntryType EntryType
	// The account that granted the preauthorization. (The destination of the preauthorized payments.)
	Account types.Address
	// The account that received the preauthorization. (The sender of the preauthorized payments.)
	Authorize types.Address
	// A hint indicating which page of the sender's owner directory links to this object, in case the directory
	// consists of multiple pages. Note: The object does not contain a direct link to the owner directory
	// containing it, since that value can be derived from the Account.
	OwnerNode string
	// The identifying hash of the transaction that most recently modified this object.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this object.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the ledger entry type for DepositPreauthObj.
func (*DepositPreauthObj) EntryType() EntryType {
	return DepositPreauthObjEntry
}
