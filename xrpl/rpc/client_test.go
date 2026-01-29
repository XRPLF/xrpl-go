package rpc

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	account "github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	requests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	rpctypes "github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {

	t.Run("Set config with valid port + ip", func(t *testing.T) {

		cfg, _ := NewClientConfig("url")

		jsonRpcClient := NewClient(cfg)

		assert.Equal(t, &Client{cfg: cfg}, jsonRpcClient)
	})
}

func TestClient_Request(t *testing.T) {

	t.Run("SendRequest - Check headers and URL", func(t *testing.T) {

		req := &account.ChannelsRequest{
			Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		}
		var capturedRequest *http.Request

		mc := &testutil.JSONRPCMockClient{}
		mc.DoFunc = func(req *http.Request) (*http.Response, error) {
			capturedRequest = req
			return testutil.MockResponse(`{}`, 200, mc)(req)
		}

		cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
		assert.NoError(t, err)

		jsonRpcClient := NewClient(cfg)

		_, err = jsonRpcClient.Request(req)

		assert.NotNil(t, capturedRequest)
		assert.NoError(t, err)
		assert.Equal(t, "POST", capturedRequest.Method)
		assert.Equal(t, "http://testnode/", capturedRequest.URL.String())
		assert.Equal(t, "application/json", capturedRequest.Header.Get("Content-Type"))
	})

	t.Run("SendRequest - sucessful response", func(t *testing.T) {

		req := &account.ChannelsRequest{
			Account:            "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
			DestinationAccount: "rnZvsWuLem5Ha46AZs61jLWR9R5esinkG3",
			LedgerIndex:        common.Validated,
		}

		response := `{
			"result": {
			  "account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			  "channels": [
				{
					"account":             "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					"amount":              "1000",
					"balance":             "0",
					"channel_id":          "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
					"destination_account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					"public_key":          "aBR7mdD75Ycs8DRhMgQ4EMUEmBArF8SEh1hfjrT2V9DQTLNbJVqw",
					"public_key_hex":      "03CFD18E689434F032A4E84C63E2A3A6472D684EAF4FD52CA67742F3E24BAE81B2",
					"settle_delay":        60
				}
			  ],
			  "ledger_hash": "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
			  "ledger_index": 71766314,
			  "validated": true
			},
			"warning": "none",
			"warnings":
			[{
				"id": 1,
				"message": "message"
			}]
		  }`

		mc := &testutil.JSONRPCMockClient{}
		mc.DoFunc = testutil.MockResponse(response, 200, mc)

		cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
		assert.NoError(t, err)

		jsonRpcClient := NewClient(cfg)

		xrplResponse, err := jsonRpcClient.Request(req)

		expectedXrplResponse := &Response{
			Result: AnyJSON{
				"account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"channels": []any{
					map[string]any{
						"account":             "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
						"amount":              "1000",
						"balance":             "0",
						"channel_id":          "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
						"destination_account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
						"public_key":          "aBR7mdD75Ycs8DRhMgQ4EMUEmBArF8SEh1hfjrT2V9DQTLNbJVqw",
						"public_key_hex":      "03CFD18E689434F032A4E84C63E2A3A6472D684EAF4FD52CA67742F3E24BAE81B2",
						"settle_delay":        json.Number("60"),
					},
				},
				"ledger_hash":  "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
				"ledger_index": json.Number("71766314"),
				"validated":    true,
			},
			Warning: "none",
			Warnings: []XRPLResponseWarning{{
				ID:      1,
				Message: "message",
			},
			},
		}

		var channelsResponse account.ChannelsResponse
		_ = xrplResponse.GetResult(&channelsResponse)

		expected := &account.ChannelsResponse{
			Account:     "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			LedgerIndex: 71766314,
			LedgerHash:  "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
		}

		assert.NoError(t, err)

		assert.Equal(t, expectedXrplResponse, xrplResponse)

		assert.Equal(t, expected.Account, channelsResponse.Account)
		assert.Equal(t, expected.LedgerIndex, channelsResponse.LedgerIndex)
		assert.Equal(t, expected.LedgerHash, channelsResponse.LedgerHash)
	})

	t.Run("SendRequest - error response", func(t *testing.T) {

		req := &account.ChannelsRequest{
			Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		}
		response := `{
			"result": {
				"error": "ledgerIndexMalformed",
				"request": {
					"account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					"command": "account_info",
					"ledger_index": "-",
					"strict": true
				},
				"status": "error"
			}
		}`

		mc := &testutil.JSONRPCMockClient{}
		mc.DoFunc = testutil.MockResponse(response, 200, mc)

		cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
		assert.NoError(t, err)

		jsonRpcClient := NewClient(cfg)

		_, err = jsonRpcClient.Request(req)

		assert.EqualError(t, err, "ledgerIndexMalformed")
	})

	t.Run("SendRequest - 503 response", func(t *testing.T) {

		req := &account.ChannelsRequest{
			Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		}
		response := `Service Unavailable`

		mc := &testutil.JSONRPCMockClient{}
		mc.DoFunc = func(req *http.Request) (*http.Response, error) {
			mc.RequestCount++
			return testutil.MockResponse(response, 503, mc)(req)
		}

		cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
		assert.NoError(t, err)

		jsonRpcClient := NewClient(cfg)

		_, err = jsonRpcClient.Request(req)

		// Check that 3 extra requests were made
		assert.Equal(t, 4, mc.RequestCount)
		assert.EqualError(t, err, "Server is overloaded, rate limit exceeded")

	})

	t.Run("SendRequest - 503 response sucessfully resolves", func(t *testing.T) {

		req := &account.ChannelsRequest{
			Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		}
		sucessResponse := `{
			"result": {
			  "account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			  "ledger_hash": "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
			  "ledger_index": 71766343
				}
			}`

		mc := &testutil.JSONRPCMockClient{}
		mc.DoFunc = func(req *http.Request) (*http.Response, error) {
			if mc.RequestCount < 3 {
				// Return 503 response for the first three requests
				mc.RequestCount++
				return testutil.MockResponse(`Service Unavailable`, 503, mc)(req)
			}
			// Return 200 response for the fourth request
			return testutil.MockResponse(sucessResponse, 200, mc)(req)
		}

		cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
		assert.NoError(t, err)

		jsonRpcClient := NewClient(cfg)

		xrplResponse, err := jsonRpcClient.Request(req)

		var channelsResponse account.ChannelsResponse
		_ = xrplResponse.GetResult(&channelsResponse)

		expected := &account.ChannelsResponse{
			Account:     "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			LedgerIndex: 71766343,
			LedgerHash:  "27F530E5C93ED5C13994812787C1ED073C822BAEC7597964608F2C049C2ACD2D",
		}

		// Check that only 2 extra requests were made
		assert.Equal(t, 3, mc.RequestCount)

		assert.NoError(t, err)
		assert.Equal(t, expected.Account, channelsResponse.Account)
		assert.Equal(t, expected.LedgerIndex, channelsResponse.LedgerIndex)
		assert.Equal(t, expected.LedgerHash, channelsResponse.LedgerHash)
	})

	t.Run("SendRequest - timeout", func(t *testing.T) {
		req := &account.ChannelsRequest{
			Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		}

		mc := &testutil.JSONRPCMockClient{}
		mc.DoFunc = func(req *http.Request) (*http.Response, error) {
			// hit the timeout by not responding
			time.Sleep(time.Second * 5)
			return nil, errors.New("timeout")
		}

		cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
		assert.NoError(t, err)

		jsonRpcClient := NewClient(cfg)

		_, err = jsonRpcClient.Request(req)

		// Check that the expected timeout error occurred
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})
}

