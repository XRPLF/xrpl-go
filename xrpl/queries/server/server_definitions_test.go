package server

import (
	"encoding/json"
	"strings"
	"testing"

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
		request DefinitionsRequest
		wantErr error
	}{
		{name: "without hash"},
		{name: "with matching-length hash", request: DefinitionsRequest{Hash: definitionsHash}},
		{name: "reject short hash", request: DefinitionsRequest{Hash: "ABCD"}, wantErr: ErrInvalidDefinitionsHash},
		{name: "reject non-hex hash", request: DefinitionsRequest{Hash: "Z685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"}, wantErr: ErrInvalidDefinitionsHash},
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

func TestDefinitionsResponseJSONVariants(t *testing.T) {
	legacyFullFixture := `{
		"FIELDS": [["TransactionType", {"isSerialized": true, "isSigningField": true, "isVLEncoded": false, "nth": 2, "type": "UInt16"}]],
		"TYPES": {"UInt16": 1},
		"LEDGER_ENTRY_TYPES": {"AccountRoot": 97},
		"TRANSACTION_TYPES": {"Payment": 0},
		"TRANSACTION_RESULTS": {"tesSUCCESS": 0},
		"hash": "` + definitionsHash + `"
	}`

	tests := []struct {
		name       string
		fixture    string
		wantErr    error
		checkValue func(*testing.T, DefinitionsResponse)
	}{
		{
			name:    "full XLS-97 response",
			fixture: fullDefinitionsFixture,
			checkValue: func(t *testing.T, got DefinitionsResponse) {
				require.Len(t, got.Fields, 2)
				require.Equal(t, "TransactionType", got.Fields[1].Name)
				require.Equal(t, "UInt16", got.Fields[1].Info.Type)
				require.Equal(t, uint32(2147483648), got.LedgerEntryFlags["AccountRoot"]["lsfAllowTrustLineClawback"])
				require.Equal(t, 0, got.TransactionTypes["Payment"])
				require.Equal(t, definitionsHash, got.Hash)
			},
		},
		{
			name:    "legacy full response without XLS-97 maps",
			fixture: legacyFullFixture,
			checkValue: func(t *testing.T, got DefinitionsResponse) {
				require.Nil(t, got.LedgerEntryFormats)
				require.Nil(t, got.TransactionFlags)
			},
		},
		{
			name:    "hash-only unchanged response",
			fixture: `{"hash":"` + definitionsHash + `"}`,
			checkValue: func(t *testing.T, got DefinitionsResponse) {
				require.Equal(t, definitionsHash, got.Hash)
				require.Nil(t, got.Fields)
				require.Nil(t, got.Types)
			},
		},
		{
			name:    "reject partial full response",
			fixture: `{"TYPES":{"UInt16":1},"hash":"` + definitionsHash + `"}`,
			wantErr: ErrInvalidDefinitionsResponse,
		},
		{
			name:    "reject null definition section",
			fixture: `{"FIELDS":null,"hash":"` + definitionsHash + `"}`,
			wantErr: ErrInvalidDefinitionsResponse,
		},
		{
			name:    "reject missing field property",
			fixture: strings.Replace(fullDefinitionsFixture, `"nth": -1, `, "", 1),
			wantErr: ErrInvalidDefinitionField,
		},
		{
			name:    "reject missing format optionality",
			fixture: strings.Replace(fullDefinitionsFixture, `, "optionality": 0`, "", 1),
			wantErr: ErrInvalidDefinitionsResponse,
		},
		{
			name:    "reject missing hash",
			fixture: `{"FIELDS":[]}`,
			wantErr: ErrInvalidDefinitionsHash,
		},
		{
			name:    "reject wrong section type with response sentinel",
			fixture: `{"TYPES":[],"hash":"` + definitionsHash + `"}`,
			wantErr: ErrInvalidDefinitionsResponse,
		},
		{
			name: "reject malformed field tuple",
			fixture: `{
				"FIELDS": [["TransactionType"]],
				"TYPES": {"UInt16": 1},
				"LEDGER_ENTRY_TYPES": {"AccountRoot": 97},
				"TRANSACTION_TYPES": {"Payment": 0},
				"TRANSACTION_RESULTS": {"tesSUCCESS": 0},
				"hash": "` + definitionsHash + `"
			}`,
			wantErr: ErrInvalidDefinitionField,
		},
		{
			name: "reject invalid optionality",
			fixture: `{
				"FIELDS": [["TransactionType", {"isSerialized": true, "isSigningField": true, "isVLEncoded": false, "nth": 2, "type": "UInt16"}]],
				"TYPES": {"UInt16": 1},
				"LEDGER_ENTRY_TYPES": {"AccountRoot": 97},
				"TRANSACTION_TYPES": {"Payment": 0},
				"TRANSACTION_RESULTS": {"tesSUCCESS": 0},
				"TRANSACTION_FORMATS": {"Payment": [{"name": "Destination", "optionality": 3}]},
				"hash": "` + definitionsHash + `"
			}`,
			wantErr: ErrInvalidDefinitionsResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DefinitionsResponse
			err := json.Unmarshal([]byte(tt.fixture), &got)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, got.Validate())
			tt.checkValue(t, got)

			encoded, err := json.Marshal(got)
			require.NoError(t, err)
			var roundTrip DefinitionsResponse
			require.NoError(t, json.Unmarshal(encoded, &roundTrip))
			require.Equal(t, got, roundTrip)
		})
	}
}

func TestDefinitionsResponseValidateForRequest(t *testing.T) {
	matchingLowercaseHash := strings.ToLower(definitionsHash)
	differentHash := "A685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"

	tests := []struct {
		name     string
		response DefinitionsResponse
		request  *DefinitionsRequest
		wantErr  error
	}{
		{
			name:     "full response does not require request hash",
			response: mustDecodeDefinitionsResponse(t, fullDefinitionsFixture),
			request:  &DefinitionsRequest{},
		},
		{
			name:     "matching hash ignores case",
			response: DefinitionsResponse{Hash: definitionsHash},
			request:  &DefinitionsRequest{Hash: matchingLowercaseHash},
		},
		{
			name:     "reject hash-only response without request hash",
			response: DefinitionsResponse{Hash: definitionsHash},
			request:  &DefinitionsRequest{},
			wantErr:  ErrInvalidDefinitionsResponse,
		},
		{
			name:     "reject hash-only response with different hash",
			response: DefinitionsResponse{Hash: definitionsHash},
			request:  &DefinitionsRequest{Hash: differentHash},
			wantErr:  ErrInvalidDefinitionsResponse,
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

func mustDecodeDefinitionsResponse(t *testing.T, fixture string) DefinitionsResponse {
	t.Helper()
	var response DefinitionsResponse
	require.NoError(t, json.Unmarshal([]byte(fixture), &response))
	return response
}
