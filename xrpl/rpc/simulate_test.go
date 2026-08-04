package rpc

import (
	"io"
	"net/http"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

const rpcSimulateTxBlob = "120000240000ACA461400000000000000168400000000000000A730074008114B5F762798A53D543A014CAF8B297CFF8F2F937E88314550FC62003E785DC231A1058A05E56E3F09CF4E6"

func TestClient_Simulate(t *testing.T) {
	tests := []struct {
		name            string
		mockResponse    string
		request         *transactions.SimulateRequest
		networkID       uint32
		expectedRequest string
		expected        *transactions.SimulateResponse
		expectedErrText string
	}{
		{
			name: "JSON request and response preserves explicit NetworkID",
			mockResponse: `{"result":{
				"applied":false,
				"engine_result":"tesSUCCESS",
				"engine_result_code":0,
				"engine_result_message":"The simulated transaction would have been applied.",
				"ledger_index":105935704,
				"meta":{"AffectedNodes":[],"TransactionIndex":0,"TransactionResult":"tesSUCCESS"},
				"tx_json":{"Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","Fee":"10","NetworkID":2048,"Sequence":44196,"SigningPubKey":"","TransactionType":"Payment","TxnSignature":""},
				"status":"success"
			}}`,
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment",
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"NetworkID":       uint32(2048),
			}},
			networkID:       2048,
			expectedRequest: `{"method":"simulate","params":[{"api_version":2,"tx_json":{"TransactionType":"Payment","Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","NetworkID":2048}}]}`,
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
			name: "blob request allows server NetworkID autofill",
			mockResponse: `{"result":{
				"applied":false,
				"engine_result":"tesSUCCESS",
				"engine_result_code":0,
				"engine_result_message":"The simulated transaction would have been applied.",
				"ledger_index":105935704,
				"meta_blob":"201C0000003BF8E5110061250644",
				"tx_blob":"` + rpcSimulateTxBlob + `",
				"status":"success"
			}}`,
			request:         &transactions.SimulateRequest{TxBlob: rpcSimulateTxBlob, Binary: true},
			networkID:       2048,
			expectedRequest: `{"method":"simulate","params":[{"api_version":2,"tx_blob":"` + rpcSimulateTxBlob + `","binary":true}]}`,
			expected: &transactions.SimulateResponse{
				Applied:             false,
				EngineResult:        "tesSUCCESS",
				EngineResultCode:    0,
				EngineResultMessage: "The simulated transaction would have been applied.",
				LedgerIndex:         common.LedgerIndex(105935704),
				TxBlob:              rpcSimulateTxBlob,
				MetaBlob:            "201C0000003BF8E5110061250644",
			},
		},
		{
			name: "JSON request allows server NetworkID autofill",
			mockResponse: `{"result":{
				"applied":false,
				"engine_result":"temREDUNDANT",
				"engine_result_code":-275,
				"engine_result_message":"The transaction is redundant.",
				"ledger_index":105935704,
				"tx_blob":"` + rpcSimulateTxBlob + `",
				"status":"success"
			}}`,
			request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			}, Binary: true},
			networkID:       2048,
			expectedRequest: `{"method":"simulate","params":[{"api_version":2,"tx_json":{"TransactionType":"Payment","Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"},"binary":true}]}`,
			expected: &transactions.SimulateResponse{
				Applied:             false,
				EngineResult:        "temREDUNDANT",
				EngineResultCode:    -275,
				EngineResultMessage: "The transaction is redundant.",
				LedgerIndex:         common.LedgerIndex(105935704),
				TxBlob:              rpcSimulateTxBlob,
			},
		},
		{
			name: "blob request with JSON response",
			mockResponse: `{"result":{
				"applied":false,
				"engine_result":"temREDUNDANT",
				"engine_result_code":-275,
				"engine_result_message":"The transaction is redundant.",
				"ledger_index":105935704,
				"tx_json":{"Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","TransactionType":"Payment"},
				"status":"success"
			}}`,
			request:         &transactions.SimulateRequest{TxBlob: rpcSimulateTxBlob},
			expectedRequest: `{"method":"simulate","params":[{"api_version":2,"tx_blob":"` + rpcSimulateTxBlob + `"}]}`,
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
			name:            "opaque blob validation is delegated to server",
			mockResponse:    `{"result":{"error":"invalidParams","error_message":"Invalid field.","status":"error"}}`,
			request:         &transactions.SimulateRequest{TxBlob: "E1"},
			networkID:       2048,
			expectedRequest: `{"method":"simulate","params":[{"api_version":2,"tx_blob":"E1"}]}`,
			expectedErrText: "invalidParams",
		},
		{
			name:            "unsupported server error",
			mockResponse:    `{"result":{"error":"unknownCmd","error_message":"Unknown method.","status":"error"}}`,
			request:         &transactions.SimulateRequest{TxBlob: rpcSimulateTxBlob},
			expectedRequest: `{"method":"simulate","params":[{"api_version":2,"tx_blob":"` + rpcSimulateTxBlob + `"}]}`,
			expectedErrText: "unknownCmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := testutil.JSONRPCMockClient{}
			mockClient.DoFunc = testutil.MockResponse(tt.mockResponse, 200, &mockClient)
			config, err := NewClientConfig("http://testnode/", WithHTTPClient(&mockClient))
			require.NoError(t, err)
			client := NewClient(config)
			client.NetworkID = tt.networkID

			response, err := client.Simulate(tt.request)
			if tt.expectedErrText != "" {
				require.EqualError(t, err, tt.expectedErrText)
				require.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, response)
			}

			requestBody, readErr := io.ReadAll(mockClient.Spy.Body)
			require.NoError(t, readErr)
			require.JSONEq(t, tt.expectedRequest, string(requestBody))
		})
	}
}

