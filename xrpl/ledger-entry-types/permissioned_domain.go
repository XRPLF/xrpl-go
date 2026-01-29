package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// PermissionedDomain represents a ledger entry that describes a single permissioned domain instance.
// You can create a permissioned domain by sending a PermissionedDomainSet transaction.
//
// ```json
//
//	{
//		"LedgerEntryType": "PermissionedDomain",
//		"Fee": "10",
//		"Flags": 0,
//		"Owner": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
//		"OwnerNode": "0000000000000000",
//		"Sequence": 390,
//		"AcceptedCredentials": [
//		  {
//			  "Credential": {
//				  "Issuer": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
//				  "CredentialType": "6D795F63726564656E7469616C"
//			  }
//		  }
//		],
//		"PreviousTxnID": "E7E3F2BBAAF48CF893896E48DC4A02BDA0C747B198D5AE18BC3D7567EE64B904",
//		"PreviousTxnLgrSeq": 8734523,
//		"index": "3DFA1DDEA27AF7E466DE395CCB16158E07ECA6BC4EB5580F75EBD39DE833645F"
//	  }
//
// ```
type PermissionedDomain struct {
	// The unique ID for this ledger entry.
	// In JSON, this field is represented with different names depending on the context and API method.
	// (Note, even though this is specified as "optional" in the code, every ledger entry should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The value 0x0082, mapped to the string PermissionedDomain, indicates that this is a PermissionedDomain ledger entry.
	LedgerEntryType EntryType
	// Fee is the transaction cost, in drops of XRP, that was paid by the
	// PermissionedDomainSet transaction which created or most recently modified
	// this PermissionedDomain ledger entry.
	Fee types.XRPCurrencyAmount
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// 	The address of the account that owns this domain.
	Owner types.Address
	// A hint indicating which page of the owner directory links to this entry,
	// in case the directory consists of multiple pages.
	OwnerNode string
	// The Sequence value of the transaction that created this entry.
	Sequence uint32
	// A list of 1 to 10 Credential objects that grant access to this domain.
	// The array is stored sorted by issuer.
	AcceptedCredentials types.AuthorizeCredentialList
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this object.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the type of this ledger entry (PermissionedDomainEntry).
func (*PermissionedDomain) EntryType() EntryType {
	return PermissionedDomainEntry
}

// Flatten converts the PermissionedDomain struct into a flat map representation.
func (p *PermissionedDomain) Flatten() FlatLedgerObject {
	flattened := make(FlatLedgerObject)
	if p.Index.String() != "" {
		flattened["index"] = p.Index.String()
	}
	flattened["LedgerEntryType"] = p.LedgerEntryType
	flattened["Fee"] = p.Fee.Flatten()
	flattened["Flags"] = p.Flags
	flattened["Owner"] = p.Owner.String()
	flattened["OwnerNode"] = p.OwnerNode
	flattened["Sequence"] = p.Sequence
	flattened["AcceptedCredentials"] = p.AcceptedCredentials.Flatten()
	flattened["PreviousTxnID"] = p.PreviousTxnID.String()
	flattened["PreviousTxnLgrSeq"] = p.PreviousTxnLgrSeq

	return flattened
}
