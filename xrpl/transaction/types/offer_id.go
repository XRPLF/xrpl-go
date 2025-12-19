package types

// OfferID represents an offer ID.
type OfferID string

func (o *OfferID) String() string {
	return string(*o)
}
