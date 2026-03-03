package vault

import "errors"

var (
	// ErrMissingLookupParam is returned when neither vault_id nor owner+seq is provided.
	ErrMissingLookupParam = errors.New("vault_info: must provide either vault_id or both owner and seq")
	// ErrConflictingLookupParams is returned when vault_id is provided together with owner or seq.
	ErrConflictingLookupParams = errors.New("vault_info: cannot use vault_id together with owner or seq")
	// ErrInvalidVaultID is returned when vault_id is not a valid 64-character hexadecimal string.
	ErrInvalidVaultID = errors.New("vault_info: vault_id must be a valid 64-character hexadecimal string")
	// ErrInvalidOwner is returned when owner is not a valid XRPL address.
	ErrInvalidOwner = errors.New("vault_info: owner must be a valid XRPL address")
	// ErrOwnerRequiresSeq is returned when owner is provided without seq.
	ErrOwnerRequiresSeq = errors.New("vault_info: owner requires seq to be specified")
	// ErrSeqRequiresOwner is returned when seq is provided without owner.
	ErrSeqRequiresOwner = errors.New("vault_info: seq requires owner to be specified")
)
