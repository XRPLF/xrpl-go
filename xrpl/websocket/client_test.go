package websocket

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	commonconstants "github.com/Peersyst/xrpl-go/xrpl/common"
	clientconfigtestutil "github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/interfaces"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestNewClientInsecureSchemeWarnings(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		wantWarning string
	}{
		{
			name:        "remote insecure scheme warns",
			host:        "ws://s1.ripple.com:6006",
			wantWarning: `xrpl-go: warning: websocket client endpoint "ws://s1.ripple.com:6006" is not using a TLS scheme`,
		},
		{
			name: "local insecure scheme does not warn",
			host: "ws://localhost:6006",
		},
		{
			name: "remote tls scheme does not warn",
			host: "wss://s1.ripple.com:6006",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs := clientconfigtestutil.CaptureLogOutput(t, func() {
				_ = NewClient(NewClientConfig().WithHost(tt.host))
			})

			if tt.wantWarning == "" {
				require.Empty(t, logs)
				return
			}

			require.Contains(t, logs, tt.wantWarning)
		})
	}
}

func TestNewClientWarnsOnceForChainedHosts(t *testing.T) {
	// Confirm fluent re-assignment of host doesn't multiply warnings:
	// only the final host should produce one warning at NewClient.
	logs := clientconfigtestutil.CaptureLogOutput(t, func() {
		_ = NewClient(NewClientConfig().WithHost("ws://a.example:6006").WithHost("ws://b.example:6006"))
	})
	require.Contains(t, logs, `endpoint "ws://b.example:6006"`)
	require.NotContains(t, logs, "a.example")
}

func TestClient_SendRequest(t *testing.T) {
	tt := []struct {
		description    string
		req            interfaces.Request
		res            *ClientResponse
		expectedErr    error
		serverMessages []map[string]any
	}{
		{
			description: "successful request",
			req: &account.ChannelsRequest{
				Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			},
			res: &ClientResponse{
				ID: 1,
				Result: map[string]any{
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
							"settle_delay":        float64(60),
						},
					},
					"ledger_hash":  "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
					"ledger_index": float64(71766314),
					"validated":    true,
				},
			},
			expectedErr: nil,
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
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
								"settle_delay":        float64(60),
							},
						},
						"ledger_hash":  "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
						"ledger_index": common.LedgerIndex(71766314),
						"validated":    true,
					},
				},
			},
		},
		{
			description: "invalid id - timeout",
			req: &account.ChannelsRequest{
				Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			},
			res: &ClientResponse{
				Result: map[string]any{
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
							"settle_delay":        float64(60),
						},
					},
					"ledger_hash":  "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
					"ledger_index": common.LedgerIndex(71766314),
					"validated":    true,
				},
			},
			expectedErr: ErrRequestTimedOut,
			serverMessages: []map[string]any{
				{
					"id": 2,
					"result": map[string]any{
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
								"settle_delay":        float64(60),
							},
						},
						"ledger_hash":  "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
						"ledger_index": common.LedgerIndex(71766314),
						"validated":    true,
					},
				},
			},
		},
		{
			description: "error response",
			req: &account.ChannelsRequest{
				Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
			},
			res: &ClientResponse{
				ID: 1,
				Result: map[string]any{
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
							"settle_delay":        float64(60),
						},
					},
					"ledger_hash":  "1EDBBA3C793863366DF5B31C2174B6B5E6DF6DB89A7212B86838489148E2A581",
					"ledger_index": common.LedgerIndex(71766314),
					"validated":    true,
				},
			},
			expectedErr: &ErrorWebsocketClientXrplResponse{
				Type: "invalidParams",
				Request: map[string]any{
					"account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				},
			},
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "invalidParams",
					"value": map[string]any{
						"account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tc.serverMessages)
			defer cleanup()

			res, err := cl.Request(tc.req)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.res, res)
			}
		})
	}
}

func TestClient_RequestDropsLateTimedOutResponse(t *testing.T) {
	serverErr := make(chan error, 1)
	allowLateResponse := make(chan struct{})
	lateResponseWritten := make(chan struct{})

	cl, cleanup := setupRequestDispatchTestClient(t, func(c *websocket.Conn) {
		defer c.Close()

		firstID, err := readWebsocketRequestID(c)
		if err != nil {
			serverErr <- err
			return
		}

		<-allowLateResponse
		if err := c.WriteJSON(map[string]any{
			"id":     firstID,
			"result": map[string]any{"request": "late"},
		}); err != nil {
			serverErr <- err
			return
		}
		close(lateResponseWritten)

		secondID, err := readWebsocketRequestID(c)
		if err != nil {
			serverErr <- err
			return
		}

		if err := c.WriteJSON(map[string]any{
			"id":     secondID,
			"result": map[string]any{"request": "current"},
		}); err != nil {
			serverErr <- err
		}
	})
	defer cleanup()

	_, err := cl.Request(newAccountChannelsRequest())
	require.ErrorIs(t, err, ErrRequestTimedOut)

	close(allowLateResponse)
	select {
	case <-lateResponseWritten:
	case err := <-serverErr:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for late response")
	}

	res, err := cl.Request(newAccountChannelsRequest())
	require.NoError(t, err)
	require.Equal(t, uint64(2), res.ID)
	require.Equal(t, "current", res.Result["request"])

	select {
	case err := <-serverErr:
		require.NoError(t, err)
	default:
	}
}

