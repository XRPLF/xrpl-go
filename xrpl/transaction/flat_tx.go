package transaction

import (
	"encoding/json"
)

var _ Tx = (*FlatTransaction)(nil)

// FlatTransaction is a flattened transaction represented as a map from field names to interface{} values.
// It satisfies the Tx interface for generic transaction handling.
type FlatTransaction map[string]any

// TxType returns the transaction type of the flattened transaction.
func (f FlatTransaction) TxType() TxType {
	txType, ok := f["TransactionType"].(string)
	if !ok {
		return TxType("")
	}
	return TxType(txType)
}

// Sequence returns the sequence number of the flattened transaction.
func (f FlatTransaction) Sequence() uint32 {
	sequence, ok := f["Sequence"].(json.Number)
	if ok {
		sequenceInt, err := sequence.Float64()
		if err != nil {
			return 0
		}
		return uint32(sequenceInt)
	}

	// Handle float64 case (when JSON is parsed as float64 instead of json.Number)
	if sequenceFloat, ok := f["Sequence"].(float64); ok {
		return uint32(sequenceFloat)
	}

	// Handle uint32 case (direct integer)
	if sequenceInt, ok := f["Sequence"].(uint32); ok {
		return sequenceInt
	}

	// Handle int case
	if sequenceInt, ok := f["Sequence"].(int); ok {
		if sequenceInt >= 0 && sequenceInt <= int(^uint32(0)) {
			return uint32(sequenceInt)
		}
	}

	return 0
}
