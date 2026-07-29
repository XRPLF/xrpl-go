package ledger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ErrInvalidEntrySelector is returned when a string-or-object selector does not
// contain exactly one representation.
var ErrInvalidEntrySelector = errors.New("ledger_entry: selector must contain exactly one of index or object")

// EntrySelector represents a ledger_entry selector accepted on the wire as
// either a ledger entry index string or a structured selector object.
type EntrySelector[T any] struct {
	// Index is the ledger entry ID to encode as the selector's string form.
	Index string
	// Object is the structured selector to encode as the selector's object form.
	Object *T
}

// IsZero reports whether neither selector representation is set. It also lets
// encoding/json omit unset EntrySelector fields tagged with omitzero.
func (s EntrySelector[T]) IsZero() bool {
	return s.Index == "" && s.Object == nil
}

func (s EntrySelector[T]) validate() error {
	if (s.Index == "") == (s.Object == nil) {
		return ErrInvalidEntrySelector
	}
	return nil
}

// MarshalJSON encodes the selected string or object wire representation.
func (s EntrySelector[T]) MarshalJSON() ([]byte, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	if s.Index != "" {
		return json.Marshal(s.Index)
	}
	return json.Marshal(s.Object)
}

// UnmarshalJSON decodes either the string or object wire representation.
func (s *EntrySelector[T]) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return fmt.Errorf("%w: empty JSON value", ErrInvalidEntrySelector)
	}

	*s = EntrySelector[T]{}
	switch data[0] {
	case '"':
		if err := json.Unmarshal(data, &s.Index); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidEntrySelector, err)
		}
		if s.Index == "" {
			return fmt.Errorf("%w: index must not be empty", ErrInvalidEntrySelector)
		}
		return nil
	case '{':
		var object T
		if err := json.Unmarshal(data, &object); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidEntrySelector, err)
		}
		s.Object = &object
		return nil
	default:
		return fmt.Errorf("%w: expected string or object", ErrInvalidEntrySelector)
	}
}

// AMMSelectorFields identifies an AMM by its two pooled assets.
type AMMSelectorFields struct {
	Asset  ledgerentry.Asset `json:"asset"`
	Asset2 ledgerentry.Asset `json:"asset2"`
}

// AMMSelector is the string-or-object selector accepted by the amm field.
type AMMSelector = EntrySelector[AMMSelectorFields]

// BridgeSelector identifies a cross-chain bridge by its door accounts and assets.
type BridgeSelector struct {
	IssuingChainDoor  types.Address     `json:"IssuingChainDoor"`
	IssuingChainIssue ledgerentry.Asset `json:"IssuingChainIssue"`
	LockingChainDoor  types.Address     `json:"LockingChainDoor"`
	LockingChainIssue ledgerentry.Asset `json:"LockingChainIssue"`
}

// CredentialSelectorFields identifies a Credential ledger entry.
type CredentialSelectorFields struct {
	Subject        types.Address        `json:"subject"`
	Issuer         types.Address        `json:"issuer"`
	CredentialType types.CredentialType `json:"credential_type"`
}

// CredentialSelector is the string-or-object selector accepted by the credential field.
type CredentialSelector = EntrySelector[CredentialSelectorFields]

// DelegateSelectorFields identifies a Delegate ledger entry.
type DelegateSelectorFields struct {
	Account   types.Address `json:"account"`
	Authorize types.Address `json:"authorize"`
}

// DelegateSelector is the string-or-object selector accepted by the delegate field.
type DelegateSelector = EntrySelector[DelegateSelectorFields]

// DepositPreauthCredential identifies one credential in an authorized_credentials selector.
type DepositPreauthCredential struct {
	Issuer         types.Address        `json:"issuer"`
	CredentialType types.CredentialType `json:"credential_type"`
}

// DepositPreauthSelectorFields identifies a DepositPreauth ledger entry. Exactly
// one of Authorized and AuthorizedCredentials is expected by the server.
type DepositPreauthSelectorFields struct {
	Owner                 types.Address              `json:"owner"`
	Authorized            types.Address              `json:"authorized,omitempty"`
	AuthorizedCredentials []DepositPreauthCredential `json:"authorized_credentials,omitempty"`
}

// DepositPreauthSelector is the string-or-object selector accepted by the deposit_preauth field.
type DepositPreauthSelector = EntrySelector[DepositPreauthSelectorFields]

// DirectorySelectorFields identifies a DirectoryNode by its root or owner.
type DirectorySelectorFields struct {
	DirRoot  string        `json:"dir_root,omitempty"`
	Owner    types.Address `json:"owner,omitempty"`
	SubIndex *uint64       `json:"sub_index,omitempty"`
}

// DirectorySelector is the string-or-object selector accepted by the directory field.
type DirectorySelector = EntrySelector[DirectorySelectorFields]

// EscrowSelectorFields identifies an Escrow by owner and creation sequence.
type EscrowSelectorFields struct {
	Owner types.Address `json:"owner"`
	Seq   uint32        `json:"seq"`
}

// EscrowSelector is the string-or-object selector accepted by the escrow field.
type EscrowSelector = EntrySelector[EscrowSelectorFields]

// MPTokenSelectorFields identifies an MPToken holding by issuance and holder account.
type MPTokenSelectorFields struct {
	MPTIssuanceID types.MPTIssuanceID `json:"mpt_issuance_id"`
	Account       types.Address       `json:"account"`
}

// MPTokenSelector is the string-or-object selector accepted by the mptoken field.
type MPTokenSelector = EntrySelector[MPTokenSelectorFields]

// OfferSelectorFields identifies an Offer by account and creation sequence.
type OfferSelectorFields struct {
	Account types.Address `json:"account"`
	Seq     uint32        `json:"seq"`
}

// OfferSelector is the string-or-object selector accepted by the offer field.
type OfferSelector = EntrySelector[OfferSelectorFields]

// RippleStateSelector identifies a trust line by its two accounts and currency.
type RippleStateSelector struct {
	Accounts [2]types.Address `json:"accounts"`
	Currency string           `json:"currency"`
}

// TicketSelectorFields identifies a Ticket by account and ticket sequence.
type TicketSelectorFields struct {
	Account   types.Address `json:"account"`
	TicketSeq uint32        `json:"ticket_seq"`
}

// TicketSelector is the string-or-object selector accepted by the ticket field.
type TicketSelector = EntrySelector[TicketSelectorFields]

// XChainOwnedClaimIDSelectorFields identifies an XChainOwnedClaimID entry.
type XChainOwnedClaimIDSelectorFields struct {
	BridgeSelector
	XChainOwnedClaimID uint32 `json:"xchain_owned_claim_id"`
}

// XChainOwnedClaimIDSelector is the string-or-object selector accepted by the
// xchain_owned_claim_id field.
type XChainOwnedClaimIDSelector = EntrySelector[XChainOwnedClaimIDSelectorFields]

// XChainOwnedCreateAccountClaimIDSelectorFields identifies an
// XChainOwnedCreateAccountClaimID entry.
type XChainOwnedCreateAccountClaimIDSelectorFields struct {
	BridgeSelector
	XChainOwnedCreateAccountClaimID uint32 `json:"xchain_owned_create_account_claim_id"`
}

// XChainOwnedCreateAccountClaimIDSelector is the string-or-object selector
// accepted by the xchain_owned_create_account_claim_id field.
type XChainOwnedCreateAccountClaimIDSelector = EntrySelector[XChainOwnedCreateAccountClaimIDSelectorFields]