func TestClient_RequestMatchesOutOfOrderResponses(t *testing.T) {
	serverErr := make(chan error, 1)
	firstRequestRead := make(chan struct{})

	cl, cleanup := setupRequestDispatchTestClient(t, func(c *websocket.Conn) {
		defer c.Close()

		firstID, err := readWebsocketRequestID(c)
		if err != nil {
			serverErr <- err
			return
		}
		close(firstRequestRead)

		secondID, err := readWebsocketRequestID(c)
		if err != nil {
			serverErr <- err
			return
		}

		if err := c.WriteJSON(map[string]any{
			"id":     secondID,
			"result": map[string]any{"request": "second"},
		}); err != nil {
			serverErr <- err
			return
		}
		if err := c.WriteJSON(map[string]any{
			"id":     firstID,
			"result": map[string]any{"request": "first"},
		}); err != nil {
			serverErr <- err
		}
	})
	defer cleanup()

	firstResult := make(chan requestResult, 1)
	go func() {
		res, err := cl.Request(newAccountChannelsRequest())
		firstResult <- requestResult{res: res, err: err}
	}()

	select {
	case <-firstRequestRead:
	case err := <-serverErr:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first request")
	}

	secondResult := make(chan requestResult, 1)
	go func() {
		res, err := cl.Request(newAccountChannelsRequest())
		secondResult <- requestResult{res: res, err: err}
	}()

	first := receiveRequestResult(t, firstResult)
	second := receiveRequestResult(t, secondResult)

	require.Equal(t, uint64(1), first.ID)
	require.Equal(t, "first", first.Result["request"])
	require.Equal(t, uint64(2), second.ID)
	require.Equal(t, "second", second.Result["request"])

	select {
	case err := <-serverErr:
		require.NoError(t, err)
	default:
	}
}

func TestClient_formatRequest(t *testing.T) {
	ws := &Client{}
	tt := []struct {
		description string
		req         interfaces.Request
		id          uint64
		marker      any
		expected    string
		expectedErr error
	}{
		{
			description: "valid request",
			req: &account.ChannelsRequest{
				Account:            "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				Limit:              70,
			},
			id:     1,
			marker: nil,
			expected: `{
				"id": 1,
				"BaseRequest": {},
				"account":"r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"api_version":2,
				"command":"account_channels",
				"destination_account":"r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"limit":70
			}`,
			expectedErr: nil,
		},
		{
			description: "valid request with marker",
			req: &account.ChannelsRequest{
				Account:            "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				DestinationAccount: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				Limit:              70,
			},
			id:     1,
			marker: "hdsohdaoidhadasd",
			expected: `{
				"id": 1,
				"BaseRequest": {},
				"account":"r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"api_version": 2,
				"command":"account_channels",
				"destination_account":"r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"limit":70,
				"marker":"hdsohdaoidhadasd"
			}`,
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			a, err := ws.formatRequest(tc.req, tc.id, tc.marker)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.JSONEq(t, tc.expected, string(a))
			}
		})
	}
}

// Exhaustive X-address and tag policy cases live in xrpl/internal/client.
// These cases only verify the client wiring: delegation through
// FlatTransaction and the re-exported error identities.
func TestClient_setValidTransactionAddresses(t *testing.T) {
	const (
		classic         = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
		testnetTag14    = "T719a5UwUCnEs54UsxG9CJYYDhwmFCvqXVCALUGJGSbNV3x"
		invalidXAddress = "XVLhHMPHU98es4dbozjVtdWzVrDjtV5fdx1mHp98tDMoQXa"
	)
	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		expected    transaction.FlatTransaction
		expectedErr error
	}{
		{
			name: "classic addresses are unchanged",
			tx: transaction.FlatTransaction{
				"Account":     classic,
				"Destination": "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
			},
			expected: transaction.FlatTransaction{
				"Account":     classic,
				"Destination": "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
			},
		},
		{
			name: "converts X-address and applies embedded tag",
			tx:   transaction.FlatTransaction{"Destination": testnetTag14},
			expected: transaction.FlatTransaction{
				"Destination":    classic,
				"DestinationTag": uint32(14),
			},
		},
		{
			name: "explicit tag conflict",
			tx: transaction.FlatTransaction{
				"Destination":    testnetTag14,
				"DestinationTag": uint32(13),
			},
			expected: transaction.FlatTransaction{
				"Destination":    testnetTag14,
				"DestinationTag": uint32(13),
			},
			expectedErr: ErrMismatchedTag,
		},
		{
			name:        "invalid X-address",
			tx:          transaction.FlatTransaction{"Account": invalidXAddress},
			expected:    transaction.FlatTransaction{"Account": invalidXAddress},
			expectedErr: ErrInvalidAddress,
		},
	}

	ws := &Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ws.setValidTransactionAddresses(&tt.tx)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, tt.tx)
		})
	}
}