func TestClient_SubmitTxBlob(t *testing.T) {
	// We'll run two sets of subtests: one for SubmitTxBlob and one for SubmitTx.
	tests := []struct {
		name         string
		mockResponse string
		txBlob       string
		expectError  error
		expectResult *requests.SubmitResponse
	}{
		{
			name: "success",
			mockResponse: `{
					"result": {
						"engine_result": "tesSUCCESS",
						"engine_result_code": 0,
						"engine_result_message": "The transaction was applied.",
						"tx_blob": "1200002280000000240000000361D4838D7EA4C6800000000000000000000000000055534400000000004B4E9C06F24296074F7BC48F92A97916C6DC5EA968400000000000000A732103AB40A0490F9B7ED8DF29D246BF2D6269820A0EE7742ACDD457BEA7C7D0931EDB74473045022100D184EB4AE5956FF600E7536EE459345C7BBCF097A84CC61A93B9AF7197EDB98702201CEA8009B7BEEBAA2AACC0359B41C427C1C5B550A4CA4B80CF2174AF2D6D5DCE81144B4E9C06F24296074F7BC48F92A97916C6DC5EA983143E9D4A2B8AA0780F682D136F7A56D6724EF53754"
					},
					"status": "success",
					"type": "response"
				}`,
			txBlob:      "1200002280000000240000000361D4838D7EA4C6800000000000000000000000000055534400000000004B4E9C06F24296074F7BC48F92A97916C6DC5EA968400000000000000A732103AB40A0490F9B7ED8DF29D246BF2D6269820A0EE7742ACDD457BEA7C7D0931EDB74473045022100D184EB4AE5956FF600E7536EE459345C7BBCF097A84CC61A93B9AF7197EDB98702201CEA8009B7BEEBAA2AACC0359B41C427C1C5B550A4CA4B80CF2174AF2D6D5DCE81144B4E9C06F24296074F7BC48F92A97916C6DC5EA983143E9D4A2B8AA0780F682D136F7A56D6724EF53754",
			expectError: nil,
			expectResult: &requests.SubmitResponse{
				EngineResult:        "tesSUCCESS",
				EngineResultCode:    0,
				EngineResultMessage: "The transaction was applied.",
			},
		},
		{
			name:        "missing signature",
			txBlob:      "1200002280000000240000000361D4838D7EA4C6800000000000000000000000000055534400000000004B4E9C06F24296074F7BC48F92A97916C6DC5EA968400000000000000A70",
			expectError: errors.New("ReadField error: parser out of bounds"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock HTTP client.
			mc := &testutil.JSONRPCMockClient{}
			if tt.mockResponse != "" {
				mc.DoFunc = testutil.MockResponse(tt.mockResponse, 200, mc)
			}

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
			assert.NoError(t, err)

			jsonRpcClient := NewClient(cfg)

			response, err := jsonRpcClient.SubmitTxBlob(tt.txBlob, false)
			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectResult.EngineResult, response.EngineResult)
			assert.Equal(t, tt.expectResult.EngineResultCode, response.EngineResultCode)
			assert.Equal(t, tt.expectResult.EngineResultMessage, response.EngineResultMessage)
		})
	}
}

