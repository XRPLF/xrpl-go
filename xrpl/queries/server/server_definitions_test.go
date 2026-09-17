package server_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/stretchr/testify/require"
)

const definitionsHash = "C685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"

const fullDefinitionsFixture = `{
	"FIELDS": [
		["Invalid", {"isSerialized": false, "isSigningField": false, "isVLEncoded": false, "nth": -1, "type": "Unknown"}],
		["TransactionType", {"isSerialized": true, "isSigningField": true, "isVLEncoded": false, "nth": 2, "type": "UInt16"}]
	],
	"TYPES": {"Done": -1, "UInt16": 1},
	"LEDGER_ENTRY_TYPES": {"Invalid": -1, "AccountRoot": 97},
	"TRANSACTION_TYPES": {"Invalid": -1, "Payment": 0},
	"TRANSACTION_RESULTS": {"temREDUNDANT": -275, "tesSUCCESS": 0},
	"LEDGER_ENTRY_FORMATS": {
		"common": [{"name": "LedgerEntryType", "optionality": 0}],
		"AccountRoot": [{"name": "Account", "optionality": 0}]
	},
	"TRANSACTION_FORMATS": {
		"common": [{"name": "TransactionType", "optionality": 0}],
		"Payment": [{"name": "Destination", "optionality": 0}]
	},
	"LEDGER_ENTRY_FLAGS": {"AccountRoot": {"lsfAllowTrustLineClawback": 2147483648}},
	"TRANSACTION_FLAGS": {"Payment": {"tfPartialPayment": 131072}},
	"ACCOUNT_SET_FLAGS": {"asfDefaultRipple": 8},
	"hash": "C685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"
}`

func TestDefinitionsRequest(t *testing.T) {
	tests := []struct {
		name    string
		request server.DefinitionsRequest
		wantErr error
	}{
		{name: "without hash"},
		{name: "with matching-length hash", request: server.DefinitionsRequest{Hash: definitionsHash}},
		{name: "reject short hash", request: server.DefinitionsRequest{Hash: "ABCD"}, wantErr: server.ErrInvalidDefinitionsHash},
		{name: "reject non-hex hash", request: server.DefinitionsRequest{Hash: "Z685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"}, wantErr: server.ErrInvalidDefinitionsHash},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "server_definitions", tt.request.Method())
			require.Equal(t, version.RippledAPIV2, tt.request.APIVersion())
		})
	}
}

func TestDefinitionTypesMarshalIncompleteValues(t *testing.T) {
	t.Run("definition field", func(t *testing.T) {
		field := server.DefinitionField{}
		require.ErrorIs(t, field.Validate(), server.ErrInvalidDefinitionField)

		encoded, err := json.Marshal(field)
		require.NoError(t, err)
		require.JSONEq(t, `["", {
			"nth": 0,
			"isVLEncoded": false,
			"isSerialized": false,
			"isSigningField": false,
			"type": ""
		}]`, string(encoded))
	})

	t.Run("definitions response", func(t *testing.T) {
		response := server.DefinitionsResponse{}
		require.ErrorIs(t, response.Validate(), server.ErrInvalidDefinitionsHash)

		encoded, err := json.Marshal(response)
		require.NoError(t, err)
		require.JSONEq(t, `{"hash":""}`, string(encoded))
	})
}

func definitionsResponseFullFixture() (server.DefinitionsResponse, string) {
	expected := server.DefinitionsResponse{
		Fields: []server.DefinitionField{
			{Name: "Invalid", Info: server.DefinitionFieldInfo{Nth: -1, Type: "Unknown"}},
			{Name: "TransactionType", Info: server.DefinitionFieldInfo{Nth: 2, Type: "UInt16", IsSerialized: true, IsSigningField: true}},
		},
		Types:              map[string]int{"Done": -1, "UInt16": 1},
		LedgerEntryTypes:   map[string]int{"Invalid": -1, "AccountRoot": 97},
		TransactionTypes:   map[string]int{"Invalid": -1, "Payment": 0},
		TransactionResults: map[string]int{"temREDUNDANT": -275, "tesSUCCESS": 0},
		LedgerEntryFormats: map[string][]server.DefinitionFormatField{
			"common":      {{Name: "LedgerEntryType", Optionality: 0}},
			"AccountRoot": {{Name: "Account", Optionality: 0}},
		},
		TransactionFormats: map[string][]server.DefinitionFormatField{
			"common":  {{Name: "TransactionType", Optionality: 0}},
			"Payment": {{Name: "Destination", Optionality: 0}},
		},
		LedgerEntryFlags: map[string]map[string]uint32{"AccountRoot": {"lsfAllowTrustLineClawback": 2147483648}},
		TransactionFlags: map[string]map[string]uint32{"Payment": {"tfPartialPayment": 131072}},
		AccountSetFlags:  map[string]uint32{"asfDefaultRipple": 8},
		Hash:             definitionsHash,
	}
	return expected, fullDefinitionsFixture
}

