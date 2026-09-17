package websocket

import (
	"encoding/json"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

const websocketSimulateTxBlob = "120000240000ACA461400000000000000168400000000000000A730074008114B5F762798A53D543A014CAF8B297CFF8F2F937E88314550FC62003E785DC231A1058A05E56E3F09CF4E6"

func TestClient_Simulate(t *testing.T) {
	// Response fields are tested at the model. These cases protect request forwarding
	// and network handling for JSON/blob input and JSON/binary output.
	for _, tt := range []struct {
		name            string
		request         *transactions.SimulateRequest
		expectedRequest string
	}{
		{
			name:            "explicit JSON NetworkID",
			request:         &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "NetworkID": uint32(2048)}},
			expectedRequest: `{"tx_json":{"TransactionType":"Payment","Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","NetworkID":2048}}`,
		},
		{
			name:            "blob input and binary output",
			request:         &transactions.SimulateRequest{TxBlob: websocketSimulateTxBlob, Binary: true},
			expectedRequest: `{"tx_blob":"` + websocketSimulateTxBlob + `","binary":true}`,
		},
		{
			name:            "JSON input leaves NetworkID to server",
			request:         &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"}, Binary: true},
			expectedRequest: `{"tx_json":{"TransactionType":"Payment","Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"},"binary":true}`,
		},
		{
			name:            "blob input and JSON output",
			request:         &transactions.SimulateRequest{TxBlob: websocketSimulateTxBlob},
			expectedRequest: `{"tx_blob":"` + websocketSimulateTxBlob + `"}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := &transactions.SimulateResponse{EngineResult: "tesSUCCESS", EngineResultMessage: "ok"}
			if tt.request.Binary {
				result.TxBlob = "00"
			} else {
				result.TxJSON = transaction.FlatTransaction{"TransactionType": "Payment"}
			}
			client, request := setupQueryTestClient(t, result)
			setTestNetworkIdentity(client, uint32Pointer(2048), "1.12.0")
			response, err := client.Simulate(tt.request)
			require.NoError(t, err)
			require.Equal(t, "tesSUCCESS", response.EngineResult)
			got := request()
			require.Equal(t, "simulate", got["command"])
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
	response, err := client.Simulate(&transactions.SimulateRequest{TxBlob: websocketSimulateTxBlob, Binary: true})
	require.ErrorIs(t, err, transactions.ErrInvalidSimulateResponse)
	require.Nil(t, response)
}

func TestClient_SimulateRejectsLocally(t *testing.T) {
	tests := []struct {
		name      string
		request   *transactions.SimulateRequest
		networkID uint32
		wantErr   error
	}{
		{
			name:    "nil request",
			wantErr: transactions.ErrInvalidSimulateRequest,
		},

		{
			name: "mismatched on restricted network",
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "NetworkID": uint32(2049),
			}},
			networkID: 2048,
			wantErr:   transactions.ErrMismatchedSimulateNetworkID,
		},
		{
			name: "mismatched on known Mainnet",
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "NetworkID": uint32(2048),
			}},
			networkID: 0,
			wantErr:   transactions.ErrMismatchedSimulateNetworkID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No connection: rejection must happen before the transport is used.
			client := NewClient(*NewClientConfig())
			setTestNetworkIdentity(client, uint32Pointer(tt.networkID), "")

			response, err := client.Simulate(tt.request)
			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, response)
		})
	}
}

func TestClient_SimulateDelegatesOpaqueBlob(t *testing.T) {
	client, cleanup := setupTestClient(t, []map[string]any{{"id": 1, "status": "error", "error": "invalidParams"}})
	defer cleanup()
	response, err := client.Simulate(&transactions.SimulateRequest{TxBlob: "E1"})
	require.EqualError(t, err, "invalidParams")
	require.Nil(t, response)
}