func TestClient_setTransactionNextValidSequenceNumber(t *testing.T) {
	tests := []struct {
		name           string
		tx             transaction.FlatTransaction
		serverMessages []map[string]any
		expected       transaction.FlatTransaction
		expectedErr    error
	}{
		{
			name: "Valid transaction",
			tx: transaction.FlatTransaction{
				"Account": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"account_data": map[string]any{
							"Sequence": uint32(42),
						},
						"ledger_current_index": uint32(100),
					},
				},
			},
			expected: transaction.FlatTransaction{
				"Account":  "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"Sequence": uint32(42),
			},
			expectedErr: nil,
		},
		{
			name:           "Missing Account",
			tx:             transaction.FlatTransaction{},
			serverMessages: []map[string]any{},
			expected:       transaction.FlatTransaction{},
			expectedErr:    errors.New("missing Account in transaction"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			err := cl.setTransactionNextValidSequenceNumber(&tt.tx)

			if tt.expectedErr != nil {
				if !reflect.DeepEqual(err.Error(), tt.expectedErr.Error()) {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, tt.tx) {
				t.Logf("Expected:")
				for k, v := range tt.expected {
					t.Logf("  %s: %v (type: %T)", k, v, v)
				}
				t.Logf("Got:")
				for k, v := range tt.tx {
					t.Logf("  %s: %v (type: %T)", k, v, v)
				}
				t.Errorf("Expected %v but got %v", tt.expected, tt.tx)
			}
		})
	}
}

func TestClient_calculateFeePerTransactionType(t *testing.T) {
	tests := []struct {
		name           string
		tx             transaction.FlatTransaction
		serverMessages []map[string]any
		expectedFee    string
		expectedErr    error
		feeCushion     float64
		nSigners       uint64
	}{
		{
			name: "Basic fee calculation",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.PaymentTx,
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
			},
			expectedFee: "10",
			expectedErr: nil,
			feeCushion:  1,
		},
		{
			name: "Half drop rounds upward",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.PaymentTx,
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": 0.000001,
							},
							"load_factor": 10,
						},
					},
				},
			},
			expectedFee: "11",
			expectedErr: nil,
			feeCushion:  1.05,
		},
		{
			name: "Fee calculation with high load factor",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.PaymentTx,
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1000),
						},
					},
				},
			},
			expectedFee: "10000",
			expectedErr: nil,
			feeCushion:  1,
		},
		{
			name: "Fee calculation with max fee limit",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.PaymentTx,
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(1),
							},
							"load_factor": float32(1000),
						},
					},
				},
			},
			expectedFee: "2000000",
			expectedErr: nil,
			feeCushion:  1,
		},
		{
			name: "EscrowFinish with Fulfillment",
			tx: transaction.FlatTransaction{
				"TransactionType": "EscrowFinish",
				"Fulfillment":     "A0028000", // 8 characters = 4 bytes
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
			},
			expectedFee: "333", // ceil(10 * (33 + 4/16))
			expectedErr: nil,
			feeCushion:  1,
		},
		{
			name: "EscrowFinish with empty Fulfillment",
			tx: transaction.FlatTransaction{
				"TransactionType": "EscrowFinish",
				"Fulfillment":     "",
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
			},
			expectedFee: "330",
			expectedErr: nil,
			feeCushion:  1,
		},
		{
			name: "EscrowFinish without Fulfillment",
			tx: transaction.FlatTransaction{
				"TransactionType": "EscrowFinish",
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
			},
			expectedFee: "10", // Regular base fee
			expectedErr: nil,
			feeCushion:  1,
		},
		{
			name: "AccountDelete special transaction cost",
			tx: transaction.FlatTransaction{
				"TransactionType": "AccountDelete",
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
				{
					"id": 2,
					"result": map[string]any{
						"state": map[string]any{
							"validated_ledger": map[string]any{
								"reserve_inc": 2000000, // 2 XRP in drops
							},
						},
					},
				},
			},
			expectedFee: "2000000", // Owner reserve fee
			expectedErr: nil,
			feeCushion:  1,
		},
		{
			name: "AMMCreate special transaction cost",
			tx: transaction.FlatTransaction{
				"TransactionType": "AMMCreate",
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
				{
					"id": 2,
					"result": map[string]any{
						"state": map[string]any{
							"validated_ledger": map[string]any{
								"reserve_inc": 2000000, // 2 XRP in drops
							},
						},
					},
				},
			},
			expectedFee: "2000000", // Owner reserve fee
			expectedErr: nil,
			feeCushion:  1,
		},
		{
			name: "Batch transaction",
			tx: transaction.FlatTransaction{
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"Flags":           uint32(0x40000000),
							"Fee":             "0",
							"SigningPubKey":   "",
						},
					},
					{
						"RawTransaction": map[string]any{
							"TransactionType": "OfferCreate",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"TakerGets":       "1000000",
							"TakerPays": map[string]any{
								"currency": "USD",
								"issuer":   "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
								"value":    "100",
							},
							"Flags":         uint32(0x40000000),
							"Fee":           "0",
							"SigningPubKey": "",
						},
					},
				},
			},
			serverMessages: []map[string]any{
				// Outer Batch fee fetch
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
				// Inner Payment fee fetch
				{
					"id": 2,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
				// Inner OfferCreate fee fetch
				{
					"id": 3,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
			},
			expectedFee: "40", // 2*10 + 10 + 10
			expectedErr: nil,
			feeCushion:  1,
		},
		{
			name: "Batch transaction with multisign",
			tx: transaction.FlatTransaction{
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"Flags":           uint32(0x40000000),
							"Fee":             "0",
							"SigningPubKey":   "",
						},
					},
					{
						"RawTransaction": map[string]any{
							"TransactionType": "OfferCreate",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"TakerGets":       "1000000",
							"TakerPays": map[string]any{
								"currency": "USD",
								"issuer":   "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
								"value":    "100",
							},
							"Flags":         uint32(0x40000000),
							"Fee":           "0",
							"SigningPubKey": "",
						},
					},
				},
			},
			serverMessages: []map[string]any{
				// Outer Batch fee fetch
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
				// Inner Payment fee fetch
				{
					"id": 2,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
				// Inner OfferCreate fee fetch
				{
					"id": 3,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
			},
			expectedFee: "50", // 2*10 + (10+10) + 10 (one extra signer)
			expectedErr: nil,
			feeCushion:  1,
			nSigners:    1,
		},
		{
			name: "Multi-signed transaction",
			tx: transaction.FlatTransaction{
				"TransactionType": transaction.PaymentTx,
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"validated_ledger": map[string]any{
								"base_fee_xrp": float32(0.00001),
							},
							"load_factor": float32(1),
						},
					},
				},
			},
			expectedFee: "30", // 10 + (10 * 2) = 30
			expectedErr: nil,
			feeCushion:  1,
			nSigners:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			cl.cfg.feeCushion = tt.feeCushion
			cl.cfg.maxFeeXRP = DefaultMaxFeeXRP

			err := cl.calculateFeePerTransactionType(&tt.tx, tt.nSigners)

			if tt.expectedErr != nil {
				if !reflect.DeepEqual(err.Error(), tt.expectedErr.Error()) {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(tt.expectedFee, tt.tx["Fee"]) {
					t.Errorf("Expected fee %v, but got %v", tt.expectedFee, tt.tx["Fee"])
				}
			}
		})
	}
}

