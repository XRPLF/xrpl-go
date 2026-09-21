//revive:disable:var-naming
package types

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
)

// XChainBridge represents the bridge configuration for cross-chain transfers, including door accounts and assets.
type XChainBridge struct {
	// The door account on the issuing chain. For an XRP-XRP bridge, this must be the
	// genesis account (the account that is created when the network is first started, which contains all of the XRP).
	IssuingChainDoor Address
	// The asset that is minted and burned on the issuing chain. For an IOU-IOU bridge,
	// the issuer of the asset must be the door account on the issuing chain, to avoid supply issues.
	IssuingChainIssue Address
	// The door account on the locking chain.
	LockingChainDoor Address
	// The asset that is locked and unlocked on the locking chain.
	LockingChainIssue Address
}

// FlatXChainBridge is a flattened representation of XChainBridge for JSON serialization.
type FlatXChainBridge map[string]string

// Flatten returns a FlatXChainBridge mapping fields to their string values.
func (x *XChainBridge) Flatten() FlatXChainBridge {
	flat := make(FlatXChainBridge)

	flat["IssuingChainDoor"] = x.IssuingChainDoor.Flatten()
	flat["IssuingChainIssue"] = x.IssuingChainIssue.Flatten()
	flat["LockingChainDoor"] = x.LockingChainDoor.Flatten()
	flat["LockingChainIssue"] = x.LockingChainIssue.Flatten()

	return flat
}

// Validate checks each address in the XChainBridge and returns false with an error if any are invalid.
func (x *XChainBridge) Validate() (bool, error) {
	if !addresscodec.IsValidAddress(x.IssuingChainDoor.String()) {
		return false, ErrInvalidIssuingChainDoorAddress
	}
	if !addresscodec.IsValidAddress(x.IssuingChainIssue.String()) {
		return false, ErrInvalidIssuingChainIssueAddress
	}
	if !addresscodec.IsValidAddress(x.LockingChainDoor.String()) {
		return false, ErrInvalidLockingChainDoorAddress
	}
	if !addresscodec.IsValidAddress(x.LockingChainIssue.String()) {
		return false, ErrInvalidLockingChainIssueAddress
	}

	return true, nil
}