func TestDefinitionsResponseFullSerialize(t *testing.T) {
	value, payload := definitionsResponseFullFixture()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, payload, string(encoded))
}

func TestDefinitionsResponseFullJSONDecode(t *testing.T) {
	want, payload := definitionsResponseFullFixture()
	var got server.DefinitionsResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestDefinitionsResponseFullClientDecode(t *testing.T) {
	want, payload := definitionsResponseFullFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got server.DefinitionsResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}

func definitionsResponseLegacyFixture() (server.DefinitionsResponse, string) {
	fixture := `{
		"FIELDS": [["TransactionType", {"isSerialized": true, "isSigningField": true, "isVLEncoded": false, "nth": 2, "type": "UInt16"}]],
		"TYPES": {"UInt16": 1},
		"LEDGER_ENTRY_TYPES": {"AccountRoot": 97},
		"TRANSACTION_TYPES": {"Payment": 0},
		"TRANSACTION_RESULTS": {"tesSUCCESS": 0},
		"hash": "` + definitionsHash + `"
	}`
	expected := server.DefinitionsResponse{
		Fields:             []server.DefinitionField{{Name: "TransactionType", Info: server.DefinitionFieldInfo{Nth: 2, Type: "UInt16", IsSerialized: true, IsSigningField: true}}},
		Types:              map[string]int{"UInt16": 1},
		LedgerEntryTypes:   map[string]int{"AccountRoot": 97},
		TransactionTypes:   map[string]int{"Payment": 0},
		TransactionResults: map[string]int{"tesSUCCESS": 0},
		Hash:               definitionsHash,
	}
	return expected, fixture
}

func TestDefinitionsResponseLegacySerialize(t *testing.T) {
	value, payload := definitionsResponseLegacyFixture()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, payload, string(encoded))
}

