package account

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ObjectType defines the type of ledger object for account_objects requests.
type ObjectType string

// ObjectType constants for account_objects requests.
const (
	CheckObject              ObjectType = "check"
	CredentialObject         ObjectType = "credential"
	DepositPreauthObject     ObjectType = "deposit_preauth"
	DIDObject                ObjectType = "did"
	EscrowObject             ObjectType = "escrow"
	MPToken                  ObjectType = "mptoken"
	MPTIssuance              ObjectType = "mpt_issuance"
	NFTOfferObject           ObjectType = "nft_offer"
	OfferObject              ObjectType = "offer"
	PaymentChannelObject     ObjectType = "payment_channel"
	PermissionedDomainObject ObjectType = "permissioned_domain"
	SignerListObject         ObjectType = "signer_list"
	StateObject              ObjectType = "state"
	TicketObject             ObjectType = "ticket"
)

// ############################################################################
// Request
// ############################################################################

// ObjectsRequest retrieves raw ledger objects owned by an account.
// For a higher-level view of trust lines and balances, see account_lines.
type ObjectsRequest struct {
	common.BaseRequest
	Account              types.Address          `json:"account"`
	Type                 ObjectType             `json:"type,omitempty"`
	DeletionBlockersOnly bool                   `json:"deletion_blockers_only,omitempty"`
	LedgerHash           common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex          common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Limit                int                    `json:"limit,omitempty"`
	Marker               any                    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name for ObjectsRequest.
func (*ObjectsRequest) Method() string {
	return "account_objects"
}

// APIVersion returns the API version supported by ObjectsRequest.
func (*ObjectsRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate performs validation of the ObjectsRequest.
// TODO: implement V2.
func (*ObjectsRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// ObjectsResponse represents the expected response from the account_objects method.
type ObjectsResponse struct {
	Account            types.Address             `json:"account"`
	AccountObjects     []ledger.FlatLedgerObject `json:"account_objects"`
	LedgerHash         common.LedgerHash         `json:"ledger_hash,omitempty"`
	LedgerIndex        common.LedgerIndex        `json:"ledger_index,omitempty"`
	LedgerCurrentIndex common.LedgerIndex        `json:"ledger_current_index,omitempty"`
	Limit              int                       `json:"limit,omitempty"`
	Marker             any                       `json:"marker,omitempty"`
	Validated          bool                      `json:"validated,omitempty"`
}
