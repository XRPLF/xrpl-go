package ledger

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

var (
	// ErrInvalidEntryRequest is returned unless exactly one ledger entry selector is set.
	ErrInvalidEntryRequest = errors.New("ledger_entry: exactly one selector must be specified")
	// ErrInvalidBridgeSelector is returned unless bridge and bridge_account are set together.
	ErrInvalidBridgeSelector = errors.New("ledger_entry: bridge and bridge_account must be specified together")
)

// EntryRequest retrieves one ledger entry using exactly one selector.
type EntryRequest struct {
	common.BaseRequest
	Index                           string                                  `json:"index,omitempty"`
	AccountRoot                     types.Address                           `json:"account_root,omitempty"`
	AMM                             AMMSelector                             `json:"amm,omitzero"`
	BridgeAccount                   types.Address                           `json:"bridge_account,omitempty"`
	Bridge                          BridgeSelector                          `json:"bridge,omitzero"`
	Check                           string                                  `json:"check,omitempty"`
	Credential                      CredentialSelector                      `json:"credential,omitzero"`
	Delegate                        DelegateSelector                        `json:"delegate,omitzero"`
	DepositPreauth                  DepositPreauthSelector                  `json:"deposit_preauth,omitzero"`
	DID                             types.Address                           `json:"did,omitempty"`
	Directory                       DirectorySelector                       `json:"directory,omitzero"`
	Escrow                          EscrowSelector                          `json:"escrow,omitzero"`
	MPTIssuance                     types.MPTIssuanceID                     `json:"mpt_issuance,omitempty"`
	MPToken                         MPTokenSelector                         `json:"mptoken,omitzero"`
	NFTPage                         string                                  `json:"nft_page,omitempty"`
	Offer                           OfferSelector                           `json:"offer,omitzero"`
	PaymentChannel                  string                                  `json:"payment_channel,omitempty"`
	RippleState                     RippleStateSelector                     `json:"ripple_state,omitzero"`
	Ticket                          TicketSelector                          `json:"ticket,omitzero"`
	XChainOwnedClaimID              XChainOwnedClaimIDSelector              `json:"xchain_owned_claim_id,omitzero"`
	XChainOwnedCreateAccountClaimID XChainOwnedCreateAccountClaimIDSelector `json:"xchain_owned_create_account_claim_id,omitzero"`
	LedgerIndex                     common.LedgerSpecifier                  `json:"ledger_index,omitempty"`
	LedgerHash                      common.LedgerHash                       `json:"ledger_hash,omitempty"`
	Binary                          bool                                    `json:"binary,omitempty"`
	IncludeDeleted                  bool                                    `json:"include_deleted,omitempty"`
}

// Method returns the JSON-RPC method name for EntryRequest.
func (*EntryRequest) Method() string {
	return "ledger_entry"
}

// APIVersion returns the Rippled API version for EntryRequest.
func (*EntryRequest) APIVersion() int {
	return version.RippledAPIV2
}

// MarshalJSON uses the standard library encoder so omitzero selector fields are
// handled consistently by both RPC and WebSocket transports.
func (r EntryRequest) MarshalJSON() ([]byte, error) {
	type entryRequestAlias EntryRequest
	return json.Marshal(entryRequestAlias(r))
}

// Validate requires exactly one ledger entry selector and validates selectors
// that have multiple wire representations.
func (r *EntryRequest) Validate() error {
	hasBridge := r.Bridge != (BridgeSelector{})

	// One row per selector, in request field order. err is only consulted for
	// the selected selector, so eager validate() results on unset selectors are
	// harmless.
	selectors := []struct {
		name     string
		selected bool
		err      error
	}{
		{name: "index", selected: r.Index != ""},
		{name: "account_root", selected: r.AccountRoot != ""},
		{name: "amm", selected: !r.AMM.IsZero(), err: r.AMM.validate()},
		{name: "bridge", selected: hasBridge},
		{name: "check", selected: r.Check != ""},
		{name: "credential", selected: !r.Credential.IsZero(), err: r.Credential.validate()},
		{name: "delegate", selected: !r.Delegate.IsZero(), err: r.Delegate.validate()},
		{name: "deposit_preauth", selected: !r.DepositPreauth.IsZero(), err: r.DepositPreauth.validate()},
		{name: "did", selected: r.DID != ""},
		{name: "directory", selected: !r.Directory.IsZero(), err: r.Directory.validate()},
		{name: "escrow", selected: !r.Escrow.IsZero(), err: r.Escrow.validate()},
		{name: "mpt_issuance", selected: r.MPTIssuance != ""},
		{name: "mptoken", selected: !r.MPToken.IsZero(), err: r.MPToken.validate()},
		{name: "nft_page", selected: r.NFTPage != ""},
		{name: "offer", selected: !r.Offer.IsZero(), err: r.Offer.validate()},
		{name: "payment_channel", selected: r.PaymentChannel != ""},
		{name: "ripple_state", selected: r.RippleState != (RippleStateSelector{})},
		{name: "ticket", selected: !r.Ticket.IsZero(), err: r.Ticket.validate()},
		{name: "xchain_owned_claim_id", selected: !r.XChainOwnedClaimID.IsZero(), err: r.XChainOwnedClaimID.validate()},
		{name: "xchain_owned_create_account_claim_id", selected: !r.XChainOwnedCreateAccountClaimID.IsZero(), err: r.XChainOwnedCreateAccountClaimID.validate()},
	}

	selectorCount := 0
	for _, selector := range selectors {
		if selector.selected {
			selectorCount++
		}
	}
	if selectorCount != 1 {
		return ErrInvalidEntryRequest
	}

	if hasBridge != (r.BridgeAccount != "") {
		return ErrInvalidBridgeSelector
	}

	for _, selector := range selectors {
		if selector.selected && selector.err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidEntrySelector, selector.name)
		}
	}

	return nil
}

// ############################################################################
// Response
// ############################################################################

// ErrInvalidEntryResponse is returned when a ledger_entry response does not
// contain exactly one valid node or node_binary payload.
var ErrInvalidEntryResponse = errors.New("ledger_entry: response must contain exactly one of node or node_binary")

// EntryResponse is the response returned by ledger_entry. Node preserves the
// existing JSON response API, while NodeBinary contains binary response data.
// Successful responses must contain exactly one of these payload fields.
type EntryResponse struct {
	Index              string                  `json:"index"`
	LedgerHash         common.LedgerHash       `json:"ledger_hash,omitempty"`
	LedgerIndex        common.LedgerIndex      `json:"ledger_index,omitempty"`
	LedgerCurrentIndex common.LedgerIndex      `json:"ledger_current_index,omitempty"`
	Node               ledger.FlatLedgerObject `json:"node,omitempty"`
	NodeBinary         string                  `json:"node_binary,omitempty"`
	DeletedLedgerIndex common.LedgerIndex      `json:"deleted_ledger_index,omitempty"`
	Validated          bool                    `json:"validated"`
}

// UnmarshalJSON decodes a response containing either a JSON or binary node.
func (r *EntryResponse) UnmarshalJSON(data []byte) error {
	type entryResponse EntryResponse

	var decoded entryResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	hasNode := len(decoded.Node) > 0
	hasNodeBinary := decoded.NodeBinary != ""
	if hasNode == hasNodeBinary {
		return ErrInvalidEntryResponse
	}

	*r = EntryResponse(decoded)
	return nil
}
