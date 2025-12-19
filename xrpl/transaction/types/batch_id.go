package types

// BatchID represents a batch ID.
type BatchID string

func (b *BatchID) String() string {
	return string(*b)
}
