package account

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// GatewayBalancesRequest calculates total balances issued by an account, optionally excluding specified hot wallets.
type GatewayBalancesRequest struct {
	common.BaseRequest

	// The Address to check. This should be the issuing address.
	Account types.Address `json:"account"`
	// If true, only accept an address or public key for the account parameter.
	// Defaults to false.
	Strict bool `json:"strict,omitempty"`
	// An operational address to exclude from the balances issued, or an array of
	// Such addresses.
	HotWallet any `json:"hotwallet,omitempty"`
	// A 20-byte hex string for the ledger version to use.
	LedgerHash common.LedgerHash `json:"ledger_hash,omitempty"`
	// The ledger index of the ledger to use, or a shortcut string.
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
}

// Method returns the JSON-RPC method name for GatewayBalancesRequest.
func (r *GatewayBalancesRequest) Method() string {
	return "gateway_balances"
}

// APIVersion returns the API version supported by GatewayBalancesRequest.
func (r *GatewayBalancesRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate performs validation on GatewayBalancesRequest.
func (r *GatewayBalancesRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// GatewayBalance represents a currency balance entry in a gateway_balances response.
type GatewayBalance struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

// GatewayBalancesResponse represents the response from the gateway_balances method, including obligations, balances, and assets.
type GatewayBalancesResponse struct {
	// The address of the account that issued the balances.
	Account types.Address `json:"account"`

	// Total amounts issued to addresses not excluded, as a map of currencies
	// to the total value issued.
	Obligations map[string]string `json:"obligations,omitempty"`

	// Amounts issued to the hotwallet addresses from the request. The keys are
	// addresses and the values are arrays of currency amounts they hold.
	Balances map[string][]GatewayBalance `json:"balances,omitempty"`

	// Total amounts held that are issued by others. In the recommended
	// configuration, the issuing address should have none.
	Assets map[string][]GatewayBalance `json:"assets,omitempty"`

	// The identifying hash of the ledger version that was used to generate
	// this response.
	LedgerHash common.LedgerHash `json:"ledger_hash,omitempty"`

	// The ledger index of the ledger version that was used to generate this
	// response.
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index,omitempty"`

	// The ledger index of the current in-progress ledger version, which was
	// used to retrieve this information.
	LedgerIndex common.LedgerIndex `json:"ledger_index,omitempty"`
}
