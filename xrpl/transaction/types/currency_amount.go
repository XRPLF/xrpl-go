//revive:disable:var-naming
package types

import (
	"encoding/json"
	"strconv"
)

// CurrencyKind indicates the type of a currency amount (XRP, ISSUED, MPT).
type CurrencyKind int

// CurrencyKind constants enumerate supported currency amount kinds.
const (
	// XRP is the native XRP amount type.
	XRP CurrencyKind = iota
	// ISSUED is a non-native currency amount issued by an account.
	ISSUED
	// MPT is a multi-party token currency amount.
	MPT
)

// CurrencyAmount defines methods for types representing XRP Ledger currency amounts.
type CurrencyAmount interface {
	Kind() CurrencyKind
	Flatten() any
}

// UnmarshalCurrencyAmount parses JSON data into the appropriate CurrencyAmount implementation.
func UnmarshalCurrencyAmount(data []byte) (CurrencyAmount, error) {
	if len(data) == 0 {
		return nil, nil
	}
	switch data[0] {
	case '{':
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}

		if _, hasMPTID := raw["mpt_issuance_id"]; hasMPTID {
			var m MPTCurrencyAmount
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, err
			}
			return m, nil
		}

		var i IssuedCurrencyAmount
		if err := json.Unmarshal(data, &i); err != nil {
			return nil, err
		}
		return i, nil
	default:
		var x XRPCurrencyAmount
		if err := json.Unmarshal(data, &x); err != nil {
			return nil, err
		}
		return x, nil
	}
}

// IssuedCurrencyAmount represents an amount of an issued (non-XRP) currency.
type IssuedCurrencyAmount struct {
	Issuer   Address `json:"issuer,omitempty"`
	Currency string  `json:"currency,omitempty"`
	Value    string  `json:"value,omitempty"`
}

// Kind returns the CurrencyKind for IssuedCurrencyAmount.
func (IssuedCurrencyAmount) Kind() CurrencyKind {
	return ISSUED
}

// Flatten returns a map[string]any representation of the issued currency amount.
func (i IssuedCurrencyAmount) Flatten() any {
	json := make(map[string]any)

	if i.Issuer != "" {
		json["issuer"] = i.Issuer.String()
	}

	if i.Currency != "" {
		json["currency"] = i.Currency
	}

	if i.Value != "" {
		json["value"] = i.Value
	}
	return json
}

// IsZero returns true if the IssuedCurrencyAmount is the zero value (empty object).
func (i IssuedCurrencyAmount) IsZero() bool {
	return i == IssuedCurrencyAmount{}
}

// XRPCurrencyAmount represents the native XRP amount in drops.
type XRPCurrencyAmount uint64

// Uint64 returns the XRP amount in drops as a uint64.
func (a XRPCurrencyAmount) Uint64() uint64 {
	return uint64(a)
}

func (a XRPCurrencyAmount) String() string {
	return strconv.FormatUint(uint64(a), 10)
}

// Kind returns the CurrencyKind for XRPCurrencyAmount.
func (XRPCurrencyAmount) Kind() CurrencyKind {
	return XRP
}

// Flatten returns the XRP amount as a decimal string.
func (a XRPCurrencyAmount) Flatten() any {
	return a.String()
}

// MarshalJSON serializes the XRP amount as a JSON string.
func (a XRPCurrencyAmount) MarshalJSON() ([]byte, error) {
	s := strconv.FormatUint(uint64(a), 10)
	return json.Marshal(s)
}

// UnmarshalJSON parses a JSON string into an XRPCurrencyAmount.
func (a *XRPCurrencyAmount) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*a = XRPCurrencyAmount(v)
	return nil
}

// UnmarshalText parses a text representation into an XRPCurrencyAmount.
func (a *XRPCurrencyAmount) UnmarshalText(data []byte) error {
	v, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return err
	}
	*a = XRPCurrencyAmount(v)
	return nil
}

// MPTCurrencyAmount represents a multi-party token currency amount with issuance ID and value.
type MPTCurrencyAmount struct {
	MPTIssuanceID string `json:"mpt_issuance_id"`
	Value         string `json:"value"`
}

// Kind returns the CurrencyKind for MPTCurrencyAmount.
func (MPTCurrencyAmount) Kind() CurrencyKind {
	return MPT
}

// Flatten returns a map[string]any representation of the MPT currency amount.
func (m MPTCurrencyAmount) Flatten() any {
	json := make(map[string]any)
	if m.MPTIssuanceID != "" {
		json["mpt_issuance_id"] = m.MPTIssuanceID
	}
	if m.Value != "" {
		json["value"] = m.Value
	}
	return json
}

// IsValid returns true if the MPTCurrencyAmount has both issuance ID and value.
func (m MPTCurrencyAmount) IsValid() bool {
	return m.MPTIssuanceID != "" && m.Value != ""
}