func TestDefinitionsResponseLegacyJSONDecode(t *testing.T) {
	want, payload := definitionsResponseLegacyFixture()
	var got server.DefinitionsResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestDefinitionsResponseLegacyClientDecode(t *testing.T) {
	want, payload := definitionsResponseLegacyFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got server.DefinitionsResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}

func TestDefinitionsResponseAcceptsIndependentEnhancedSections(t *testing.T) {
	clearEnhancedSections := func(r *server.DefinitionsResponse) {
		r.LedgerEntryFormats = nil
		r.TransactionFormats = nil
		r.LedgerEntryFlags = nil
		r.TransactionFlags = nil
		r.AccountSetFlags = nil
	}
	sections := []struct {
		name string
		keep func(*server.DefinitionsResponse)
	}{
		{
			name: "ledger entry formats",
			keep: func(r *server.DefinitionsResponse) {
				value := r.LedgerEntryFormats
				clearEnhancedSections(r)
				r.LedgerEntryFormats = value
			},
		},
		{
			name: "transaction formats",
			keep: func(r *server.DefinitionsResponse) {
				value := r.TransactionFormats
				clearEnhancedSections(r)
				r.TransactionFormats = value
			},
		},
		{
			name: "ledger entry flags",
			keep: func(r *server.DefinitionsResponse) {
				value := r.LedgerEntryFlags
				clearEnhancedSections(r)
				r.LedgerEntryFlags = value
			},
		},
		{
			name: "transaction flags",
			keep: func(r *server.DefinitionsResponse) {
				value := r.TransactionFlags
				clearEnhancedSections(r)
				r.TransactionFlags = value
			},
		},
		{
			name: "account set flags",
			keep: func(r *server.DefinitionsResponse) {
				value := r.AccountSetFlags
				clearEnhancedSections(r)
				r.AccountSetFlags = value
			},
		},
	}

	for _, section := range sections {
		t.Run(section.name, func(t *testing.T) {
			response, _ := definitionsResponseFullFixture()
			section.keep(&response)
			require.NoError(t, response.Validate())

			encoded, err := json.Marshal(response)
			require.NoError(t, err)

			var decoded server.DefinitionsResponse
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			require.Equal(t, response, decoded)
		})
	}
}

func definitionsResponseHashOnlyFixture() (server.DefinitionsResponse, string) {
	fixture := `{"hash":"` + definitionsHash + `"}`
	expected := server.DefinitionsResponse{Hash: definitionsHash}
	return expected, fixture
}

func TestDefinitionsResponseHashOnlySerialize(t *testing.T) {
	value, payload := definitionsResponseHashOnlyFixture()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.JSONEq(t, payload, string(encoded))
}

func TestDefinitionsResponseHashOnlyJSONDecode(t *testing.T) {
	want, payload := definitionsResponseHashOnlyFixture()
	var got server.DefinitionsResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestDefinitionsResponseHashOnlyClientDecode(t *testing.T) {
	want, payload := definitionsResponseHashOnlyFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got server.DefinitionsResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}

func TestDefinitionsResponseRejectsInvalidJSON(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		wantErr error
	}{
		{
			name:    "partial full response",
			fixture: `{"TYPES":{"UInt16":1},"hash":"` + definitionsHash + `"}`,
			wantErr: server.ErrInvalidDefinitionsResponse,
		},
		{
			name:    "null definition section",
			fixture: `{"FIELDS":null,"hash":"` + definitionsHash + `"}`,
			wantErr: server.ErrInvalidDefinitionsResponse,
		},
		{
			name:    "missing field property",
			fixture: strings.Replace(fullDefinitionsFixture, `"nth": -1, `, "", 1),
			wantErr: server.ErrInvalidDefinitionField,
		},
		{
			name:    "missing format optionality",
			fixture: strings.Replace(fullDefinitionsFixture, `, "optionality": 0`, "", 1),
			wantErr: server.ErrInvalidDefinitionsResponse,
		},
		{
			name:    "missing hash",
			fixture: `{"FIELDS":[]}`,
			wantErr: server.ErrInvalidDefinitionsHash,
		},
		{
			name:    "wrong section type with response sentinel",
			fixture: `{"TYPES":[],"hash":"` + definitionsHash + `"}`,
			wantErr: server.ErrInvalidDefinitionsResponse,
		},
		{
			name: "malformed field tuple",
			fixture: `{
				"FIELDS": [["TransactionType"]],
				"TYPES": {"UInt16": 1},
				"LEDGER_ENTRY_TYPES": {"AccountRoot": 97},
				"TRANSACTION_TYPES": {"Payment": 0},
				"TRANSACTION_RESULTS": {"tesSUCCESS": 0},
				"hash": "` + definitionsHash + `"
			}`,
			wantErr: server.ErrInvalidDefinitionField,
		},
		{
			name: "invalid optionality",
			fixture: `{
				"FIELDS": [["TransactionType", {"isSerialized": true, "isSigningField": true, "isVLEncoded": false, "nth": 2, "type": "UInt16"}]],
				"TYPES": {"UInt16": 1},
				"LEDGER_ENTRY_TYPES": {"AccountRoot": 97},
				"TRANSACTION_TYPES": {"Payment": 0},
				"TRANSACTION_RESULTS": {"tesSUCCESS": 0},
				"TRANSACTION_FORMATS": {"Payment": [{"name": "Destination", "optionality": 3}]},
				"hash": "` + definitionsHash + `"
			}`,
			wantErr: server.ErrInvalidDefinitionsResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response server.DefinitionsResponse
			require.ErrorIs(t, json.Unmarshal([]byte(tt.fixture), &response), tt.wantErr)
		})
	}
}

func TestDefinitionsResponseValidateForRequest(t *testing.T) {
	fullResponse, _ := definitionsResponseFullFixture()
	matchingLowercaseHash := strings.ToLower(definitionsHash)
	differentHash := "A685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"

	tests := []struct {
		name     string
		response server.DefinitionsResponse
		request  *server.DefinitionsRequest
		wantErr  error
	}{
		{
			name:     "full response does not require request hash",
			response: fullResponse,
			request:  &server.DefinitionsRequest{},
		},
		{
			name:     "matching hash ignores case",
			response: server.DefinitionsResponse{Hash: definitionsHash},
			request:  &server.DefinitionsRequest{Hash: matchingLowercaseHash},
		},
		{
			name:     "reject hash-only response without request hash",
			response: server.DefinitionsResponse{Hash: definitionsHash},
			request:  &server.DefinitionsRequest{},
			wantErr:  server.ErrInvalidDefinitionsResponse,
		},
		{
			name:     "reject hash-only response with different hash",
			response: server.DefinitionsResponse{Hash: definitionsHash},
			request:  &server.DefinitionsRequest{Hash: differentHash},
			wantErr:  server.ErrInvalidDefinitionsResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.response.ValidateForRequest(tt.request)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
