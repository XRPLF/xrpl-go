package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// DefinitionsRequest is the request type for the server_definitions command.
// Hash is optional. When it matches the server's definitions hash, the server
// returns a hash-only response.
type DefinitionsRequest struct {
	common.BaseRequest
	Hash string `json:"hash,omitempty"`
}

// Method returns the JSON-RPC method name for DefinitionsRequest.
func (*DefinitionsRequest) Method() string {
	return "server_definitions"
}

// APIVersion returns the rippled API version for DefinitionsRequest.
func (*DefinitionsRequest) APIVersion() int {
	return version.RippledAPIV2
}

// MarshalJSON routes the WebSocket transport through the standard library
// encoder: without a json.Marshaler implementation, formatRequest falls back to
// mapstructure, which nests the embedded BaseRequest under its own key instead
// of flattening it.
func (r DefinitionsRequest) MarshalJSON() ([]byte, error) {
	type definitionsRequestAlias DefinitionsRequest
	return json.Marshal(definitionsRequestAlias(r))
}

// Validate verifies that an optional definitions hash is a 256-bit hexadecimal string.
func (r *DefinitionsRequest) Validate() error {
	if r.Hash == "" {
		return nil
	}
	return validateDefinitionsHash(r.Hash)
}

// DefinitionFieldInfo describes one serialized field in a server definitions payload.
type DefinitionFieldInfo struct {
	Nth            int    `json:"nth"`
	IsVLEncoded    bool   `json:"isVLEncoded"`
	IsSerialized   bool   `json:"isSerialized"`
	IsSigningField bool   `json:"isSigningField"`
	Type           string `json:"type"`
}

// UnmarshalJSON requires every FIELDS properties entry to include all wire fields.
func (i *DefinitionFieldInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		Nth            *int    `json:"nth"`
		IsVLEncoded    *bool   `json:"isVLEncoded"`
		IsSerialized   *bool   `json:"isSerialized"`
		IsSigningField *bool   `json:"isSigningField"`
		Type           *string `json:"type"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: properties: %w", ErrInvalidDefinitionField, err)
	}
	if raw.Nth == nil || raw.IsVLEncoded == nil || raw.IsSerialized == nil ||
		raw.IsSigningField == nil || raw.Type == nil {
		return fmt.Errorf("%w: properties must contain nth, isVLEncoded, isSerialized, isSigningField, and type", ErrInvalidDefinitionField)
	}

	*i = DefinitionFieldInfo{
		Nth:            *raw.Nth,
		IsVLEncoded:    *raw.IsVLEncoded,
		IsSerialized:   *raw.IsSerialized,
		IsSigningField: *raw.IsSigningField,
		Type:           *raw.Type,
	}
	return nil
}

// DefinitionField represents one FIELDS tuple. On the wire it is encoded as
// [field name, field properties], rather than as a JSON object.
type DefinitionField struct {
	Name string
	Info DefinitionFieldInfo
}

// MarshalJSON encodes a DefinitionField as its two-element wire tuple.
func (f DefinitionField) MarshalJSON() ([]byte, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal([2]any{f.Name, f.Info})
}

// UnmarshalJSON decodes a two-element FIELDS wire tuple.
func (f *DefinitionField) UnmarshalJSON(data []byte) error {
	var tuple []json.RawMessage
	if err := json.Unmarshal(data, &tuple); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDefinitionField, err)
	}
	if len(tuple) != 2 {
		return fmt.Errorf("%w: expected a two-element tuple", ErrInvalidDefinitionField)
	}

	var decoded DefinitionField
	if err := json.Unmarshal(tuple[0], &decoded.Name); err != nil {
		return fmt.Errorf("%w: name: %w", ErrInvalidDefinitionField, err)
	}
	if err := json.Unmarshal(tuple[1], &decoded.Info); err != nil {
		return fmt.Errorf("%w: properties: %w", ErrInvalidDefinitionField, err)
	}
	if err := decoded.Validate(); err != nil {
		return err
	}

	*f = decoded
	return nil
}

// Validate verifies that a definition field has a name and serialized type.
func (f DefinitionField) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("%w: name must not be empty", ErrInvalidDefinitionField)
	}
	if f.Info.Type == "" {
		return fmt.Errorf("%w: type must not be empty", ErrInvalidDefinitionField)
	}
	return nil
}

// DefinitionFormatField describes a field in a transaction or ledger-entry format.
// Optionality uses rippled's SOEStyle values: -1 invalid, 0 required, 1 optional,
// and 2 default.
type DefinitionFormatField struct {
	Name        string `json:"name"`
	Optionality int    `json:"optionality"`
}

// UnmarshalJSON requires both wire fields, including an explicit zero optionality.
func (f *DefinitionFormatField) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name        *string `json:"name"`
		Optionality *int    `json:"optionality"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: format field: %w", ErrInvalidDefinitionsResponse, err)
	}
	if raw.Name == nil || raw.Optionality == nil {
		return fmt.Errorf("%w: format field must contain name and optionality", ErrInvalidDefinitionsResponse)
	}

	*f = DefinitionFormatField{Name: *raw.Name, Optionality: *raw.Optionality}
	return f.Validate()
}

// Validate verifies a format field name and SOEStyle optionality value.
func (f DefinitionFormatField) Validate() error {
	if f.Name == "" {
		return errors.New("name must not be empty")
	}
	if f.Optionality < -1 || f.Optionality > 2 {
		return errors.New("optionality must be between -1 and 2")
	}
	return nil
}

