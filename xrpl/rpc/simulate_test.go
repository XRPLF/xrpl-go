package rpc

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

const rpcSimulateTxBlob = "120000240000ACA461400000000000000168400000000000000A730074008114B5F762798A53D543A014CAF8B297CFF8F2F937E88314550FC62003E785DC231A1058A05E56E3F09CF4E6"

func TestClient_Simulate(t *testing.T) {
	// Cover request encoding and result decoding for each input/output mode.
	// Detailed response validation is tested at the model.
	for _, tt := range []struct {
		name            string
		request         *transactions.SimulateRequest
		expectedRequest string
	}{
		{
			name: "JSON input and JSON output",
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"Destination": "r3kmLJN5D28dHuH8vZNUZpMC43pEHpaocV", "Amount": "1", "NetworkID": uint32(2048),
			}},
			expectedRequest: `{"tx_json":{"TransactionType":"Payment","Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","Destination":"r3kmLJN5D28dHuH8vZNUZpMC43pEHpaocV","Amount":"1","NetworkID":2048}}`,
		},
		{
			name:            "blob input and binary output",
			request:         &transactions.SimulateRequest{TxBlob: rpcSimulateTxBlob, Binary: true},
			expectedRequest: `{"tx_blob":"` + rpcSimulateTxBlob + `","binary":true}`,
		},
		{
			name: "JSON input and binary output",
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"Destination": "r3kmLJN5D28dHuH8vZNUZpMC43pEHpaocV", "Amount": "1",
			}, Binary: true},
			expectedRequest: `{"tx_json":{"TransactionType":"Payment","Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","Destination":"r3kmLJN5D28dHuH8vZNUZpMC43pEHpaocV","Amount":"1"},"binary":true}`,
		},
		{
			name:            "blob input and JSON output",
			request:         &transactions.SimulateRequest{TxBlob: rpcSimulateTxBlob},
			expectedRequest: `{"tx_blob":"` + rpcSimulateTxBlob + `"}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := &transactions.SimulateResponse{EngineResult: "tesSUCCESS", EngineResultMessage: "ok"}
			if tt.request.Binary {
				result.TxBlob = "00"
			} else {
				result.TxJSON = transaction.FlatTransaction{"TransactionType": "Payment"}
			}
			body, err := json.Marshal(map[string]any{"result": result})
			require.NoError(t, err)
			mock := &testutil.JSONRPCMockClient{}
			mock.DoFunc = testutil.MockResponse(string(body), http.StatusOK, mock)
			config, err := NewClientConfig("http://testnode/", WithHTTPClient(mock), WithNetworkIdentity(2048, "1.12.0"))
			require.NoError(t, err)
			client := NewClient(config)
			response, err := client.Simulate(tt.request)
			require.NoError(t, err)
			require.Equal(t, result, response)
			var envelope struct {
				Method string           `json:"method"`
				Params []map[string]any `json:"params"`
			}
			require.NoError(t, json.NewDecoder(mock.Spy.Body).Decode(&envelope))
			require.Len(t, envelope.Params, 1)
			got := envelope.Params[0]
			got["command"] = envelope.Method
			require.Equal(t, "simulate", got["command"])
			require.InDelta(t, 2, got["api_version"], 0)
			delete(got, "command")
			delete(got, "api_version")
			delete(got, "id")
			encoded, err := json.Marshal(got)
			require.NoError(t, err)
			require.JSONEq(t, tt.expectedRequest, string(encoded))
		})
	}
}

func TestClient_SimulateRejectsMismatchedResponseMode(t *testing.T) {
	client, _ := setupQueryTestClient(t, &transactions.SimulateResponse{
		EngineResult: "tesSUCCESS", EngineResultMessage: "ok", TxJSON: transaction.FlatTransaction{"TransactionType": "Payment"},
	})
	response, err := client.Simulate(&transactions.SimulateRequest{TxBlob: rpcSimulateTxBlob, Binary: true})
	require.ErrorIs(t, err, transactions.ErrInvalidSimulateResponse)
	require.Nil(t, response)
}

func TestClient_SimulateRejectsNil(t *testing.T) {
	called := false
	mockClient := testutil.JSONRPCMockClient{}
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		called = true
		return testutil.MockResponse(`{"result":{"status":"success"}}`, http.StatusOK, &mockClient)(req)
	}
	config, err := NewClientConfig("http://testnode/", WithHTTPClient(&mockClient))
	require.NoError(t, err)
	client := NewClient(config)

	response, err := client.Simulate(nil)
	require.False(t, called, "nil requests must not reach the transport")
	require.ErrorIs(t, err, transactions.ErrInvalidSimulateRequest)
	require.Nil(t, response)
}

func TestClient_SimulateReturnsServerError(t *testing.T) {
	client, request := setupQueryTestClient(t, map[string]any{"error": "tooBusy", "status": "error"})
	response, err := client.Simulate(&transactions.SimulateRequest{TxBlob: rpcSimulateTxBlob})
	require.EqualError(t, err, "tooBusy")
	require.Nil(t, response)
	require.Equal(t, rpcSimulateTxBlob, request()["tx_blob"])
}
