// Package types provides data structures for server query responses.
//
//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// ClosedLedger contains metadata for a closed ledger, such as age, fees, hash, and sequence.
type ClosedLedger struct {
	Age            uint          `json:"age"`
	BaseFeeXRP     *float64      `json:"base_fee_xrp"`
	Hash           types.Hash256 `json:"hash"`
	ReserveBaseXRP float32       `json:"reserve_base_xrp"`
	ReserveIncXRP  float32       `json:"reserve_inc_xrp"`
	Seq            uint          `json:"seq"`
}
