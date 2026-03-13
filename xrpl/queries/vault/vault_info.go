// Package vault contains vault-related queries for XRPL.
package vault

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// InfoRequest retrieves information about a Vault instance.
type InfoRequest struct {
	common.BaseRequest
	// The object ID of the Vault to be returned.
	VaultID string `json:"vault_id,omitempty"`
	// ID of the Vault Owner account.
	Owner types.Address `json:"owner,omitempty"`
	// Sequence number of the vault entry.
	Seq *uint32 `json:"seq,omitempty"`
}

// Method returns the JSON-RPC method name for InfoRequest.
func (r *InfoRequest) Method() string {
	return "vault_info"
}

// APIVersion returns the Rippled JSON-RPC API version for InfoRequest.
func (r *InfoRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate ensures the InfoRequest contains valid fields.
// Either VaultID alone, or both Owner and Seq must be provided.
func (r *InfoRequest) Validate() error {
	hasVaultID := r.VaultID != ""
	hasOwner := r.Owner != ""
	hasSeq := r.Seq != nil

	// Seq requires Owner
	if hasSeq && !hasOwner {
		return ErrSeqRequiresOwner
	}

	// Owner requires Seq
	if hasOwner && !hasSeq {
		return ErrOwnerRequiresSeq
	}

	// Must provide at least one lookup method: vault_id or owner+seq
	if !hasVaultID && !hasOwner {
		return ErrMissingLookupParam
	}

	// VaultID and Owner/Seq are mutually exclusive
	if hasVaultID && hasOwner {
		return ErrConflictingLookupParams
	}

	// Validate VaultID format
	if hasVaultID && !transaction.IsLedgerEntryID(r.VaultID) {
		return ErrInvalidVaultID
	}

	// Validate Owner address
	if hasOwner && !addresscodec.IsValidAddress(r.Owner.String()) {
		return ErrInvalidOwner
	}

	return nil
}

// ############################################################################
// Response
// ############################################################################

// Shares contains details about the shares issued by a Vault.
type Shares struct {
	// The ID of the Issuer of the Share. It will always be the pseudo-account ID.
	Issuer types.Address `json:"Issuer"`
	// Ledger entry type, always "MPTokenIssuance".
	LedgerEntryType string `json:"LedgerEntryType"`
	// Total outstanding shares issued.
	OutstandingAmount string `json:"OutstandingAmount"`
	// Transaction ID of the last modification to the shares' issuance.
	PreviousTxnID types.Hash256 `json:"PreviousTxnID"`
	// Ledger sequence number of the last transaction modifying the shares' issuance.
	PreviousTxnLgrSeq uint32 `json:"PreviousTxnLgrSeq"`
	// Sequence number of the shares issuance entry.
	Sequence uint32 `json:"Sequence"`
	// Unique index of the shares ledger entry.
	Index string `json:"index"`
	// Identifier for the owner node of the shares.
	OwnerNode string `json:"OwnerNode,omitempty"`
	// The ID of the MPTokenIssuance object. It will always be equal to vault.ShareMPTID.
	MPTIssuanceID string `json:"mpt_issuance_id,omitempty"`
	// The PermissionedDomain object ID associated with the shares of this Vault.
	DomainID string `json:"DomainID,omitempty"`
	// Bit-field flags associated with the shares issuance.
	Flags *uint32 `json:"Flags,omitempty"`
}

// Vault contains the vault data returned by the vault_info method.
type Vault struct {
	// The pseudo-account ID of the vault.
	Account types.Address `json:"Account"`
	// Object representing the asset held in the vault.
	Asset ledger.Asset `json:"Asset"`
	// Amount of assets currently available for withdrawal.
	AssetsAvailable string `json:"AssetsAvailable,omitempty"`
	// Total amount of assets in the vault.
	AssetsTotal string `json:"AssetsTotal,omitempty"`
	// Ledger entry type, always "Vault".
	LedgerEntryType string `json:"LedgerEntryType"`
	// ID of the Vault Owner account.
	Owner types.Address `json:"Owner"`
	// Transaction ID of the last modification to this vault.
	PreviousTxnID types.Hash256 `json:"PreviousTxnID"`
	// Ledger sequence number of the last transaction modifying this vault.
	PreviousTxnLgrSeq uint32 `json:"PreviousTxnLgrSeq"`
	// Sequence number of the vault entry.
	Sequence uint32 `json:"Sequence"`
	// Unique index of the vault ledger entry.
	Index string `json:"index"`
	// Object containing details about issued shares.
	Shares Shares `json:"shares"`
	// Unrealized loss associated with the vault.
	LossUnrealized string `json:"LossUnrealized,omitempty"`
	// Identifier for the owner node in the ledger tree.
	OwnerNode string `json:"OwnerNode,omitempty"`
	// Multi-Purpose token ID associated with this vault.
	ShareMPTID string `json:"ShareMPTID,omitempty"`
	// Policy defining withdrawal conditions.
	WithdrawalPolicy *types.VaultWithdrawalPolicy `json:"WithdrawalPolicy,omitempty"`
	// The maximum asset amount that can be held in the vault. Zero value indicates there is no cap.
	AssetsMaximum string `json:"AssetsMaximum,omitempty"`
	// Arbitrary metadata about the Vault. Limited to 256 bytes.
	Data string `json:"Data,omitempty"`
	// The scaling factor for vault shares. Only applicable for IOU assets.
	// Valid values are between 0 and 18 inclusive. For XRP and MPT, this is always 0.
	Scale *uint8 `json:"Scale,omitempty"`
	// Flags.
	Flags *uint32 `json:"Flags,omitempty"`
}

// Response is the response from the vault_info method.
type Response struct {
	// The vault data.
	Vault Vault `json:"vault"`
	// The identifying hash of the ledger that was used to generate this response.
	LedgerHash string `json:"ledger_hash,omitempty"`
	// The ledger index of the ledger version that was used to generate this response.
	LedgerIndex *uint32 `json:"ledger_index,omitempty"`
	// If included and set to true, the information in this response comes from
	// a validated ledger version. Otherwise, the information is subject to change.
	Validated bool `json:"validated,omitempty"`
}
