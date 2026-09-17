package clio

import (
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// AccountSponsoringRequest requests ledger objects sponsored by an account.
// Requires a server that supports account_sponsoring.
type AccountSponsoringRequest struct {
	common.BaseRequest
	Account              types.Address          `json:"account"`
	Type                 account.ObjectType     `json:"type,omitempty"`
	DeletionBlockersOnly bool                   `json:"deletion_blockers_only,omitempty"`
	LedgerHash           common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex          common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Limit                int                    `json:"limit,omitempty"`
	Marker               any                    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name.
func (*AccountSponsoringRequest) Method() string {
	return "account_sponsoring"
}

// APIVersion returns the supported API version.
func (*AccountSponsoringRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks that the required account is present.
func (r *AccountSponsoringRequest) Validate() error {
	if r.Account == "" {
		return account.ErrNoAccountID
	}
	return nil
}

// AccountSponsoringResponse contains raw sponsored ledger entries and pagination metadata.
type AccountSponsoringResponse struct {
	Account            types.Address             `json:"account"`
	SponsoredObjects   []ledger.FlatLedgerObject `json:"sponsored_objects"`
	LedgerHash         common.LedgerHash         `json:"ledger_hash,omitempty"`
	LedgerIndex        common.LedgerIndex        `json:"ledger_index,omitempty"`
	LedgerCurrentIndex common.LedgerIndex        `json:"ledger_current_index,omitempty"`
	Limit              int                       `json:"limit,omitempty"`
	Marker             any                       `json:"marker,omitempty"`
	Validated          bool                      `json:"validated,omitempty"`
}
