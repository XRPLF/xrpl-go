package ledger

import "encoding/json"

// Embedding aliases keeps the custom JSON representation limited to UInt64
// overrides, so fields added to the public ledger models are not dropped. Any
// future UInt64 field must also be mirrored here with quotedUInt64 for base-ten
// amounts or hexUInt64 for other UInt64 fields.
type mpTokenAlias MPToken

type mpTokenJSON struct {
	mpTokenAlias
	MPTAmount    quotedUInt64 `json:",omitempty"`
	LockedAmount quotedUInt64 `json:",omitempty"`
	OwnerNode    hexUInt64
}

func (m MPToken) wire() mpTokenJSON {
	return mpTokenJSON{
		mpTokenAlias: mpTokenAlias(m),
		MPTAmount:    quotedUInt64(m.MPTAmount),
		LockedAmount: quotedUInt64(m.LockedAmount),
		OwnerNode:    hexUInt64(m.OwnerNode),
	}
}

func (w mpTokenJSON) entry() MPToken {
	entry := MPToken(w.mpTokenAlias)
	entry.MPTAmount = uint64(w.MPTAmount)
	entry.LockedAmount = uint64(w.LockedAmount)
	entry.OwnerNode = uint64(w.OwnerNode)
	return entry
}

// MarshalJSON serializes MPToken UInt64 fields in canonical XRPL JSON form.
func (m MPToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.wire())
}

// UnmarshalJSON parses canonical strings and unsigned JSON integers.
func (m *MPToken) UnmarshalJSON(data []byte) error {
	// Seed from the receiver so absent fields keep their current values, matching
	// encoding/json merge semantics.
	decoded := m.wire()
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*m = decoded.entry()
	return nil
}

type mpTokenIssuanceAlias MPTokenIssuance

type mpTokenIssuanceJSON struct {
	mpTokenIssuanceAlias
	MaximumAmount     quotedUInt64 `json:",omitempty"`
	OutstandingAmount quotedUInt64
	LockedAmount      quotedUInt64 `json:",omitempty"`
	OwnerNode         hexUInt64
}

func (m MPTokenIssuance) wire() mpTokenIssuanceJSON {
	return mpTokenIssuanceJSON{
		mpTokenIssuanceAlias: mpTokenIssuanceAlias(m),
		MaximumAmount:        quotedUInt64(m.MaximumAmount),
		OutstandingAmount:    quotedUInt64(m.OutstandingAmount),
		LockedAmount:         quotedUInt64(m.LockedAmount),
		OwnerNode:            hexUInt64(m.OwnerNode),
	}
}

func (w mpTokenIssuanceJSON) entry() MPTokenIssuance {
	entry := MPTokenIssuance(w.mpTokenIssuanceAlias)
	entry.MaximumAmount = uint64(w.MaximumAmount)
	entry.OutstandingAmount = uint64(w.OutstandingAmount)
	entry.LockedAmount = uint64(w.LockedAmount)
	entry.OwnerNode = uint64(w.OwnerNode)
	return entry
}

// MarshalJSON serializes MPTokenIssuance UInt64 fields in canonical XRPL JSON form.
func (m MPTokenIssuance) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.wire())
}

// UnmarshalJSON parses canonical strings and unsigned JSON integers.
func (m *MPTokenIssuance) UnmarshalJSON(data []byte) error {
	// Seed from the receiver so absent fields keep their current values, matching
	// encoding/json merge semantics.
	decoded := m.wire()
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*m = decoded.entry()
	return nil
}