func TestClient_FeePresenceSemantics(t *testing.T) {
	tests := []struct {
		name        string
		ledger      map[string]any
		expected    string
		expectedErr error
	}{
		{name: "missing base fee", ledger: map[string]any{}, expectedErr: ErrCouldNotGetBaseFeeXrp},
		{name: "null base fee", ledger: map[string]any{"base_fee_xrp": nil}, expectedErr: ErrCouldNotGetBaseFeeXrp},
		{name: "explicit zero base fee", ledger: map[string]any{"base_fee_xrp": 0}, expected: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestClient(t, []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"info": map[string]any{"validated_ledger": tt.ledger, "load_factor": 1},
				},
			}})
			defer cleanup()
			actual, err := client.getFeeXrp(1)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestClient_OwnerReservePresenceSemantics(t *testing.T) {
	tests := []struct {
		name        string
		ledger      map[string]any
		expected    uint64
		expectedErr error
	}{
		{name: "missing reserve", ledger: map[string]any{}, expectedErr: ErrCouldNotFetchOwnerReserve},
		{name: "null reserve", ledger: map[string]any{"reserve_inc": nil}, expectedErr: ErrCouldNotFetchOwnerReserve},
		{name: "explicit zero reserve", ledger: map[string]any{"reserve_inc": 0}, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestClient(t, []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"state": map[string]any{"validated_ledger": tt.ledger},
				},
			}})
			defer cleanup()
			actual, err := client.fetchOwnerReserveFee()
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestClient_LoanSetFeeUsesValidatedLedger(t *testing.T) {
	const owner = "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"
	requests := make(chan map[string]any, 2)
	ws := &testutil.MockWebSocketServer{}
	server := ws.TestWebSocketServer(func(conn *websocket.Conn) {
		defer conn.Close()
		for range 2 {
			var request map[string]any
			if err := conn.ReadJSON(&request); err != nil {
				return
			}
			requests <- request
			var result map[string]any
			switch request["command"] {
			case "ledger_entry":
				result = map[string]any{"node": map[string]any{"Owner": owner}}
			case "account_info":
				result = map[string]any{"signer_lists": []any{}}
			default:
				return
			}
			if err := conn.WriteJSON(map[string]any{
				"id": request["id"], "status": "success", "type": "response", "result": result,
			}); err != nil {
				return
			}
		}
	})
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	client := NewClient(NewClientConfig().WithHost(url).WithNetworkIdentity(0, "1.12.0"))
	require.NoError(t, client.Connect())
	defer client.Disconnect()

	count, err := client.fetchCounterPartySignersCount(transaction.FlatTransaction{"LoanBrokerID": "ABC"})
	require.NoError(t, err)
	require.Equal(t, uint64(1), count)
	close(requests)
	for request := range requests {
		require.Equal(t, "validated", request["ledger_index"])
	}
}

func TestClient_InvalidMaximumFeeUsesPublicError(t *testing.T) {
	serverMessages := []map[string]any{{
		"id": 1,
		"result": map[string]any{
			"info": map[string]any{
				"validated_ledger": map[string]any{"base_fee_xrp": 0.00001},
				"load_factor":      1,
			},
		},
	}}
	client, cleanup := setupTestClient(t, serverMessages)
	defer cleanup()
	client.cfg.maxFeeXRP = "invalid"
	tx := transaction.FlatTransaction{"TransactionType": "Payment"}
	require.ErrorIs(t, client.calculateFeePerTransactionType(&tx, 0), ErrInvalidFeeValue)
}

func TestClient_setLastLedgerSequence(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		tx             transaction.FlatTransaction
		expectedTx     transaction.FlatTransaction
		expectedErr    error
	}{
		{
			name: "Successfully set LastLedgerSequence",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": transaction.FlatTransaction{
						"ledger_index": 1000,
					},
				},
			},
			tx:          transaction.FlatTransaction{},
			expectedTx:  transaction.FlatTransaction{"LastLedgerSequence": uint32(1000 + commonconstants.LedgerOffset)},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			err := cl.setLastLedgerSequence(&tt.tx)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !reflect.DeepEqual(tt.expectedTx, tt.tx) {
					t.Errorf("Expected tx %v, but got %v", tt.expectedTx, tt.tx)
				}
			}
		})
	}
}

