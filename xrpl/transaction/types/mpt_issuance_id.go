package types

// MPTIssuanceID represents an MPT issuance ID.
type MPTIssuanceID string

func (mpt *MPTIssuanceID) String() string {
	return string(*mpt)
}