func TestClient_SubmitTx(t *testing.T) {

	tests := []struct {
		name         string
		mockResponse string
		tx           map[string]interface{}
		opts         *rpctypes.SubmitOptions
		expectError  error
		expectResult *requests.SubmitResponse
	}{
		{
			name: "pass - tx already signed",
			mockResponse: `{
		"result": {
			"engine_result": "tesSUCCESS",
			"engine_result_code": 0,
			"engine_result_message": "The transaction was applied.",
			"tx_blob": "dummyBlob"
		},
		"status": "success",
		"type": "response"
	}`,
			tx: map[string]interface{}{
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Destination":     "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"Fee":             "10",
				"TransactionType": "Payment",
				"SigningPubKey":   "03AB40A0490F9B7ED8DF29D246BF2D6269820A0EE7742ACDD457BEA7C7D0931EDB",
				"TxnSignature":    "3045022100D184EB4AE5956FF600E7536EE459345C7BBCF097A84CC61A93B9AF7197EDB98702201CEA8009B7BEEBAA2AACC0359B41C427C1C5B550A4CA4B80CF2174AF2D6D5DCE",
				"Sequence":        uint32(359),
			},
			opts: &rpctypes.SubmitOptions{
				Autofill: false,
				FailHard: false,
			},
			expectError: nil,
			expectResult: &requests.SubmitResponse{
				EngineResult:        "tesSUCCESS",
				EngineResultCode:    0,
				EngineResultMessage: "The transaction was applied.",
			},
		},
		{
			name: "fail - no wallet provided for unsigned tx",
			tx: map[string]interface{}{
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Destination":     "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"Fee":             "10",
				"TransactionType": "Payment",
				"Sequence":        uint32(359),
			},
			opts:        &rpctypes.SubmitOptions{},
			expectError: ErrMissingWallet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &testutil.JSONRPCMockClient{}
			if tt.mockResponse != "" {
				mc.DoFunc = testutil.MockResponse(tt.mockResponse, 200, mc)
			}

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
			assert.NoError(t, err)
			jsonRpcClient := NewClient(cfg)

			response, err := jsonRpcClient.SubmitTx(tt.tx, tt.opts)
			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectResult.EngineResult, response.EngineResult)
			assert.Equal(t, tt.expectResult.EngineResultCode, response.EngineResultCode)
			assert.Equal(t, tt.expectResult.EngineResultMessage, response.EngineResultMessage)
		})
	}
}