func TestClient_checkAccountDeleteBlockers(t *testing.T) {
	tests := []struct {
		name           string
		address        types.Address
		serverMessages []map[string]any
		expectedErr    error
	}{
		{
			name:    "No blockers",
			address: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
						"account_objects": []any{},
						"ledger_hash":     "4BC50C9B0D8515D3EAAE1E74B29A95804346C491EE1A95BF25E4AAB854A6A651",
						"ledger_index":    30,
						"validated":       true,
					},
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &testutil.MockWebSocketServer{Msgs: tt.serverMessages}
			s := ws.TestWebSocketServer(func(c *websocket.Conn) {
				writeMessagesAfterRequests(t, c, tt.serverMessages)
			})
			defer s.Close()

			url, _ := testutil.ConvertHTTPToWS(s.URL)
			cl := NewClient(NewClientConfig().WithHost(url))
			setTrustedTestNetworkIdentity(cl, 0)

			if err := cl.Connect(); err != nil {
				t.Errorf("Error connecting to server: %v", err)
			}
			defer cl.Disconnect()

			err := cl.checkAccountDeleteBlockers(tt.address)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestClient_Autofill(t *testing.T) {
	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		expectedTx  transaction.FlatTransaction
		expectedErr error
	}{
		{
			name: "fail - missing TransactionType",
			tx: transaction.FlatTransaction{
				"Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
			},
			expectedErr: transaction.ErrTransactionTypeMissing,
		},
		{
			name: "fail - invalid Flags type",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "AccountSet",
				"Flags":           "abc",
			},
			expectedErr: transaction.ErrInvalidFlagsValue,
		},
		{
			name: "pass - missing Flags defaults to uint32(0)",
			tx: transaction.FlatTransaction{
				"Account":            "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType":    "AccountSet",
				"Sequence":           uint32(42),
				"Fee":                "10",
				"LastLedgerSequence": uint32(100),
			},
			expectedTx: transaction.FlatTransaction{
				"Account":            "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType":    "AccountSet",
				"Flags":              uint32(0),
				"Sequence":           uint32(42),
				"Fee":                "10",
				"LastLedgerSequence": uint32(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClientForAutofill(t, nil)
			defer cleanup()

			err := cl.Autofill(&tt.tx)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedTx, tt.tx)
		})
	}
}

