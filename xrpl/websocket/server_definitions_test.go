package websocket

import (
	"testing"

	serverquery "github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/stretchr/testify/require"
)

func TestClient_FormatServerDefinitionsRequest(t *testing.T) {
	const hash = "C685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"
	tests := []struct {
		name     string
		request  *serverquery.DefinitionsRequest
		expected string
	}{
		{
			name:     "without hash",
			request:  &serverquery.DefinitionsRequest{},
			expected: `{"api_version":2,"command":"server_definitions","id":7}`,
		},
		{
			name:     "with hash",
			request:  &serverquery.DefinitionsRequest{Hash: hash},
			expected: `{"api_version":2,"command":"server_definitions","hash":"` + hash + `","id":7}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(*NewClientConfig())
			message, err := client.formatRequest(tt.request, 7, nil)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, string(message))
		})
	}
}

func TestClient_GetServerDefinitions(t *testing.T) {
	const hash = "C685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"

	tests := []struct {
		name            string
		message         map[string]any
		request         *serverquery.DefinitionsRequest
		expected        *serverquery.DefinitionsResponse
		expectedErr     error
		expectedErrText string
	}{
		{
			name: "full response",
			message: map[string]any{
				"id": 1,
				"result": map[string]any{
					"FIELDS": []any{
						[]any{"TransactionType", map[string]any{
							"isSerialized": true, "isSigningField": true, "isVLEncoded": false,
							"nth": 2, "type": "UInt16",
						}},
					},
					"TYPES":               map[string]any{"UInt16": 1},
					"LEDGER_ENTRY_TYPES":  map[string]any{"AccountRoot": 97},
					"TRANSACTION_TYPES":   map[string]any{"Payment": 0},
					"TRANSACTION_RESULTS": map[string]any{"tesSUCCESS": 0},
					"LEDGER_ENTRY_FORMATS": map[string]any{
						"AccountRoot": []any{map[string]any{"name": "Account", "optionality": 0}},
					},
					"TRANSACTION_FORMATS": map[string]any{
						"Payment": []any{map[string]any{"name": "Destination", "optionality": 0}},
					},
					"LEDGER_ENTRY_FLAGS": map[string]any{
						"AccountRoot": map[string]any{"lsfAllowTrustLineClawback": uint32(2147483648)},
					},
					"TRANSACTION_FLAGS": map[string]any{
						"Payment": map[string]any{"tfPartialPayment": uint32(131072)},
					},
					"ACCOUNT_SET_FLAGS": map[string]any{"asfDefaultRipple": uint32(8)},
					"hash":              hash,
				},
			},
			request: &serverquery.DefinitionsRequest{},
			expected: &serverquery.DefinitionsResponse{
				Fields: []serverquery.DefinitionField{{
					Name: "TransactionType",
					Info: serverquery.DefinitionFieldInfo{
						Nth:            2,
						IsSerialized:   true,
						IsSigningField: true,
						Type:           "UInt16",
					},
				}},
				Types:              map[string]int{"UInt16": 1},
				LedgerEntryTypes:   map[string]int{"AccountRoot": 97},
				TransactionTypes:   map[string]int{"Payment": 0},
				TransactionResults: map[string]int{"tesSUCCESS": 0},
				LedgerEntryFormats: map[string][]serverquery.DefinitionFormatField{
					"AccountRoot": {{Name: "Account", Optionality: 0}},
				},
				TransactionFormats: map[string][]serverquery.DefinitionFormatField{
					"Payment": {{Name: "Destination", Optionality: 0}},
				},
				LedgerEntryFlags: map[string]map[string]uint32{
					"AccountRoot": {"lsfAllowTrustLineClawback": 2147483648},
				},
				TransactionFlags: map[string]map[string]uint32{
					"Payment": {"tfPartialPayment": 131072},
				},
				AccountSetFlags: map[string]uint32{"asfDefaultRipple": 8},
				Hash:            hash,
			},
		},
		{
			name: "hash-only unchanged response",
			message: map[string]any{
				"id":     1,
				"result": map[string]any{"hash": hash},
			},
			request:  &serverquery.DefinitionsRequest{Hash: hash},
			expected: &serverquery.DefinitionsResponse{Hash: hash},
		},
		{
			name: "reject hash-only response without request hash",
			message: map[string]any{
				"id":     1,
				"result": map[string]any{"hash": hash},
			},
			request:     &serverquery.DefinitionsRequest{},
			expectedErr: serverquery.ErrInvalidDefinitionsResponse,
		},
		{
			name: "unsupported server error",
			message: map[string]any{
				"id":            1,
				"status":        "error",
				"type":          "response",
				"error":         "unknownCmd",
				"error_message": "Unknown method.",
			},
			request:         &serverquery.DefinitionsRequest{},
			expectedErrText: "unknownCmd",
		},
		{
			name: "client rejects malformed partial response",
			message: map[string]any{
				"id": 1,
				"result": map[string]any{
					"TYPES": map[string]any{"UInt16": 1},
					"hash":  hash,
				},
			},
			request:     &serverquery.DefinitionsRequest{},
			expectedErr: serverquery.ErrInvalidDefinitionsResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestClient(t, []map[string]any{tt.message})
			defer cleanup()

			response, err := client.GetServerDefinitions(tt.request)
			switch {
			case tt.expectedErr != nil:
				require.ErrorIs(t, err, tt.expectedErr)
				require.Nil(t, response)
			case tt.expectedErrText != "":
				require.EqualError(t, err, tt.expectedErrText)
				require.Nil(t, response)
			default:
				require.NoError(t, err)
				require.Equal(t, tt.expected, response)
			}
		})
	}
}