func TestClient_SubmitMultisigned(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse string
		txBlob       string
		expectError  error
		expectResult *requests.SubmitResponse
	}{
		{
			name: "successful multisign submit",
			mockResponse: `{
				"result": {
					"engine_result": "tesSUCCESS",
					"engine_result_code": 0,
					"engine_result_message": "The transaction was applied.",
					"tx_blob": "1200002280000000240000000361D4838D7EA4C6800000000000000000000000000055534400000000004B4E9C06F24296074F7BC48F92A97916C6DC5EA968400000000000000A732103AB40A0490F9B7ED8DF29D246BF2D6269820A0EE7742ACDD457BEA7C7D0931EDB74473045022100D184EB4AE5956FF600E7536EE459345C7BBCF097A84CC61A93B9AF7197EDB98702201CEA8009B7BEEBAA2AACC0359B41C427C1C5B550A4CA4B80CF2174AF2D6D5DCE81144B4E9C06F24296074F7BC48F92A97916C6DC5EA983143E9D4A2B8AA0780F682D136F7A56D6724EF53754",
					"tx_json": {
						"Account": "rUAi7pipxGpYfPNg3LtPcf2ApiS8aw9A93",
						"Fee": "10",
						"Flags": 2147483648,
						"Sequence": 4,
						"SigningPubKey": "",
						"TransactionType": "Payment",
						"TxnSignature": "3045022100D184EB4AE5956FF600E7536EE459345C7BBCF097A84CC61A93B9AF7197EDB98702201CEA8009B7BEEBAA2AACC0359B41C427C1C5B550A4CA4B80CF2174AF2D6D5DCE",
						"hash": "4D5D90890F8D49519E4151938601EF3D0B30B16CD6A519D9C99102C9FA77F7E0"
					}
				},
				"status": "success",
				"type": "response"
			}`,
			txBlob:      "1200002280000000240000000361D4838D7EA4C6800000000000000000000000000055534400000000004B4E9C06F24296074F7BC48F92A97916C6DC5EA968400000000000000A732103AB40A0490F9B7ED8DF29D246BF2D6269820A0EE7742ACDD457BEA7C7D0931EDB74473045022100D184EB4AE5956FF600E7536EE459345C7BBCF097A84CC61A93B9AF7197EDB98702201CEA8009B7BEEBAA2AACC0359B41C427C1C5B550A4CA4B80CF2174AF2D6D5DCE81144B4E9C06F24296074F7BC48F92A97916C6DC5EA983143E9D4A2B8AA0780F682D136F7A56D6724EF53754",
			expectError: nil,
			expectResult: &requests.SubmitResponse{
				EngineResult:        "tesSUCCESS",
				EngineResultCode:    0,
				EngineResultMessage: "The transaction was applied.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &testutil.JSONRPCMockClient{}
			if tt.mockResponse != "" {
				mc.DoFunc = testutil.MockResponse(tt.mockResponse, 200, mc)
			}

			cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
			require.NoError(t, err)

			jsonRpcClient := NewClient(cfg)

			response, err := jsonRpcClient.SubmitMultisigned(tt.txBlob, false)

			if tt.expectError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectError.Error(), err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectResult.EngineResult, response.EngineResult)
			require.Equal(t, tt.expectResult.EngineResultCode, response.EngineResultCode)
			require.Equal(t, tt.expectResult.EngineResultMessage, response.EngineResultMessage)
		})
	}
}