func TestClient_autofillRawTransactions(t *testing.T) {
	tests := []struct {
		name           string
		tx             transaction.FlatTransaction
		serverMessages []map[string]any
		networkID      uint32
		expectedTx     transaction.FlatTransaction
		expectedErr    error
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
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"account_data": map[string]any{
							"Sequence": uint32(42),
						},
					},
				},
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
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"account_data": map[string]any{
							"Sequence": uint32(42),
						},
					},
				},
				{
					"id": 2,
					"result": map[string]any{
						"account_data": map[string]any{
							"Sequence": uint32(100),
						},
					},
				},
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
			name: "pass - multiple transactions same account sequence increment",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Destination":     "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Amount":          "1000000",
						},
					},
					{
						"RawTransaction": map[string]any{
							"TransactionType": "OfferCreate",
							"Account":         "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"TakerGets":       "2000000",
							"TakerPays":       "3000000",
						},
					},
				},
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"account_data": map[string]any{
							"Sequence": uint32(100),
						},
					},
				},
			},
			networkID: 0,
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Destination":     "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Amount":          "1000000",
							"Fee":             "0",
							"SigningPubKey":   "",
							"Sequence":        uint32(100), // First use of this account
						},
					},
					{
						"RawTransaction": map[string]any{
							"TransactionType": "OfferCreate",
							"Account":         "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"TakerGets":       "2000000",
							"TakerPays":       "3000000",
							"Fee":             "0",
							"SigningPubKey":   "",
							"Sequence":        uint32(101), // Incremented from cached value
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "pass - transaction with NetworkID needed",
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
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"account_data": map[string]any{
							"Sequence": uint32(42),
						},
					},
				},
			},
			networkID: 2000, // Above RestrictedNetworks threshold
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       uint32(2000),
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"Fee":             "0",
							"SigningPubKey":   "",
							"NetworkID":       uint32(2000),
							"Sequence":        uint32(43),
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
			serverMessages: []map[string]any{},
			networkID:      0,
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
		{
			name: "pass - fee field already set to 0 - valid",
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
							"Fee":             "0",
						},
					},
				},
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"account_data": map[string]any{
							"Sequence": uint32(42),
						},
					},
				},
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
							"Sequence":        uint32(43),
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "pass - signingPubKey field already empty - valid",
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
							"SigningPubKey":   "",
						},
					},
				},
			},
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"account_data": map[string]any{
							"Sequence": uint32(42),
						},
					},
				},
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
							"Sequence":        uint32(43),
						},
					},
				},
			},
			expectedErr: nil,
		},
		// Error cases
		{
			name: "pass - inner NetworkID matches client NetworkID and outer NetworkID",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       uint32(2000),
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"NetworkID":       uint32(2000),
							"TicketSequence":  uint32(7),
						},
					},
				},
			},
			serverMessages: []map[string]any{},
			networkID:      uint32(2000),
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       uint32(2000),
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"NetworkID":       uint32(2000),
							"TicketSequence":  uint32(7),
							"Fee":             "0",
							"SigningPubKey":   "",
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "fail - inner NetworkID does not match client NetworkID",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       uint32(2000),
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"NetworkID":       uint32(2001),
						},
					},
				},
			},
			serverMessages: []map[string]any{},
			networkID:      uint32(2000),
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       uint32(2000),
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
							"Amount":          "1000000",
							"NetworkID":       uint32(2001),
						},
					},
				},
			},
			expectedErr: ErrNetworkIDFieldMismatch,
		},
		{
			name: "fail - inner NetworkID is not a uint32",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"NetworkID":       "2000",
						},
					},
				},
			},
			serverMessages: []map[string]any{},
			networkID:      uint32(2000),
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"NetworkID":       "2000",
						},
					},
				},
			},
			expectedErr: ErrNetworkIDFieldIsNotAUint32,
		},
		{
			name: "fail - outer NetworkID does not match client NetworkID",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       uint32(3000),
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"NetworkID":       uint32(2000),
						},
					},
				},
			},
			serverMessages: []map[string]any{},
			networkID:      uint32(2000),
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       uint32(3000),
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"NetworkID":       uint32(2000),
						},
					},
				},
			},
			expectedErr: ErrNetworkIDFieldMismatch,
		},
		{
			name: "fail - outer NetworkID does not match default client NetworkID",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       uint32(2000),
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"TicketSequence":  uint32(7),
						},
					},
				},
			},
			serverMessages: []map[string]any{},
			networkID:      0,
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       uint32(2000),
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"TicketSequence":  uint32(7),
						},
					},
				},
			},
			expectedErr: ErrNetworkIDFieldMismatch,
		},
		{
			name: "fail - outer NetworkID is not a uint32",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       "2000",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"NetworkID":       uint32(2000),
						},
					},
				},
			},
			serverMessages: []map[string]any{},
			networkID:      uint32(2000),
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"NetworkID":       "2000",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
							"NetworkID":       uint32(2000),
						},
					},
				},
			},
			expectedErr: ErrNetworkIDFieldIsNotAUint32,
		},
		{
			name: "fail - RawTransactions field not an array",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": "not_an_array",
			},
			serverMessages: []map[string]any{},
			networkID:      0,
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": "not_an_array",
			},
			expectedErr: ErrRawTransactionsFieldIsNotAnArray,
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
			serverMessages: []map[string]any{},
			networkID:      0,
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": "not_an_object",
					},
				},
			},
			expectedErr: ErrRawTransactionFieldIsNotAnObject,
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
			serverMessages: []map[string]any{},
			networkID:      0,
			expectedTx: transaction.FlatTransaction{
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
			expectedErr: types.ErrBatchInnerTransactionInvalid,
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
			serverMessages: []map[string]any{},
			networkID:      0,
			expectedTx: transaction.FlatTransaction{
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
			expectedErr: ErrSigningPubKeyFieldMustBeEmpty,
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
			serverMessages: []map[string]any{},
			networkID:      0,
			expectedTx: transaction.FlatTransaction{
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
			expectedErr: ErrTxnSignatureFieldMustBeEmpty,
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
			serverMessages: []map[string]any{},
			networkID:      0,
			expectedTx: transaction.FlatTransaction{
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
			expectedErr: ErrSignersFieldMustBeEmpty,
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
			serverMessages: []map[string]any{},
			networkID:      0,
			expectedTx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": []map[string]any{
					{
						"RawTransaction": map[string]any{
							"TransactionType": "Payment",
							"Account":         12345,
						},
					},
				},
			},
			expectedErr: ErrAccountFieldIsNotAString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClientForAutofill(t, tt.serverMessages)
			defer cleanup()

			// Set a trusted network identity for this direct helper test.
			setTrustedTestNetworkIdentity(cl, tt.networkID)

			err := cl.autofillRawTransactions(&tt.tx)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error %v, but got nil", tt.expectedErr)
					return
				}

				// Check error type and message
				switch expectedErr := tt.expectedErr.(type) {
				case *ErrorWebsocketClientXrplResponse:
					if wsErr, ok := err.(*ErrorWebsocketClientXrplResponse); ok {
						if wsErr.Type != expectedErr.Type {
							t.Errorf("Expected error type %v, but got %v", expectedErr.Type, wsErr.Type)
						}
					} else {
						t.Errorf("Expected ErrorWebsocketClientXrplResponse, but got %T", err)
					}
				default:
					if err.Error() != tt.expectedErr.Error() {
						t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
					}
				}
				require.Equal(t, tt.expectedTx, tt.tx)
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

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

// Helper function to setup test client for autofill tests
func setupTestClientForAutofill(t *testing.T, serverMessages []map[string]any) (*Client, func()) {
	ws := &testutil.MockWebSocketServer{Msgs: serverMessages}
	s := ws.TestWebSocketServer(func(c *websocket.Conn) {
		writeMessagesAfterRequests(t, c, serverMessages)
	})

	url, _ := testutil.ConvertHTTPToWS(s.URL)
	cl := NewClient(NewClientConfig().WithHost(url))
	setTrustedTestNetworkIdentity(cl, 0)

	if err := cl.Connect(); err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}

	return cl, func() {
		cl.Disconnect()
		s.Close()
	}
}

type mockFaucetProvider struct {
	err error
}

func (m *mockFaucetProvider) FundWallet(_ types.Address) error {
	return m.err
}

type requestResult struct {
	res *ClientResponse
	err error
}

