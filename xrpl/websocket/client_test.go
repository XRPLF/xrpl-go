package websocket

import (
	"errors"
	"reflect"
	"testing"

	commonconstants "github.com/Peersyst/xrpl-go/xrpl/common"
	account "github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/interfaces"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

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
					"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"channels": []any{
						map[string]any{
							"account":             "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
							"amount":              "1000",
							"balance":             "0",
							"channel_id":          "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
							"destination_account": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
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
						"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
						"channels": []any{
							map[string]any{
								"account":             "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
								"amount":              "1000",
								"balance":             "0",
								"channel_id":          "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
								"destination_account": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
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
					"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"channels": []any{
						map[string]any{
							"account":             "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
							"amount":              "1000",
							"balance":             "0",
							"channel_id":          "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
							"destination_account": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
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
						"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
						"channels": []any{
							map[string]any{
								"account":             "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
								"amount":              "1000",
								"balance":             "0",
								"channel_id":          "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
								"destination_account": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
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
					"account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
					"channels": []any{
						map[string]any{
							"account":             "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
							"amount":              "1000",
							"balance":             "0",
							"channel_id":          "C7F634794B79DB40E87179A9D1BF05D05797AE7E92DF8E93FD6656E8C4BE3AE7",
							"destination_account": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX",
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
				require.EqualValues(t, tc.res, res)
			}
		})
	}
}

func TestClient_formatRequest(t *testing.T) {
	ws := &Client{}
	tt := []struct {
		description string
		req         interfaces.Request
		id          int
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

func TestClient_convertTransactionAddressToClassicAddress(t *testing.T) {
	ws := &Client{}
	tests := []struct {
		name      string
		tx        transaction.FlatTransaction
		fieldName string
		expected  transaction.FlatTransaction
	}{
		{
			name: "No conversion for classic address",
			tx: transaction.FlatTransaction{
				"Destination": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
			},
			fieldName: "Destination",
			expected: transaction.FlatTransaction{
				"Destination": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
			},
		},
		{
			name: "Field not present in transaction",
			tx: transaction.FlatTransaction{
				"Amount": "1000000",
			},
			fieldName: "Destination",
			expected: transaction.FlatTransaction{
				"Amount": "1000000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws.convertTransactionAddressToClassicAddress(&tt.tx, tt.fieldName)
			if reflect.DeepEqual(tt.expected, &tt.tx) {
				t.Errorf("expected %+v, result %+v", tt.expected, &tt.tx)
			}
		})
	}
}

func TestClient_validateTransactionAddress(t *testing.T) {
	ws := &Client{}
	tests := []struct {
		name         string
		tx           transaction.FlatTransaction
		addressField string
		tagField     string
		expected     transaction.FlatTransaction
		expectedErr  error
	}{
		{
			name: "Valid classic address without tag",
			tx: transaction.FlatTransaction{
				"Account": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
			},
			addressField: "Account",
			tagField:     "SourceTag",
			expected: transaction.FlatTransaction{
				"Account": "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
			},
			expectedErr: nil,
		},
		{
			name: "Valid classic address with tag",
			tx: transaction.FlatTransaction{
				"Destination":    "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"DestinationTag": uint32(12345),
			},
			addressField: "Destination",
			tagField:     "DestinationTag",
			expected: transaction.FlatTransaction{
				"Destination":    "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"DestinationTag": uint32(12345),
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ws.validateTransactionAddress(&tt.tx, tt.addressField, tt.tagField)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tt.expected, tt.tx) {
				t.Errorf("Expected %v, but got %v", tt.expected, tt.tx)
			}
		})
	}
}

func TestClient_setValidTransactionAddresses(t *testing.T) {
	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		expected    transaction.FlatTransaction
		expectedErr error
	}{
		{
			name: "Valid transaction with classic addresses",
			tx: transaction.FlatTransaction{
				"Account":     "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"Destination": "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
			},
			expected: transaction.FlatTransaction{
				"Account":     "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"Destination": "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
			},
			expectedErr: nil,
		},
		{
			name: "Transaction with additional address fields",
			tx: transaction.FlatTransaction{
				"Account":     "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"Destination": "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
				"Owner":       "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"RegularKey":  "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
			},
			expected: transaction.FlatTransaction{
				"Account":     "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"Destination": "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
				"Owner":       "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
				"RegularKey":  "rGWrZyQqhTp9Xu7G5Pkayo7bXjH4k4QYpf",
			},
			expectedErr: nil,
		},
	}

	ws := &Client{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ws.setValidTransactionAddresses(&tt.tx)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tt.expected, tt.tx) {
				t.Errorf("Expected %v, but got %v", tt.expected, tt.tx)
			}
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
		feeCushion     float32
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
			expectedFee: "340", // 10 * (33 + 1) = 340
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
		}, {
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
				for _, m := range tt.serverMessages {
					err := c.WriteJSON(m)
					if err != nil {
						t.Errorf("error writing message: %v", err)
					}
				}
			})
			defer s.Close()

			url, _ := testutil.ConvertHTTPToWS(s.URL)
			cl := NewClient(NewClientConfig().WithHost(url))

			if err := cl.Connect(); err != nil {
				t.Errorf("Error connecting to server: %v", err)
			}

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

func TestClient_setTransactionFlags(t *testing.T) {
	tests := []struct {
		name     string
		tx       transaction.FlatTransaction
		expected uint32
		wantErr  bool
	}{
		{
			name: "No flags set",
			tx: transaction.FlatTransaction{
				"TransactionType": string(transaction.PaymentTx),
			},
			expected: uint32(0),
			wantErr:  false,
		},
		{
			name: "Flags already set",
			tx: transaction.FlatTransaction{
				"TransactionType": string(transaction.PaymentTx),
				"Flags":           uint32(1),
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name: "Missing TransactionType",
			tx: transaction.FlatTransaction{
				"Flags": uint32(1),
			},
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			err := c.setTransactionFlags(&tt.tx)

			if (err != nil) != tt.wantErr {

				t.Errorf("setTransactionFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				flags, ok := tt.tx["Flags"]
				if !ok && tt.expected != 0 {
					t.Errorf("setTransactionFlags() got = %v (type %T), want %v (type %T)", flags, flags, tt.expected, tt.expected)
				}
			}
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
						"info": map[string]any{
							"build_version": "1.12.0",
						},
					},
				},
				{
					"id": 2,
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
			name: "fail - RawTransactions field not an array",
			tx: transaction.FlatTransaction{
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TransactionType": "Batch",
				"RawTransactions": "not_an_array",
			},
			serverMessages: []map[string]any{},
			networkID:      0,
			expectedTx:     transaction.FlatTransaction{},
			expectedErr:    ErrRawTransactionsFieldIsNotAnArray,
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
			expectedTx:     transaction.FlatTransaction{},
			expectedErr:    ErrRawTransactionFieldIsNotAnObject,
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
			expectedTx:     transaction.FlatTransaction{},
			expectedErr:    types.ErrBatchInnerTransactionInvalid,
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
			expectedTx:     transaction.FlatTransaction{},
			expectedErr:    ErrSigningPubKeyFieldMustBeEmpty,
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
			expectedTx:     transaction.FlatTransaction{},
			expectedErr:    ErrTxnSignatureFieldMustBeEmpty,
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
			expectedTx:     transaction.FlatTransaction{},
			expectedErr:    ErrSignersFieldMustBeEmpty,
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
			expectedTx:     transaction.FlatTransaction{},
			expectedErr:    ErrAccountFieldIsNotAString,
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
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "actNotFound",
				},
			},
			networkID:   0,
			expectedTx:  transaction.FlatTransaction{},
			expectedErr: &ErrorWebsocketClientXrplResponse{Type: "actNotFound"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClientForAutofill(t, tt.serverMessages)
			defer cleanup()

			// Set NetworkID for test
			cl.NetworkID = tt.networkID

			// Make a copy of the original tx for comparison
			originalTx := make(transaction.FlatTransaction)
			for k, v := range tt.tx {
				originalTx[k] = v
			}

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
		for _, m := range serverMessages {
			err := c.WriteJSON(m)
			if err != nil {
				t.Errorf("error writing message: %v", err)
			}
		}
	})

	url, _ := testutil.ConvertHTTPToWS(s.URL)
	cl := NewClient(NewClientConfig().WithHost(url))

	if err := cl.Connect(); err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}

	return cl, func() {
		cl.Disconnect()
		s.Close()
	}
}
