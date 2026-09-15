package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// Delegate entry type represents a set of permissions that an account has delegated to another account.
// This allows one account to authorize another account to perform specific transactions on its behalf.
// (Requires the AccountPermissionDelegation amendment)
//
// ```json
//
//	{
//	  "LedgerEntryType": "Delegate",
//	  "Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
//	  "Authorize": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
//	  "Permissions": ["Payment", "TrustlineAuthorize"],
//	  "OwnerNode": "0000000000000000",
//	  "Flags": 0,
//	  "PreviousTxnID": "F19AD4577212D3BEACA0F75FE1BA1644F2E854D46E8D62E9C95D18E9708CBFB1",
//	  "PreviousTxnLgrSeq": 4
//	}
//
// ```
type Delegate struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The type of ledger entry. Always "Delegate" for this entry type.
	LedgerEntryType EntryType
	// A bit-map of boolean flags. No flags are defined for the Delegate object type, so this value is always 0.
	Flags uint32
	// The account that wants to authorize another account.
	Account types.Address
	// The authorized account.
	Authorize types.Address
	// The transaction permissions that the authorized account has access to.
	Permissions []types.Permission
	// A hint indicating which page of the sender's owner directory links to this object,
	// in case the directory consists of multiple pages.
	OwnerNode string
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this entry.
	PreviousTxnLgrSeq uint32
	// The account paying this object's owner reserve. Requires the Sponsor amendment.
	Sponsor types.Address `json:",omitempty"`
}

// EntryType returns the type of the ledger entry.
func (*Delegate) EntryType() EntryType {
	return DelegateEntry
}
