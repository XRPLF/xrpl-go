package ledger

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// quotedUInt64 marshals a UInt64 MPT amount field as a quoted decimal string,
// the XRPL JSON wire format for these fields (other ledger UInt64 fields, such
// as OwnerNode, use zero-padded hex instead).
type quotedUInt64 uint64

func (u quotedUInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(u), 10))
}

func (u *quotedUInt64) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("UInt64 must be a quoted decimal string: %w", err)
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fmt.Errorf("UInt64 must be a non-empty string of decimal digits: %w", err)
	}
	*u = quotedUInt64(parsed)
	return nil
}

type mpTokenJSON struct {
	Index             types.Hash256 `json:"index,omitempty"`
	LedgerEntryType   EntryType
	Flags             uint32
	Account           types.Address
	MPTokenIssuanceID types.Hash192
	MPTAmount         quotedUInt64
	LockedAmount      quotedUInt64 `json:",omitempty"`
	PreviousTxnID     types.Hash256
	PreviousTxnLgrSeq uint32
	OwnerNode         uint64
}

func (m MPToken) wire() mpTokenJSON {
	return mpTokenJSON{
		Index:             m.Index,
		LedgerEntryType:   m.LedgerEntryType,
		Flags:             m.Flags,
		Account:           m.Account,
		MPTokenIssuanceID: m.MPTokenIssuanceID,
		MPTAmount:         quotedUInt64(m.MPTAmount),
		LockedAmount:      quotedUInt64(m.LockedAmount),
		PreviousTxnID:     m.PreviousTxnID,
		PreviousTxnLgrSeq: m.PreviousTxnLgrSeq,
		OwnerNode:         m.OwnerNode,
	}
}

func (w mpTokenJSON) entry() MPToken {
	return MPToken{
		Index:             w.Index,
		LedgerEntryType:   w.LedgerEntryType,
		Flags:             w.Flags,
		Account:           w.Account,
		MPTokenIssuanceID: w.MPTokenIssuanceID,
		MPTAmount:         uint64(w.MPTAmount),
		LockedAmount:      uint64(w.LockedAmount),
		PreviousTxnID:     w.PreviousTxnID,
		PreviousTxnLgrSeq: w.PreviousTxnLgrSeq,
		OwnerNode:         w.OwnerNode,
	}
}

// MarshalJSON serializes MPToken UInt64 amount fields as quoted decimal strings.
func (m MPToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.wire())
}

// UnmarshalJSON parses MPToken UInt64 amount fields from quoted decimal strings.
func (m *MPToken) UnmarshalJSON(data []byte) error {
	// Seed from the receiver so fields absent from the JSON keep their current
	// values, matching encoding/json merge semantics.
	decoded := m.wire()
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*m = decoded.entry()
	return nil
}

type mpTokenIssuanceJSON struct {
	Index             types.Hash256 `json:"index,omitempty"`
	LedgerEntryType   EntryType
	Flags             uint32
	Issuer            types.Address
	AssetScale        uint8
	MaximumAmount     quotedUInt64 `json:",omitempty"`
	OutstandingAmount quotedUInt64
	TransferFee       uint16
	MPTokenMetadata   string
	OwnerNode         uint64
	PreviousTxnID     types.Hash256
	PreviousTxnLgrSeq uint32
	Sequence          uint32
	LockedAmount      quotedUInt64 `json:",omitempty"`
	DomainID          string       `json:",omitempty"`
	MutableFlags      uint32       `json:",omitempty"`
}

func (m MPTokenIssuance) wire() mpTokenIssuanceJSON {
	return mpTokenIssuanceJSON{
		Index:             m.Index,
		LedgerEntryType:   m.LedgerEntryType,
		Flags:             m.Flags,
		Issuer:            m.Issuer,
		AssetScale:        m.AssetScale,
		MaximumAmount:     quotedUInt64(m.MaximumAmount),
		OutstandingAmount: quotedUInt64(m.OutstandingAmount),
		TransferFee:       m.TransferFee,
		MPTokenMetadata:   m.MPTokenMetadata,
		OwnerNode:         m.OwnerNode,
		PreviousTxnID:     m.PreviousTxnID,
		PreviousTxnLgrSeq: m.PreviousTxnLgrSeq,
		Sequence:          m.Sequence,
		LockedAmount:      quotedUInt64(m.LockedAmount),
		DomainID:          m.DomainID,
		MutableFlags:      m.MutableFlags,
	}
}

func (w mpTokenIssuanceJSON) entry() MPTokenIssuance {
	return MPTokenIssuance{
		Index:             w.Index,
		LedgerEntryType:   w.LedgerEntryType,
		Flags:             w.Flags,
		Issuer:            w.Issuer,
		AssetScale:        w.AssetScale,
		MaximumAmount:     uint64(w.MaximumAmount),
		OutstandingAmount: uint64(w.OutstandingAmount),
		TransferFee:       w.TransferFee,
		MPTokenMetadata:   w.MPTokenMetadata,
		OwnerNode:         w.OwnerNode,
		PreviousTxnID:     w.PreviousTxnID,
		PreviousTxnLgrSeq: w.PreviousTxnLgrSeq,
		Sequence:          w.Sequence,
		LockedAmount:      uint64(w.LockedAmount),
		DomainID:          w.DomainID,
		MutableFlags:      w.MutableFlags,
	}
}

// MarshalJSON serializes MPTokenIssuance UInt64 amount fields as quoted decimal strings.
func (m MPTokenIssuance) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.wire())
}

// UnmarshalJSON parses MPTokenIssuance UInt64 amount fields from quoted decimal strings.
func (m *MPTokenIssuance) UnmarshalJSON(data []byte) error {
	// Seed from the receiver so fields absent from the JSON keep their current
	// values, matching encoding/json merge semantics.
	decoded := m.wire()
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*m = decoded.entry()
	return nil
}