func setupRequestDispatchTestClient(t *testing.T, handler func(*websocket.Conn)) (*Client, func()) {
	t.Helper()

	ws := &testutil.MockWebSocketServer{}
	s := ws.TestWebSocketServer(handler)

	url, err := testutil.ConvertHTTPToWS(s.URL)
	require.NoError(t, err)

	cl := NewClient(NewClientConfig().
		WithHost(url).
		WithTimeout(100 * time.Millisecond))
	setTrustedTestNetworkIdentity(cl, 0)

	require.NoError(t, cl.Connect())

	return cl, func() {
		_ = cl.Disconnect()
		s.Close()
	}
}

func readWebsocketRequestID(c *websocket.Conn) (uint64, error) {
	var req struct {
		ID uint64 `json:"id"`
	}

	if err := c.ReadJSON(&req); err != nil {
		return 0, err
	}

	return req.ID, nil
}

func writeMessagesAfterRequests(t *testing.T, c *websocket.Conn, messages []map[string]any) {
	t.Helper()

	for _, m := range messages {
		if _, err := readWebsocketRequestID(c); err != nil {
			t.Errorf("read websocket request id: %v", err)
			return
		}

		if err := c.WriteJSON(m); err != nil {
			t.Errorf("write websocket response: %v", err)
			return
		}
	}

	// Drain until the client closes the connection. Without this, the orphaned
	// server-side conn is finalized by GC, surfacing as close-1006 on the
	// client and triggering an auto-reconnect that spawns a second handler
	// goroutine which can outlive the test and panic via t.Errorf.
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}

func receiveRequestResult(t *testing.T, resultChan <-chan requestResult) *ClientResponse {
	t.Helper()

	select {
	case result := <-resultChan:
		require.NoError(t, result.err)
		return result.res
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for request result")
	}

	return nil
}

func newAccountChannelsRequest() *account.ChannelsRequest {
	return &account.ChannelsRequest{
		Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	}
}

func TestClient_FundWallet(t *testing.T) {
	const testAddr = "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn"
	prevMaxAttempts := fundWalletMaxAttempts
	prevPollInterval := fundWalletPollInterval
	fundWalletMaxAttempts = 3
	fundWalletPollInterval = time.Millisecond
	t.Cleanup(func() {
		fundWalletMaxAttempts = prevMaxAttempts
		fundWalletPollInterval = prevPollInterval
	})

	accountInfoMsg := func(id int, balance string) map[string]any {
		return map[string]any{
			"id": id,
			"result": map[string]any{
				"account_data": map[string]any{
					"Account": testAddr,
					"Balance": balance,
				},
			},
		}
	}

	actNotFoundMsg := func(id int) map[string]any {
		return map[string]any{
			"id":    id,
			"error": actNotFound,
		}
	}
	invalidParamsMsg := func(id int) map[string]any {
		return map[string]any{
			"id":    id,
			"error": "invalidParams",
		}
	}

	tests := []struct {
		name           string
		address        string
		faucetErr      error
		serverMessages []map[string]any
		expectedErr    error
	}{
		{
			name:      "pass - new account funded successfully",
			address:   testAddr,
			faucetErr: nil,
			serverMessages: []map[string]any{
				actNotFoundMsg(1),
				accountInfoMsg(2, "1000000000"),
			},
			expectedErr: nil,
		},
		{
			name:      "pass - existing account balance increases",
			address:   testAddr,
			faucetErr: nil,
			serverMessages: []map[string]any{
				accountInfoMsg(1, "1000"),
				accountInfoMsg(2, "1000"),
				accountInfoMsg(3, "2000"),
			},
			expectedErr: nil,
		},
		{
			name:      "fail - balance never updates",
			address:   testAddr,
			faucetErr: nil,
			serverMessages: []map[string]any{
				accountInfoMsg(1, "1000"),
				accountInfoMsg(2, "1000"),
				accountInfoMsg(3, "1000"),
				accountInfoMsg(4, "1000"),
			},
			expectedErr: ErrFundWalletBalanceNotUpdated,
		},
		{
			name:      "fail - polling balance error returns immediately",
			address:   testAddr,
			faucetErr: nil,
			serverMessages: []map[string]any{
				accountInfoMsg(1, "1000"),
				invalidParamsMsg(2),
			},
			expectedErr: &ErrorWebsocketClientXrplResponse{Type: "invalidParams"},
		},
		{
			name:      "fail - faucet returns error",
			address:   testAddr,
			faucetErr: errors.New("faucet unavailable"),
			serverMessages: []map[string]any{
				actNotFoundMsg(1),
			},
			expectedErr: errors.New("faucet unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &testutil.MockWebSocketServer{Msgs: tt.serverMessages}
			s := ws.TestWebSocketServer(func(c *websocket.Conn) {
				writeMessagesAfterRequests(t, c, tt.serverMessages)
			})
			defer s.Close()

			url, _ := testutil.ConvertHTTPToWS(s.URL)
			cl := NewClient(
				NewClientConfig().
					WithHost(url).
					WithTimeout(1 * time.Second).
					WithFaucetProvider(&mockFaucetProvider{err: tt.faucetErr}),
			)
			setTrustedTestNetworkIdentity(cl, 0)

			require.NoError(t, cl.Connect())
			defer cl.Disconnect()

			w := &wallet.Wallet{ClassicAddress: types.Address(tt.address)}
			err := cl.FundWallet(w)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("fail - missing classic address", func(t *testing.T) {
		cl := NewClient(*NewClientConfig())
		w := &wallet.Wallet{ClassicAddress: ""}
		err := cl.FundWallet(w)
		require.ErrorIs(t, err, ErrCannotFundWalletWithoutClassicAddress)
	})
}

func TestClient_ReconnectsAfterNonCloseReadError(t *testing.T) {
	var dialCount atomic.Int32
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dialCount.Add(1)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var request struct {
			ID uint64 `json:"id"`
		}
		if err := json.Unmarshal(message, &request); err != nil {
			return
		}
		_ = conn.WriteJSON(map[string]any{
			"id":     request.ID,
			"status": "success",
			"type":   "response",
			"result": map[string]any{},
		})
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)

	cfg := withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithMaxReconnects(2).
			WithTimeout(time.Second),
		time.Millisecond,
		time.Millisecond,
	)
	client := NewClient(cfg)
	setTrustedTestNetworkIdentity(client, 0)
	client.conn.conn = newFakeWebsocketConnection()
	ctx := client.resetLifecycle()
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		client.readMessages(ctx)
	}()
	defer func() {
		client.cancelLifecycle()
		if client.IsConnected() {
			_ = client.conn.Disconnect()
		}
		select {
		case <-readDone:
		case <-time.After(time.Second):
			t.Fatal("read loop did not stop")
		}
	}()

	require.Eventually(t, func() bool {
		return dialCount.Load() == 1 && client.IsConnected()
	}, time.Second, time.Millisecond)

	response, err := client.Request(newAccountChannelsRequest())
	require.NoError(t, err)
	require.NotNil(t, response)
}