func TestClient_autofillRawTransactions(t *testing.T) {
	tests := []struct {
		name          string
		tx            transaction.FlatTransaction
		mockResponses []string
		networkID     uint32
		expectedTx    transaction.FlatTransaction
		expectedErr   error
	}{
		{
			name: "pass - valid single transaction autofill",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
						},
					},
				},
			},
			mockResponses: []string{
				`{
					"result": {
						"account_data": {
							"Sequence": 42
						}
					}
				}`,
			},
			networkID: 0,
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"Fee":             "0",
							"SigningPubKey":   "",
							"Sequence":        uint32(43), // 42 + 1 since same account
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "pass - multiple transactions with different accounts",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
						},
					},
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Destination":     "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Amount":          "2000000",
						},
					},
				},
			},
			mockResponses: []string{
				`{
					"result": {
						"account_data": {
							"Sequence": 42
						}
					}
				}`,
				`{
					"result": {
						"account_data": {
							"Sequence": 100
						}
					}
				}`,
			},
			networkID: 0,
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"Fee":             "0",
							"SigningPubKey":   "",
							"Sequence":        uint32(43), // 42 + 1 since same account
						},
					},
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Destination":     "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Amount":          "2000000",
							"Fee":             "0",
							"SigningPubKey":   "",
							"Sequence":        uint32(100), // Different account, use actual sequence
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "pass - transaction with TicketSequence - no Sequence needed",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"TicketSequence":  uint32(100),
						},
					},
				},
			},
			mockResponses: []string{}, // No API calls needed
			networkID:     0,
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"TicketSequence":  uint32(100),
							"Fee":             "0",
							"SigningPubKey":   "",
						},
					},
				},
			},
			expectedErr: nil,
		},
		// Error cases
		{
			name: "fail - RawTransactions field not an array",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": "not_an_array",
			},
			mockResponses: []string{},
			networkID:     0,
			expectedTx:    transaction.FlatTransaction{},
			expectedErr:   ErrRawTransactionsFieldIsNotAnArray,
		},
		{
			name: "fail - RawTransaction field not an object",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": "not_an_object",
					},
				},
			},
			mockResponses: []string{},
			networkID:     0,
			expectedTx:    transaction.FlatTransaction{},
			expectedErr:   ErrRawTransactionFieldIsNotAnObject,
		},
		{
			name: "fail - Fee field set to non-zero value - error",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Fee":             "10",
						},
					},
				},
			},
			mockResponses: []string{},
			networkID:     0,
			expectedTx:    transaction.FlatTransaction{},
			expectedErr:   types.ErrBatchInnerTransactionInvalid,
		},
		{
			name: "fail - SigningPubKey field set to non-empty value - error",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"SigningPubKey":   "03ABC123",
						},
					},
				},
			},
			mockResponses: []string{},
			networkID:     0,
			expectedTx:    transaction.FlatTransaction{},
			expectedErr:   ErrSigningPubKeyFieldMustBeEmpty,
		},
		{
			name: "fail - TxnSignature field present - error",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"TxnSignature":    "304502",
						},
					},
				},
			},
			mockResponses: []string{},
			networkID:     0,
			expectedTx:    transaction.FlatTransaction{},
			expectedErr:   ErrTxnSignatureFieldMustBeEmpty,
		},
		{
			name: "fail - Signers field present - error",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Signers":         []any{},
						},
					},
				},
			},
			mockResponses: []string{},
			networkID:     0,
			expectedTx:    transaction.FlatTransaction{},
			expectedErr:   ErrSignersFieldMustBeEmpty,
		},
		{
			name: "fail - Account field not a string - error",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         12345, // Invalid: not a string
						},
					},
				},
			},
			mockResponses: []string{},
			networkID:     0,
			expectedTx:    transaction.FlatTransaction{},
			expectedErr:   ErrAccountFieldIsNotAString,
		},
		{
			name: "fail - Error from GetAccountInfo",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
						},
					},
				},
			},
			mockResponses: []string{
				`{
					"result": {
						"error": "actNotFound"
					}
				}`,
			},
			networkID:   0,
			expectedTx:  transaction.FlatTransaction{},
			expectedErr: &ClientError{ErrorString: "actNotFound"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := setupTestRPCClientForAutofill(t, tt.mockResponses)

			// Set NetworkID for test
			cl.NetworkID = tt.networkID

			// Make a copy of the original tx for comparison
			originalTx := make(transaction.FlatTransaction)
			for k, v := range tt.tx {
				originalTx[k] = v
			}

			err := cl.autofillRawTransactions(&tt.tx)

			if tt.expectedErr != nil {
				require.Error(t, err)

				// Check error type and message
				switch expectedErr := tt.expectedErr.(type) {
				case *ClientError:
					if clientErr, ok := err.(*ClientError); ok {
						require.Equal(t, expectedErr.ErrorString, clientErr.ErrorString)
					} else {
						t.Errorf("Expected ClientError, but got %T", err)
					}
				default:
					require.Equal(t, tt.expectedErr.Error(), err.Error())
				}
				return
			}

			require.NoError(t, err)

			// Compare the resulting transaction
			if !reflect.DeepEqual(tt.expectedTx, tt.tx) {
				t.Errorf("Expected tx %+v, but got %+v", tt.expectedTx, tt.tx)

				// Detailed comparison for debugging
				if rawTxs, ok := tt.tx["RawTransactions"].([]map[string]any); ok {
					expectedRawTxs := tt.expectedTx["RawTransactions"].([]map[string]any)
					for i, rawTx := range rawTxs {
						if i < len(expectedRawTxs) {
							t.Logf("RawTransaction[%d] expected: %+v", i, expectedRawTxs[i]["RawTransaction"])
							t.Logf("RawTransaction[%d] actual:   %+v", i, rawTx["RawTransaction"])
						}
					}
				}
			}
		})
	}
}

// Helper function to setup test RPC client for autofill tests
func setupTestRPCClientForAutofill(t *testing.T, mockResponses []string) *Client {
	mc := &testutil.JSONRPCMockClient{}
	responseIndex := 0

	mc.DoFunc = func(req *http.Request) (*http.Response, error) {
		if responseIndex < len(mockResponses) {
			response := mockResponses[responseIndex]
			responseIndex++
			return testutil.MockResponse(response, 200, mc)(req)
		}
		// Default empty response if no more responses configured
		return testutil.MockResponse(`{"result": {}}`, 200, mc)(req)
	}

	cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc))
	require.NoError(t, err)

	return NewClient(cfg)
}
