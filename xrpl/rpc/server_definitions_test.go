package rpc

import (
	"io"
	"testing"

	serverquery "github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/stretchr/testify/require"
)

const serverDefinitionsTestHash = "C685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"

func TestClient_GetServerDefinitions(t *testing.T) {
	fullResponse := &serverquery.DefinitionsResponse{
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
		Hash:            serverDefinitionsTestHash,
	}

	tests := []struct {
		name            string
		mockResponse    string
		request         *serverquery.DefinitionsRequest
		expectedRequest string
		expected        *serverquery.DefinitionsResponse
		expectedErr     error
		expectedErrText string
	}{
		{
			name: "full response",
			mockResponse: `{"result":{
				"FIELDS":[["TransactionType",{"isSerialized":true,"isSigningField":true,"isVLEncoded":false,"nth":2,"type":"UInt16"}]],
				"TYPES":{"UInt16":1},
				"LEDGER_ENTRY_TYPES":{"AccountRoot":97},
				"TRANSACTION_TYPES":{"Payment":0},
				"TRANSACTION_RESULTS":{"tesSUCCESS":0},
				"LEDGER_ENTRY_FORMATS":{"AccountRoot":[{"name":"Account","optionality":0}]},
				"TRANSACTION_FORMATS":{"Payment":[{"name":"Destination","optionality":0}]},
				"LEDGER_ENTRY_FLAGS":{"AccountRoot":{"lsfAllowTrustLineClawback":2147483648}},
				"TRANSACTION_FLAGS":{"Payment":{"tfPartialPayment":131072}},
				"ACCOUNT_SET_FLAGS":{"asfDefaultRipple":8},
				"hash":"` + serverDefinitionsTestHash + `","status":"success"
			}}`,
			request:         &serverquery.DefinitionsRequest{},
			expectedRequest: `{"method":"server_definitions","params":[{"api_version":2}]}`,
			expected:        fullResponse,
		},
		{
			name:            "hash-only unchanged response",
			mockResponse:    `{"result":{"hash":"` + serverDefinitionsTestHash + `","status":"success"}}`,
			request:         &serverquery.DefinitionsRequest{Hash: serverDefinitionsTestHash},
			expectedRequest: `{"method":"server_definitions","params":[{"api_version":2,"hash":"` + serverDefinitionsTestHash + `"}]}`,
			expected:        &serverquery.DefinitionsResponse{Hash: serverDefinitionsTestHash},
		},
		{
			name:            "unsupported server error",
			mockResponse:    `{"result":{"error":"unknownCmd","error_message":"Unknown method.","status":"error"}}`,
			request:         &serverquery.DefinitionsRequest{},
			expectedRequest: `{"method":"server_definitions","params":[{"api_version":2}]}`,
			expectedErrText: "unknownCmd",
		},
		{
			name:            "client rejects malformed partial response",
			mockResponse:    `{"result":{"TYPES":{"UInt16":1},"hash":"` + serverDefinitionsTestHash + `"}}`,
			request:         &serverquery.DefinitionsRequest{},
			expectedRequest: `{"method":"server_definitions","params":[{"api_version":2}]}`,
			expectedErr:     serverquery.ErrInvalidDefinitionsResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := testutil.JSONRPCMockClient{}
			mockClient.DoFunc = testutil.MockResponse(tt.mockResponse, 200, &mockClient)

			config, err := NewClientConfig("http://testnode/", WithHTTPClient(&mockClient))
			require.NoError(t, err)

			response, err := NewClient(config).GetServerDefinitions(tt.request)
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

			requestBody, readErr := io.ReadAll(mockClient.Spy.Body)
			require.NoError(t, readErr)
			require.JSONEq(t, tt.expectedRequest, string(requestBody))
		})
	}
}
