// Package types provides data structures for server query responses.
//
//revive:disable:var-naming
package types

import (
	"encoding/json"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ClosedLedger contains metadata for a closed ledger, such as age, fees, hash, and sequence.
type ClosedLedger struct {
	Age                   uint          `json:"age"`
	BaseFeeXRP            float64       `json:"base_fee_xrp"`
	Hash                  types.Hash256 `json:"hash"`
	ReserveBaseXRP        float32       `json:"reserve_base_xrp"`
	ReserveIncXRP         float32       `json:"reserve_inc_xrp"`
	Seq                   uint          `json:"seq"`
	baseFeeXRPZeroPresent bool
}

// UnmarshalJSON records whether base_fee_xrp was present and non-null while
// preserving the public numeric field.
func (l *ClosedLedger) UnmarshalJSON(data []byte) error {
	type Alias ClosedLedger
	aux := struct {
		BaseFeeXRP *float64 `json:"base_fee_xrp"`
		Alias
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	*l = ClosedLedger(aux.Alias)
	if aux.BaseFeeXRP != nil {
		l.BaseFeeXRP = *aux.BaseFeeXRP
		l.baseFeeXRPZeroPresent = l.BaseFeeXRP == 0
	}
	return nil
}

// BaseFeeXRPValue returns the base fee and whether the JSON field was present
// and non-null.
func (l ClosedLedger) BaseFeeXRPValue() (float64, bool) {
	return l.BaseFeeXRP, l.BaseFeeXRP != 0 || l.baseFeeXRPZeroPresent
}
