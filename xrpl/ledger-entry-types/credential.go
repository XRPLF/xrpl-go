package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

const (
	// LsfAccepted if enabled, the subject of the credential has accepted the credential.
	// Otherwise, the issuer created the credential but the subject has not yet accepted it, meaning it is not yet valid.
	LsfAccepted uint32 = 0x00010000
)

// Credential entry represents a credential, which contains an attestation about a subject account from a credential issuer account.
// The meaning of the attestation is defined by the issuer.
// Requires the Credentials amendment to be enabled: https://xrpl.org/resources/known-amendments#credentials
type Credential struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The value 0x0044, mapped to the string Credential, indicates that this is a Credential ledger entry.
	LedgerEntryType EntryType
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// Arbitrary data defining the type of credential this entry represents. The minimum length is 1 byte and the maximum length is 64 bytes.
	CredentialType types.CredentialType
	// Time after which the credential is expired, in seconds since the Ripple Epoch.
	Expiration uint32 `json:",omitempty"`
	// The account that issued the credential.
	Issuer types.Address
	// A hint indicating which page of the issuer's directory links to this entry, in case the directory consists of multiple pages.
	IssuerNode string
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this entry.
	PreviousTxnLgrSeq uint32
	// The account that this credential is for.
	Subject types.Address
	// A hint indicating which page of the subject's owner directory links to this entry, in case the directory consists of multiple pages.
	SubjectNode string
	// Arbitrary additional data about the credential, for example a URL where a W3C-formatted Verifiable Credential can be retrieved.
	URI string `json:",omitempty"`
}

// EntryType returns the type of the ledger entry.
func (*Credential) EntryType() EntryType {
	return CredentialEntry
}

// SetLsfAccepted sets the one owner count flag.
func (c *Credential) SetLsfAccepted() {
	c.Flags |= LsfAccepted
}