// TestClient_ReconnectConsumesBudgetOnConnectFailures verifies that a failed
// reconnect Connect() does not abort the read loop early. The loop must keep
// retrying until the full WithMaxReconnects budget is exhausted, and only then
// report ErrMaxReconnectionAttemptsReached. The server accepts the initial
// upgrade once (then closes the conn to trigger the reconnect path) and
// rejects every subsequent dial with HTTP 500, so every retry's Connect()
// fails. Before the fix, the first failed Connect() would surface the dial
// error and return, after the fix, retries continue and the budget exhausts.
func TestClient_ReconnectConsumesBudgetOnConnectFailures(t *testing.T) {
	const budget = 3
	var dialCount atomic.Int32
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if dialCount.Add(1) == 1 {
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			c.Close()
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)

	cfg := withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithTimeout(1*time.Second).
			WithMaxReconnects(budget),
		time.Millisecond,
		time.Millisecond,
	)

	cl := NewClient(cfg)
	setTrustedTestNetworkIdentity(cl, 0)

	errCh := make(chan error, 1)
	cl.OnError(func(err error) {
		select {
		case errCh <- err:
		default:
		}
	})

	require.NoError(t, cl.Connect())
	defer cl.Disconnect()

	select {
	case got := <-errCh:
		var maxErr ErrMaxReconnectionAttemptsReached
		require.ErrorAs(t, got, &maxErr)
		require.Equal(t, budget, maxErr.Attempts)
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for ErrMaxReconnectionAttemptsReached, dial count=%d", dialCount.Load())
	}

	require.GreaterOrEqual(t, dialCount.Load(), int32(1+budget))
}

// TestClient_ReconnectConsumesBudgetWhenReconnectClosesBeforeMessage verifies
// that a reconnect only resets the retry budget after the socket becomes
// usable by delivering a message. The server accepts every upgrade and
// immediately closes the socket, so the client must eventually exhaust
// WithMaxReconnects instead of looping forever.
func TestClient_ReconnectConsumesBudgetWhenReconnectClosesBeforeMessage(t *testing.T) {
	const budget = 2

	var dialCount atomic.Int32
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dialCount.Add(1)
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c.Close()
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)

	cfg := withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithTimeout(1*time.Second).
			WithMaxReconnects(budget),
		time.Millisecond,
		time.Millisecond,
	)

	cl := NewClient(cfg)
	setTrustedTestNetworkIdentity(cl, 0)

	errCh := make(chan error, 1)
	cl.OnError(func(err error) {
		select {
		case errCh <- err:
		default:
		}
	})

	require.NoError(t, cl.Connect())
	defer cl.Disconnect()

	select {
	case got := <-errCh:
		var maxErr ErrMaxReconnectionAttemptsReached
		require.ErrorAs(t, got, &maxErr)
		require.Equal(t, budget, maxErr.Attempts)
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for ErrMaxReconnectionAttemptsReached, dial count=%d", dialCount.Load())
	}
}

func TestReconnectDelayUsesCappedExponentialBackoff(t *testing.T) {
	const (
		baseDelay = time.Millisecond
		maxDelay  = 30 * time.Millisecond
	)

	require.Equal(t, time.Millisecond, reconnectDelay(1, baseDelay, maxDelay))
	require.Equal(t, 2*time.Millisecond, reconnectDelay(2, baseDelay, maxDelay))
	require.Equal(t, 4*time.Millisecond, reconnectDelay(3, baseDelay, maxDelay))
	require.Equal(t, 30*time.Millisecond, reconnectDelay(6, baseDelay, maxDelay))
	require.Equal(t, 30*time.Millisecond, reconnectDelay(100, baseDelay, maxDelay))
}

// withReconnectDelays returns a config with test-specific reconnect delays.
// The config is copied into one client before its read loop starts.
func withReconnectDelays(cfg ClientConfig, baseDelay, maxDelay time.Duration) ClientConfig {
	cfg.reconnectBaseDelay = baseDelay
	cfg.reconnectMaxDelay = maxDelay
	return cfg
}