func TestClient_SimulateRejectsLocally(t *testing.T) {
	tests := []struct {
		name      string
		request   *transactions.SimulateRequest
		networkID uint32
		wantErr   error
	}{
		{name: "nil request", wantErr: transactions.ErrInvalidSimulateRequest},
		{name: "both inputs", request: &transactions.SimulateRequest{
			TxJSON: transaction.FlatTransaction{"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"},
			TxBlob: rpcSimulateTxBlob,
		}, wantErr: transactions.ErrInvalidSimulateRequest},
		{name: "neither input", request: &transactions.SimulateRequest{}, wantErr: transactions.ErrInvalidSimulateRequest},
		{name: "signed JSON", request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "TxnSignature": "DEADBEEF",
		}}, wantErr: transactions.ErrSignedSimulateTransaction},
		{name: "network ID mismatch", request: &transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "NetworkID": uint32(2049),
		}}, networkID: 2048, wantErr: transactions.ErrMismatchedSimulateNetworkID},
		{name: "non-hex blob", request: &transactions.SimulateRequest{TxBlob: "not-hex"}, networkID: 2048, wantErr: transactions.ErrInvalidSimulateTxBlob},
		{name: "odd-length blob", request: &transactions.SimulateRequest{TxBlob: "ABC"}, networkID: 2048, wantErr: transactions.ErrInvalidSimulateTxBlob},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			mockClient := testutil.JSONRPCMockClient{}
			mockClient.DoFunc = func(*http.Request) (*http.Response, error) {
				called = true
				return nil, nil
			}
			config, err := NewClientConfig("http://testnode/", WithHTTPClient(&mockClient))
			require.NoError(t, err)
			client := NewClient(config)
			client.NetworkID = tt.networkID

			response, err := client.Simulate(tt.request)
			require.ErrorIs(t, err, tt.wantErr)
			require.Nil(t, response)
			require.False(t, called, "locally invalid requests must not reach the transport")
		})
	}
}
