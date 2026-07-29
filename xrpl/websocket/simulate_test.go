package websocket

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

const websocketSimulateTxBlob = "120000240000ACA461400000000000000168400000000000000A730074008114B5F762798A53D543A014CAF8B297CFF8F2F937E88314550FC62003E785DC231A1058A05E56E3F09CF4E6"

func TestClient_FormatSimulateRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  *transactions.SimulateRequest
		expected string
	}{
		{
			name: "JSON input preserves NetworkID",
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment",
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"NetworkID":       uint32(2048),
			}},
			expected: `{"api_version":2,"command":"simulate","id":7,"tx_json":{"TransactionType":"Payment","Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","NetworkID":2048}}`,
		},
		{
			name:     "blob input selects binary output",
			request:  &transactions.SimulateRequest{TxBlob: websocketSimulateTxBlob, Binary: true},
			expected: `{"api_version":2,"binary":true,"command":"simulate","id":7,"tx_blob":"` + websocketSimulateTxBlob + `"}`,
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

func TestClient_Simulate(t *testing.T) {
	tests := []struct {
		name            string
		message         map[string]any
		request         *transactions.SimulateRequest
		networkID       uint32
		expected        *transactions.SimulateResponse
		expectedErrText string
	}{
		{
			name: "JSON request and response preserves explicit NetworkID",
			message: map[string]any{
				"id": 1,
				"result": map[string]any{
					"applied":               false,
					"engine_result":         "tesSUCCESS",
					"engine_result_code":    0,
					"engine_result_message": "The simulated transaction would have been applied.",
					"ledger_index":          uint32(105935704),
					"meta": map[string]any{
						"AffectedNodes": []any{}, "TransactionIndex": 0, "TransactionResult": "tesSUCCESS",
					},
					"tx_json": map[string]any{
						"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "Fee": "10", "NetworkID": uint32(2048),
						"Sequence": uint32(44196), "SigningPubKey": "", "TransactionType": "Payment", "TxnSignature": "",
					},
				},
			},
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment",
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"NetworkID":       uint32(2048),
			}},
			networkID: 2048,
			expected: &transactions.SimulateResponse{
				Applied:             false,
				EngineResult:        "tesSUCCESS",
				EngineResultCode:    0,
				EngineResultMessage: "The simulated transaction would have been applied.",
				LedgerIndex:         common.LedgerIndex(105935704),
				Meta: &transaction.TxMetadataBuilder{
					AffectedNodes:     []transaction.AffectedNode{},
					TransactionResult: "tesSUCCESS",
				},
				TxJSON: transaction.FlatTransaction{
					"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"Fee":             "10",
					"NetworkID":       float64(2048),
					"Sequence":        float64(44196),
					"SigningPubKey":   "",
					"TransactionType": "Payment",
					"TxnSignature":    "",
				},
			},
		},
		{
			name: "blob request and binary response",
			message: map[string]any{
				"id": 1,
				"result": map[string]any{
					"applied":               false,
					"engine_result":         "tesSUCCESS",
					"engine_result_code":    0,
					"engine_result_message": "The simulated transaction would have been applied.",
					"ledger_index":          uint32(105935704),
					"meta_blob":             "201C0000003BF8E5110061250644",
					"tx_blob":               websocketSimulateTxBlob,
				},
			},
			request: &transactions.SimulateRequest{TxBlob: websocketSimulateTxBlob, Binary: true},
			expected: &transactions.SimulateResponse{
				Applied:             false,
				EngineResult:        "tesSUCCESS",
				EngineResultCode:    0,
				EngineResultMessage: "The simulated transaction would have been applied.",
				LedgerIndex:         common.LedgerIndex(105935704),
				TxBlob:              websocketSimulateTxBlob,
				MetaBlob:            "201C0000003BF8E5110061250644",
			},
		},
		{
			name: "non-tec response omits metadata",
			message: map[string]any{
				"id": 1,
				"result": map[string]any{
					"applied":               false,
					"engine_result":         "temREDUNDANT",
					"engine_result_code":    -275,
					"engine_result_message": "The transaction is redundant.",
					"ledger_index":          uint32(105935694),
					"tx_json": map[string]any{
						"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "TransactionType": "Payment",
					},
				},
			},
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			}},
			expected: &transactions.SimulateResponse{
				Applied:             false,
				EngineResult:        "temREDUNDANT",
				EngineResultCode:    -275,
				EngineResultMessage: "The transaction is redundant.",
				LedgerIndex:         common.LedgerIndex(105935694),
				TxJSON: transaction.FlatTransaction{
					"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "TransactionType": "Payment",
				},
			},
		},
		{
			name: "JSON request with binary response",
			message: map[string]any{
				"id": 1,
				"result": map[string]any{
					"applied":               false,
					"engine_result":         "temREDUNDANT",
					"engine_result_code":    -275,
					"engine_result_message": "The transaction is redundant.",
					"ledger_index":          uint32(105935704),
					"tx_blob":               websocketSimulateTxBlob,
				},
			},
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			}, Binary: true},
			expected: &transactions.SimulateResponse{
				Applied:             false,
				EngineResult:        "temREDUNDANT",
				EngineResultCode:    -275,
				EngineResultMessage: "The transaction is redundant.",
				LedgerIndex:         common.LedgerIndex(105935704),
				TxBlob:              websocketSimulateTxBlob,
			},
		},
		{
			name: "blob request with JSON response",
			message: map[string]any{
				"id": 1,
				"result": map[string]any{
					"applied":               false,
					"engine_result":         "temREDUNDANT",
					"engine_result_code":    -275,
					"engine_result_message": "The transaction is redundant.",
					"ledger_index":          uint32(105935704),
					"tx_json": map[string]any{
						"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "TransactionType": "Payment",
					},
				},
			},
			request: &transactions.SimulateRequest{TxBlob: websocketSimulateTxBlob},
			expected: &transactions.SimulateResponse{
				Applied:             false,
				EngineResult:        "temREDUNDANT",
				EngineResultCode:    -275,
				EngineResultMessage: "The transaction is redundant.",
				LedgerIndex:         common.LedgerIndex(105935704),
				TxJSON: transaction.FlatTransaction{
					"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "TransactionType": "Payment",
				},
			},
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
			request:         &transactions.SimulateRequest{TxBlob: websocketSimulateTxBlob},
			expectedErrText: "unknownCmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestClient(t, []map[string]any{tt.message})
			defer cleanup()
			client.NetworkID = tt.networkID

			response, err := client.Simulate(tt.request)
			if tt.expectedErrText != "" {
				require.EqualError(t, err, tt.expectedErrText)
				require.Nil(t, response)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, response)
		})
	}
}

func TestClient_SimulateRejectsNetworkIDLocally(t *testing.T) {
	tests := []struct {
		name    string
		request *transactions.SimulateRequest
		wantErr error
	}{
		{
			name: "missing on restricted network",
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			}},
			wantErr: transactions.ErrMissingSimulateNetworkID,
		},
		{
			name: "mismatched on restricted network",
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "NetworkID": uint32(2049),
			}},
			wantErr: transactions.ErrMismatchedSimulateNetworkID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No connection: rejection must happen before the transport is used.
			client := NewClient(*NewClientConfig())
			client.NetworkID = 2048

			response, err := client.Simulate(tt.request)
			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, response)
		})
	}
}
