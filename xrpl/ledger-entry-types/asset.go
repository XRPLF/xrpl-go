package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// AssetKind represents the kind of asset.
type AssetKind int

const (
	// AssetXRP represents a native XRP asset.
	AssetXRP AssetKind = iota
	// AssetIOU represents an issued currency asset.
	AssetIOU
	// AssetMPT represents a multi-purpose token asset.
	AssetMPT
)

// Asset defines an asset identifier (without an amount). It can represent XRP, an IOU, or an MPT.
type Asset struct {
	Currency      string        `json:"currency,omitempty"`
	Issuer        types.Address `json:"issuer,omitempty"`
	MPTIssuanceID string        `json:"mpt_issuance_id,omitempty"`
}

// Kind returns the kind of asset: AssetXRP, AssetIOU, or AssetMPT.
func (a Asset) Kind() AssetKind {
	if a.MPTIssuanceID != "" {
		return AssetMPT
	}
	if a.Issuer != "" {
		return AssetIOU
	}
	return AssetXRP
}

// Flatten returns the flattened representation of the Asset.
func (a *Asset) Flatten() map[string]any {
	flattened := make(map[string]any)

	if a.MPTIssuanceID != "" {
		flattened["mpt_issuance_id"] = a.MPTIssuanceID
		return flattened
	}

	if a.Issuer.String() != "" {
		flattened["issuer"] = a.Issuer
	}

	if a.Currency != "" {
		flattened["currency"] = a.Currency
	}

	return flattened
}
