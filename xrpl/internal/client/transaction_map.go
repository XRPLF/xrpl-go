package client

// transactionMap is an alias, not a defined type, so FlatTransaction values
// and raw Batch maps remain assignable without conversions. Batch inner
// transactions use map[string]any in transaction/types, which cannot import its
// parent transaction package without an import cycle. The alias lets these
// helpers handle outer and inner transactions with one type.
type transactionMap = map[string]any