// DefinitionsResponse is the response returned by server_definitions.
// Hash is always present. The five core definition sections are present together
// in a full response and omitted together in a hash-only unchanged response.
// Format and flag maps were added later by XLS-97 and remain optional so responses
// from servers predating those additions can still be decoded.
//
// These fields are wire DTOs and are independent from binary-codec's embedded
// definitions and singleton state.
type DefinitionsResponse struct {
	Fields             []DefinitionField                  `json:"FIELDS,omitempty"`
	Types              map[string]int                     `json:"TYPES,omitempty"`
	LedgerEntryTypes   map[string]int                     `json:"LEDGER_ENTRY_TYPES,omitempty"`
	TransactionTypes   map[string]int                     `json:"TRANSACTION_TYPES,omitempty"`
	TransactionResults map[string]int                     `json:"TRANSACTION_RESULTS,omitempty"`
	LedgerEntryFormats map[string][]DefinitionFormatField `json:"LEDGER_ENTRY_FORMATS,omitempty"`
	TransactionFormats map[string][]DefinitionFormatField `json:"TRANSACTION_FORMATS,omitempty"`
	LedgerEntryFlags   map[string]map[string]uint32       `json:"LEDGER_ENTRY_FLAGS,omitempty"`
	TransactionFlags   map[string]map[string]uint32       `json:"TRANSACTION_FLAGS,omitempty"`
	AccountSetFlags    map[string]uint32                  `json:"ACCOUNT_SET_FLAGS,omitempty"`
	Hash               string                             `json:"hash"`
}

func (r DefinitionsResponse) isHashOnly() bool {
	return r.Fields == nil && r.Types == nil && r.LedgerEntryTypes == nil &&
		r.TransactionTypes == nil && r.TransactionResults == nil &&
		r.LedgerEntryFormats == nil && r.TransactionFormats == nil &&
		r.LedgerEntryFlags == nil && r.TransactionFlags == nil && r.AccountSetFlags == nil
}

// ValidateForRequest verifies that a hash-only response matches the request hash.
func (r DefinitionsResponse) ValidateForRequest(req *DefinitionsRequest) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if !r.isHashOnly() {
		return nil
	}
	if req == nil || req.Hash == "" || !strings.EqualFold(req.Hash, r.Hash) {
		return fmt.Errorf("%w: hash-only response does not match request hash", ErrInvalidDefinitionsResponse)
	}
	return nil
}

// Validate verifies the full and hash-only server_definitions response variants.
func (r DefinitionsResponse) Validate() error {
	if err := validateDefinitionsHash(r.Hash); err != nil {
		return err
	}

	hasPayload := r.Fields != nil || r.Types != nil || r.LedgerEntryTypes != nil ||
		r.TransactionTypes != nil || r.TransactionResults != nil ||
		r.LedgerEntryFormats != nil || r.TransactionFormats != nil ||
		r.LedgerEntryFlags != nil || r.TransactionFlags != nil || r.AccountSetFlags != nil
	if !hasPayload {
		return nil
	}

	if len(r.Fields) == 0 || len(r.Types) == 0 || len(r.LedgerEntryTypes) == 0 ||
		len(r.TransactionTypes) == 0 || len(r.TransactionResults) == 0 {
		return ErrInvalidDefinitionsResponse
	}

	for _, field := range r.Fields {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidDefinitionsResponse, err)
		}
	}
	if err := validateFormatSection("LEDGER_ENTRY_FORMATS", r.LedgerEntryFormats); err != nil {
		return err
	}
	return validateFormatSection("TRANSACTION_FORMATS", r.TransactionFormats)
}

func validateFormatSection(section string, formats map[string][]DefinitionFormatField) error {
	for formatName, fields := range formats {
		if formatName == "" {
			return fmt.Errorf("%w: %s contains an empty format name", ErrInvalidDefinitionsResponse, section)
		}
		for _, field := range fields {
			if err := field.Validate(); err != nil {
				return fmt.Errorf("%w: %s.%s: %w", ErrInvalidDefinitionsResponse, section, formatName, err)
			}
		}
	}
	return nil
}

// MarshalJSON validates and encodes a full or hash-only definitions response.
func (r DefinitionsResponse) MarshalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	type definitionsResponseAlias DefinitionsResponse
	return json.Marshal(definitionsResponseAlias(r))
}

// UnmarshalJSON decodes and validates a full or hash-only definitions response.
func (r *DefinitionsResponse) UnmarshalJSON(data []byte) error {
	fields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDefinitionsResponse, err)
	}

	sections := []string{
		"FIELDS",
		"TYPES",
		"LEDGER_ENTRY_TYPES",
		"TRANSACTION_TYPES",
		"TRANSACTION_RESULTS",
		"LEDGER_ENTRY_FORMATS",
		"TRANSACTION_FORMATS",
		"LEDGER_ENTRY_FLAGS",
		"TRANSACTION_FLAGS",
		"ACCOUNT_SET_FLAGS",
	}
	for _, section := range sections {
		value, present := fields[section]
		if present && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("%w: %s must not be null", ErrInvalidDefinitionsResponse, section)
		}
	}

	type definitionsResponseAlias DefinitionsResponse
	var decoded definitionsResponseAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDefinitionsResponse, err)
	}
	response := DefinitionsResponse(decoded)
	if err := response.Validate(); err != nil {
		return err
	}
	*r = response
	return nil
}

func validateDefinitionsHash(hash string) error {
	if len(hash) != 64 || !typecheck.IsHex(hash) {
		return ErrInvalidDefinitionsHash
	}
	return nil
}
