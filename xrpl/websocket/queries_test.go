package websocket

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/channel"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerqueries "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	ledgertypes "github.com/Peersyst/xrpl-go/xrpl/queries/ledger/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/nft"
	nfttypes "github.com/Peersyst/xrpl-go/xrpl/queries/nft/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/oracle"
	"github.com/Peersyst/xrpl-go/xrpl/queries/path"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	servertypes "github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/utility"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
)

func setupTestClient(t *testing.T, messages []map[string]any) (*Client, func()) {
	ws := &testutil.MockWebSocketServer{Msgs: messages}
	s := ws.TestWebSocketServer(func(c *websocket.Conn) {
		for _, m := range messages {
			err := c.WriteJSON(m)
			if err != nil {
				t.Errorf("error writing message: %v", err)
			}
		}
	})

	url, _ := testutil.ConvertHTTPToWS(s.URL)
	cl := NewClient(NewClientConfig().
		WithHost(url).
		WithTimeout(1 * time.Second))

	if err := cl.Connect(); err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}

	cleanup := func() {
		fmt.Println("Disconnecting from server")
		cl.Disconnect()
		s.Close()
	}

	return cl, cleanup
}

func TestClient_GetServerInfo(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *server.InfoResponse
		expectedErr    error
	}{
		{
			name: "Valid server info",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"info": map[string]any{
							"build_version":     "1.9.4",
							"complete_ledgers":  "32570-62964740",
							"hostid":            "MIST",
							"load_factor":       float64(1),
							"peers":             float64(96),
							"pubkey_node":       "n9KUjqxCr5FKThSNXdzb7oqN8rYwScB2dUnNqxQxbEA17JkaWy5x",
							"server_state":      "full",
							"validation_quorum": float64(28),
						},
					},
				},
			},
			expected: &server.InfoResponse{
				Info: servertypes.Info{
					BuildVersion:     "1.9.4",
					CompleteLedgers:  "32570-62964740",
					HostID:           "MIST",
					LoadFactor:       1,
					Peers:            96,
					PubkeyNode:       "n9KUjqxCr5FKThSNXdzb7oqN8rYwScB2dUnNqxQxbEA17JkaWy5x",
					ServerState:      "full",
					ValidationQuorum: 28,
				},
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetServerInfo(&server.InfoRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetAccountInfo(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *account.InfoResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{
				{

					"id": 1,
					"result": map[string]any{
						"account_data": map[string]any{
							"Account":           "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
							"Flags":             0,
							"LedgerEntryType":   "AccountRoot",
							"OwnerCount":        0,
							"PreviousTxnID":     "4294BEBE5B569A18C0A2702387C9B1E7146DC3A5850C1E87204951C6FDAA4C42",
							"PreviousTxnLgrSeq": 3,
							"Sequence":          6,
						},
						"validated": false,
					},
				},
			},
			expected: &account.InfoResponse{
				AccountData: ledger.AccountRoot{
					Account:           "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
					Flags:             0,
					LedgerEntryType:   "AccountRoot",
					OwnerCount:        0,
					PreviousTxnID:     "4294BEBE5B569A18C0A2702387C9B1E7146DC3A5850C1E87204951C6FDAA4C42",
					PreviousTxnLgrSeq: 3,
					Sequence:          6,
				},
				Validated: false,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAccountInfo(&account.InfoRequest{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetAccountChannels(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *account.ChannelsResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
					"channels": []map[string]any{
						{
							"account":             "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
							"destination_account": "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
							"amount":              "100",
							"balance":             "0",
							"channel_id":          "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3",
						},
					},
					"ledger_hash":  "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
					"ledger_index": 14380380,
					"validated":    true,
				},
			}},
			expected: &account.ChannelsResponse{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				Channels: []accounttypes.ChannelResult{
					{
						Account:            "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
						DestinationAccount: "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
						Amount:             "100",
						Balance:            "0",
						ChannelID:          "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3",
					},
				},
				LedgerHash:  "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
				LedgerIndex: 14380380,
				Validated:   true,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAccountChannels(&account.ChannelsRequest{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetAccountObjects(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *account.ObjectsResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
					"account_objects": []map[string]any{
						{
							"LedgerEntryType": "RippleState",
							"Balance": map[string]any{
								"currency": "USD",
								"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
								"value":    "100",
							},
						},
					},
					"ledger_hash":  "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
					"ledger_index": 14380380,
					"validated":    true,
				},
			}},
			expected: &account.ObjectsResponse{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				AccountObjects: []ledger.FlatLedgerObject{
					{
						"LedgerEntryType": "RippleState",
						"Balance": map[string]any{
							"currency": "USD",
							"issuer":   "rrrrrrrrrrrrrrrrrrrrBZbvji",
							"value":    "100",
						},
					},
				},
				LedgerHash:  "4C99E5F63C0D0B1C2283B4F5DCE2239F80CE92E8B1A6AED1E110C198FC96E659",
				LedgerIndex: 14380380,
				Validated:   true,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAccountObjects(&account.ObjectsRequest{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetXrpBalance(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       string
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"account_data": map[string]any{
						"Account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
						"Balance": "1000000000",
					},
					"validated": true,
				},
			}},
			expected:    "1000",
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetXrpBalance("rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn")

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.expected != result {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestClient_GetAccountLines(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *account.LinesResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
					"lines": []map[string]any{
						{
							"account":  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
							"balance":  "10",
							"currency": "USD",
						},
					},
					"ledger_current_index": 14380380,
					"validated":            true,
				},
			}},
			expected: &account.LinesResponse{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				Lines: []accounttypes.TrustLine{
					{
						Account:  "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
						Balance:  "10",
						Currency: "USD",
					},
				},
				LedgerCurrentIndex: 14380380,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAccountLines(&account.LinesRequest{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetGatewayBalances(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *account.GatewayBalancesResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"account": "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
					"assets": map[string][]map[string]string{
						"r9F6wk8HkXrgYWoJ7fsv4VrUBVoqDVtzkH": {
							{
								"currency": "BTC",
								"value":    "5444166510000000e-26",
							},
						},
					},
					"balances": map[string][]map[string]string{
						"rKm4uWpg9tfwbVSeATv4KxDe6mpE9yPkgJ": {
							{
								"currency": "EUR",
								"value":    "29826.1965999999",
							},
						},
					},
					"ledger_hash":  "61DDBF304AF6E8101576BF161D447CA8E4F0170DDFBEAFFD993DC9383D443388",
					"ledger_index": 14483212,
					"obligations": map[string]string{
						"EUR": "5599.716599999999",
						"USD": "12345.9",
					},
				},
			}},
			expected: &account.GatewayBalancesResponse{
				Account: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
				Assets: map[string][]account.GatewayBalance{
					"r9F6wk8HkXrgYWoJ7fsv4VrUBVoqDVtzkH": {
						{
							Currency: "BTC",
							Value:    "5444166510000000e-26",
						},
					},
				},
				Balances: map[string][]account.GatewayBalance{
					"rKm4uWpg9tfwbVSeATv4KxDe6mpE9yPkgJ": {
						{
							Currency: "EUR",
							Value:    "29826.1965999999",
						},
					},
				},
				LedgerHash:  "61DDBF304AF6E8101576BF161D447CA8E4F0170DDFBEAFFD993DC9383D443388",
				LedgerIndex: 14483212,
				Obligations: map[string]string{
					"EUR": "5599.716599999999",
					"USD": "12345.9",
				},
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetGatewayBalances(&account.GatewayBalancesRequest{
				Account: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetLedgerIndex(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       common.LedgerIndex
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"ledger_index": 14380380,
					"validated":    true,
				},
			}},
			expected:    common.LedgerIndex(14380380),
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetLedgerIndex()

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetLedgerEntry(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		request        *ledgerqueries.EntryRequest
		expected       *ledgerqueries.EntryResponse
		expectedErr    error
	}{
		{
			name: "Valid ledger entry",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"index": "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
						"ledger_current_index": uint32(61809073),
						"node_binary": "test",
						"validated": true,
						"deleted_ledger_index": uint32(0),
					},
				},
			},
			request: &ledgerqueries.EntryRequest{
				Index: "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
			},
			expected: &ledgerqueries.EntryResponse{
				Index: "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
				LedgerCurrentIndex: 61809073,
				NodeBinary: "test",
				Validated: true,
				DeletedLedgerIndex: 0,
			},
			expectedErr: nil,
		},
		{
			name:           "Error response",
			serverMessages: []map[string]any{{"error": "Entry not found"}},
			request: &ledgerqueries.EntryRequest{
				Index: "13F1A95D7AAB7108D5CE7EEAF504B2894B8C674E6D68499076441C4837282BF8",
			},
			expected:    nil,
			expectedErr: errors.New("incorrect id"),
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
			cl := &Client{
				cfg: ClientConfig{
					host: url,
				},
			}

			if err := cl.Connect(); err != nil {
				t.Errorf("Error connecting to server: %v", err)
			}

			result, err := cl.GetLedgerEntry(tt.request)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}

			cl.Disconnect()
		})
	}
}


func TestClient_GetAccountNFTs(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *account.NFTsResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
					"account_nfts": []map[string]any{
						{
							"Flags":        3,
							"Issuer":       "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
							"NFTokenID":    "00081388DC5AA8D5F45498ED961B89E0C69E5846E2F22B68845F76A79",
							"NFTokenTaxon": 0,
							"URI":          "697066733A2F2F62616679626569676479727A74357366703775646D37687537367568377932366E6634646675796C71616266336F636C67747179353566627A6469",
							"nft_serial":   15,
						},
					},
					"ledger_index": 14380380,
					"validated":    true,
				},
			}},
			expected: &account.NFTsResponse{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				AccountNFTs: []accounttypes.NFT{
					{
						Flags:        3,
						Issuer:       "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
						NFTokenID:    "00081388DC5AA8D5F45498ED961B89E0C69E5846E2F22B68845F76A79",
						NFTokenTaxon: 0,
						URI:          "697066733A2F2F62616679626569676479727A74357366703775646D37687537367568377932366E6634646675796C71616266336F636C67747179353566627A6469",
						NFTSerial:    15,
					},
				},
				LedgerIndex: 14380380,
				Validated:   true,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAccountNFTs(&account.NFTsRequest{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetAccountCurrencies(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *account.CurrenciesResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"ledger_index":       14380380,
					"receive_currencies": []string{"USD", "EUR"},
					"send_currencies":    []string{"USD", "JPY"},
					"validated":          true,
				},
			}},
			expected: &account.CurrenciesResponse{
				LedgerIndex:       14380380,
				ReceiveCurrencies: []string{"USD", "EUR"},
				SendCurrencies:    []string{"USD", "JPY"},
				Validated:         true,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAccountCurrencies(&account.CurrenciesRequest{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetAccountOffers(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *account.OffersResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
					"offers": []map[string]any{
						{
							"flags":      float64(0),
							"seq":        float64(1337),
							"taker_gets": "100000000",
							"taker_pays": map[string]any{
								"currency": "USD",
								"issuer":   "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
								"value":    "100",
							},
						},
					},
					"ledger_current_index": float64(14380380),
				},
			}},
			expected: &account.OffersResponse{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				Offers: []accounttypes.OfferResult{
					{
						Flags:     0,
						Sequence:  1337,
						TakerGets: "100000000",
						TakerPays: map[string]any{
							"currency": "USD",
							"issuer":   "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
							"value":    "100",
						},
					},
				},
				LedgerCurrentIndex: 14380380,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAccountOffers(&account.OffersRequest{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetAccountTransactions(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *account.TransactionsResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
					"transactions": []map[string]any{
						{
							"tx_json": map[string]any{
								"Account":         "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
								"Fee":             "10",
								"SigningPubKey":   "0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020",
								"TransactionType": "Payment",
								"TxnSignature":    "304402...",
							},
							"validated": true,
						},
					},
					"ledger_index_min": uint32(14380380),
					"ledger_index_max": uint32(14380381),
				},
			}},
			expected: &account.TransactionsResponse{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
				Transactions: []account.Transaction{
					{
						Tx: transaction.FlatTransaction{
							"Account":         "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
							"Fee":             "10",
							"SigningPubKey":   "0330E7FC9D56BB25D6893BA3F317AE5BCF33B3291BD63DB32654A313222F7FD020",
							"TransactionType": "Payment",
							"TxnSignature":    "304402...",
						},
						Validated: true,
					},
				},
				LedgerIndexMin: 14380380,
				LedgerIndexMax: 14380381,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAccountTransactions(&account.TransactionsRequest{
				Account: "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetChannelVerify(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *channel.VerifyResponse
		expectedErr    error
	}{
		{
			name: "Valid channel verify",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"signature_verified": true,
					},
				},
			},
			expected: &channel.VerifyResponse{
				SignatureVerified: true,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetChannelVerify(&channel.VerifyRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetClosedLedger(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *ledgerqueries.ClosedResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"ledger_hash":  "ABC123",
					"ledger_index": uint32(14380380),
				},
			}},
			expected: &ledgerqueries.ClosedResponse{
				LedgerHash:  "ABC123",
				LedgerIndex: 14380380,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetClosedLedger()

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetCurrentLedger(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *ledgerqueries.CurrentResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"ledger_current_index": uint32(14380380),
				},
			}},
			expected: &ledgerqueries.CurrentResponse{
				LedgerCurrentIndex: 14380380,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetCurrentLedger()

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetLedgerData(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *ledgerqueries.DataResponse
		expectedErr    error
	}{
		{
			name: "Valid ledger data",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"ledger_hash":  "abc123",
						"ledger_index": "123",
						"state": []map[string]any{
							{
								"data":  "1100612200000000000000000000000000000000",
								"index": "E6DBAFC99223B42257915A63DFC6B0C032D4070F9A574B255AD97466726FC321",
							},
						},
					},
				},
			},
			expected: &ledgerqueries.DataResponse{
				LedgerHash:  "abc123",
				LedgerIndex: "123",
				State: []ledgertypes.State{
					{
						Data:  "1100612200000000000000000000000000000000",
						Index: "E6DBAFC99223B42257915A63DFC6B0C032D4070F9A574B255AD97466726FC321",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetLedgerData(&ledgerqueries.DataRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetLedger(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *ledgerqueries.Response
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"ledger": map[string]any{
						"ledger_index":          uint32(14380380),
						"total_coins":           "99999999999999997",
						"parent_hash":           "ABC123",
						"transaction_hash":      "DEF456",
						"account_hash":          "GHI789",
						"parent_close_time":     uint32(123456),
						"close_time":            uint32(123457),
						"close_time_human":      "2023-Aug-01 12:34:56.789",
						"close_time_resolution": uint32(10),
						"closed":                true,
					},
				},
			}},
			expected: &ledgerqueries.Response{
				Ledger: ledgertypes.BaseLedger{
					LedgerIndex:         14380380,
					TotalCoins:          types.XRPCurrencyAmount(99999999999999997),
					ParentHash:          "ABC123",
					TransactionHash:     "DEF456",
					AccountHash:         "GHI789",
					ParentCloseTime:     123456,
					CloseTime:           123457,
					CloseTimeHuman:      "2023-Aug-01 12:34:56.789",
					CloseTimeResolution: 10,
					Closed:              true,
				},
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetLedger(&ledgerqueries.Request{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetNFTBuyOffers(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *nft.NFTokenBuyOffersResponse
		expectedErr    error
	}{
		{
			name: "successful response",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"nft_id": "123",
						"offers": []map[string]any{
							{
								"owner":           "r456",
								"amount":          "100",
								"flags":           uint32(1),
								"nft_offer_index": "123",
							},
						},
					},
				},
			},
			expected: &nft.NFTokenBuyOffersResponse{
				NFTokenID: "123",
				Offers: []nfttypes.NFTokenOffer{
					{
						Amount:            "100",
						Flags:             1,
						NFTokenOfferIndex: "123",
						Owner:             "r456",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetNFTBuyOffers(&nft.NFTokenBuyOffersRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetNFTSellOffers(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *nft.NFTokenSellOffersResponse
		expectedErr    error
	}{
		{
			name: "successful response",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"nft_id": "123",
						"offers": []map[string]any{
							{
								"owner":           "r456",
								"amount":          "100",
								"flags":           uint32(1),
								"nft_offer_index": "123",
							},
						},
					},
				},
			},
			expected: &nft.NFTokenSellOffersResponse{
				NFTokenID: "123",
				Offers: []nfttypes.NFTokenOffer{
					{
						Amount:            "100",
						Flags:             1,
						NFTokenOfferIndex: "123",
						Owner:             "r456",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetNFTSellOffers(&nft.NFTokenSellOffersRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetBookOffers(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *path.BookOffersResponse
		expectedErr    error
	}{
		{
			name: "successful response",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"ledger_current_index": uint32(123),
						"ledger_index":         uint32(123),
						"ledger_hash":          "123",
						"offers": []map[string]any{
							{
								"owner_funds":       "100",
								"taker_gets_funded": "100",
								"taker_pays_funded": "100",
								"quality":           "100",
							},
						},
						"validated": true,
					},
				},
			},
			expected: &path.BookOffersResponse{
				LedgerCurrentIndex: 123,
				LedgerIndex:        123,
				LedgerHash:         "123",
				Offers: []pathtypes.BookOffer{
					{
						OwnerFunds:      "100",
						TakerGetsFunded: "100",
						TakerPaysFunded: "100",
						Quality:         "100",
					},
				},
				Validated: true,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetBookOffers(&path.BookOffersRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetDepositAuthorized(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *path.DepositAuthorizedResponse
		expectedErr    error
	}{
		{
			name: "Valid deposit authorized response",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"deposit_authorized":   true,
						"destination_account":  "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
						"ledger_current_index": uint32(70825689),
						"source_account":       "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
						"validated":            true,
					},
				},
			},
			expected: &path.DepositAuthorizedResponse{
				DepositAuthorized:  true,
				DestinationAccount: "rEhxGqkqPPSxQ3P25J66ft5TwpzV14k2de",
				LedgerCurrentIndex: 70825689,
				SourceAccount:      "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				Validated:          true,
			},
			expectedErr: nil,
		},
		{
			name: "invalid id - timeout",
			serverMessages: []map[string]any{
				{
					"id": 2,
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetDepositAuthorized(&path.DepositAuthorizedRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_FindPathCreate(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *path.FindResponse
		expectedErr    error
	}{
		{
			name: "Valid path find",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"alternatives": []map[string]any{
							{
								"paths_computed": [][]map[string]any{
									{
										{
											"currency": "USD",
											"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
											"account":  "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
										},
									},
								},
								"source_amount": "100000",
							},
						},
						"destination_account": "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
						"destination_amount":  "100",
						"source_account":      "rLHmBn4fT93D1NuWEGNxnYvhvGxzPVVJ5C",
					},
				},
			},
			expected: &path.FindResponse{
				Alternatives: []pathtypes.Alternative{
					{
						PathsComputed: [][]transaction.PathStep{
							{
								{
									Currency: "USD",
									Issuer:   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
									Account:  "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
								},
							},
						},
						SourceAmount: "100000",
					},
				},
				DestinationAccount: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
				DestinationAmount:  "100",
				SourceAccount:      "rLHmBn4fT93D1NuWEGNxnYvhvGxzPVVJ5C",
			},
			expectedErr: nil,
		},
		{
			name:           "error response",
			serverMessages: []map[string]any{{"id": 1, "error": "incorrect id"}},
			expected:       nil,
			expectedErr:    ErrIncorrectID,
		},
		{
			name:           "invalid id timeout",
			serverMessages: []map[string]any{{"id": 2, "result": map[string]any{}}},
			expected:       nil,
			expectedErr:    ErrRequestTimedOut,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.FindPathCreate(&path.FindCreateRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_FindPathClose(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *path.FindResponse
		expectedErr    error
	}{
		{
			name: "successful path close",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"alternatives":        []map[string]any{},
						"destination_account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
						"destination_amount":  "100",
						"source_account":      "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
						"full_reply":          true,
						"closed":              true,
						"status":              true,
					},
				},
			},
			expected: &path.FindResponse{
				Alternatives:       []pathtypes.Alternative{},
				DestinationAccount: "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				DestinationAmount:  "100",
				SourceAccount:      "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn",
				FullReply:          true,
				Closed:             true,
				Status:             true,
			},
			expectedErr: nil,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    2,
					"error": "incorrect id",
				},
			},
			expected:    nil,
			expectedErr: ErrRequestTimedOut,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expectedErr: ErrIncorrectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.FindPathClose(&path.FindCloseRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_FindPathStatus(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *path.FindResponse
		expectedErr    error
	}{
		{
			name: "successful path status request",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"alternatives": []map[string]any{
							{
								"paths_computed": [][]map[string]any{
									{
										{"currency": "USD", "issuer": "rXXXXXXXXXXXXXXXXXXXXX"},
									},
								},
								"source_amount": "100",
							},
						},
						"destination_account": "rXXXXXXXXXXXXXXXXXXXXX",
						"destination_amount":  "100",
						"source_account":      "rYYYYYYYYYYYYYYYYYYYYY",
						"full_reply":          true,
						"status":              true,
					},
				},
			},
			expected: &path.FindResponse{
				Alternatives: []pathtypes.Alternative{
					{
						PathsComputed: [][]transaction.PathStep{
							{
								{Currency: "USD", Issuer: "rXXXXXXXXXXXXXXXXXXXXX"},
							},
						},
						SourceAmount: "100",
					},
				},
				DestinationAccount: "rXXXXXXXXXXXXXXXXXXXXX",
				DestinationAmount:  "100",
				SourceAccount:      "rYYYYYYYYYYYYYYYYYYYYY",
				FullReply:          true,
				Status:             true,
			},
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":    1,
					"error": "incorrect id",
				},
			},
			expectedErr: errors.New("incorrect id"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.FindPathStatus(&path.FindStatusRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetRipplePathFind(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *path.RipplePathFindResponse
		expectedErr    error
	}{
		{
			name: "successful response",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"alternatives": []map[string]any{
							{
								"paths_computed": [][]map[string]any{
									{
										{
											"account":  "rMAZ5ZnK73nyNUL4foAvaxdreczCkG3vA6",
											"currency": "USD",
											"issuer":   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
										},
									},
								},
								"source_amount": "100",
							},
						},
						"destination_account": "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
						"source_account":      "rMAZ5ZnK73nyNUL4foAvaxdreczCkG3vA6",
					},
				},
			},
			expected: &path.RipplePathFindResponse{
				Alternatives: []pathtypes.RippleAlternative{
					{
						PathsComputed: [][]transaction.PathStep{
							{
								{
									Account:  "rMAZ5ZnK73nyNUL4foAvaxdreczCkG3vA6",
									Currency: "USD",
									Issuer:   "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
								},
							},
						},
						SourceAmount: "100",
					},
				},
				DestinationAccount: "rMwjYedjc7qqtKYVLiAccJSmCwih4LnE2q",
				SourceAccount:      "rMAZ5ZnK73nyNUL4foAvaxdreczCkG3vA6",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetRipplePathFind(&path.RipplePathFindRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetAllFeatures(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *server.FeatureAllResponse
		expectedErr    error
	}{
		{
			name: "successful response",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"features": map[string]any{
							"MultiSign": map[string]any{
								"enabled":   true,
								"name":      "Multi-Signing",
								"supported": true,
								"vetoed":    false,
							},
						},
					},
					"status": "success",
					"type":   "response",
				},
			},
			expected: &server.FeatureAllResponse{
				Features: map[string]servertypes.FeatureStatus{
					"MultiSign": {
						Enabled:   true,
						Name:      "Multi-Signing",
						Supported: true,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"status": "error",
					"error":  "invalidParams",
					"type":   "response",
				},
			},
			expected:    nil,
			expectedErr: errors.New("invalidParams"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAllFeatures(&server.FeatureAllRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetFeature(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *server.FeatureResponse
		expectedErr    error
	}{
		{
			name: "successful response",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"testFeature": map[string]any{
							"enabled":   true,
							"name":      "Test Feature",
							"supported": true,
						},
					},
					"status": "success",
					"type":   "response",
				},
			},
			expected: &server.FeatureResponse{
				"testFeature": {
					Enabled:   true,
					Name:      "Test Feature",
					Supported: true,
				},
			},
			expectedErr: nil,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"error":  "invalidParams",
					"status": "error",
					"type":   "response",
				},
			},
			expected:    nil,
			expectedErr: errors.New("invalidParams"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetFeature(&server.FeatureOneRequest{Feature: "testFeature"})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetFee(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *server.FeeResponse
		expectedErr    error
	}{
		{
			name: "successful response",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"current_ledger_size": "14",
						"current_queue_size":  "20",
						"drops": map[string]any{
							"base_fee":        "10",
							"median_fee":      "10000",
							"minimum_fee":     "10",
							"open_ledger_fee": "10",
						},
						"expected_ledger_size": "30",
						"ledger_current_index": uint32(1),
						"levels": map[string]any{
							"median_level":      "128000",
							"minimum_level":     "256",
							"open_ledger_level": "256",
							"reference_level":   "256",
						},
						"max_queue_size": "20",
						"status":         "success",
					},
				},
			},
			expected: &server.FeeResponse{
				CurrentLedgerSize:  "14",
				CurrentQueueSize:   "20",
				ExpectedLedgerSize: "30",
				LedgerCurrentIndex: 1,
				MaxQueueSize:       "20",
				Drops: servertypes.FeeDrops{
					BaseFee:       types.XRPCurrencyAmount(10),
					MedianFee:     types.XRPCurrencyAmount(10000),
					MinimumFee:    types.XRPCurrencyAmount(10),
					OpenLedgerFee: types.XRPCurrencyAmount(10),
				},
				Levels: servertypes.FeeLevels{
					MedianLevel:     types.XRPCurrencyAmount(128000),
					MinimumLevel:    types.XRPCurrencyAmount(256),
					OpenLedgerLevel: types.XRPCurrencyAmount(256),
					ReferenceLevel:  types.XRPCurrencyAmount(256),
				},
			},
			expectedErr: nil,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"error":  "invalidParams",
					"status": "error",
					"type":   "response",
				},
			},
			expected:    nil,
			expectedErr: errors.New("invalidParams"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetFee(&server.FeeRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetManifest(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *server.ManifestResponse
		expectedErr    error
	}{
		{
			name: "successful response",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"status": "success",
					"type":   "response",
					"result": map[string]any{
						"details": map[string]any{
							"domain":        "example.com",
							"ephemeral_key": "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p",
							"master_key":    "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p",
							"seq":           1,
						},
						"manifest":  "manifest",
						"requested": "manifest",
					},
				},
			},
			expected: &server.ManifestResponse{
				Details: server.ManifestDetails{
					Domain:       "example.com",
					EphemeralKey: "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p",
					MasterKey:    "nHUFE9prPXPrHcG3SkwP1UzAQbSphqyQkQK9ATXLZsfkezhhda3p",
					Seq:          1,
				},
				Manifest:  "manifest",
				Requested: "manifest",
			},
			expectedErr: nil,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"status": "error",
					"type":   "response",
					"error":  "invalidParams",
				},
			},
			expected:    nil,
			expectedErr: errors.New("invalidParams"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetManifest(&server.ManifestRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetServerState(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *server.StateResponse
		expectedErr    error
	}{
		{
			name: "successful response",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"state": map[string]any{
							"build_version":    "1.9.4",
							"complete_ledgers": "32570-6595042",
							"io_latency_ms":    1,
							"last_close": map[string]any{
								"converge_time": 2,
								"proposers":     4,
							},
							"load_factor":  1,
							"peers":        21,
							"pubkey_node":  "n9KwwpYCU3ctereLW9S48fKjK4rcsvYbHmjgiRXkgWReQR9nDjCw",
							"server_state": "proposing",
							"validated_ledger": map[string]any{
								"close_time": 638329271,
								"hash":       "4BC50C9B0D8515D3EAAE1E74B29A95804346C491EE1A95BF25E4AAB854A6A652",
								"seq":        6595042,
							},
							"validation_quorum": 4,
						},
					},
				},
			},
			expected: &server.StateResponse{
				State: servertypes.State{
					BuildVersion:     "1.9.4",
					CompleteLedgers:  "32570-6595042",
					IOLatencyMS:      1,
					LastClose:        servertypes.CloseState{ConvergeTime: 2, Proposers: 4},
					LoadFactor:       1,
					Peers:            21,
					PubkeyNode:       "n9KwwpYCU3ctereLW9S48fKjK4rcsvYbHmjgiRXkgWReQR9nDjCw",
					ServerState:      "proposing",
					ValidationQuorum: 4,
					ValidatedLedger: servertypes.LedgerState{
						CloseTime: 638329271,
						Hash:      "4BC50C9B0D8515D3EAAE1E74B29A95804346C491EE1A95BF25E4AAB854A6A652",
						Seq:       6595042,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"error":  "invalidParams",
					"status": "error",
					"type":   "response",
				},
			},
			expected:    nil,
			expectedErr: errors.New("invalidParams"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetServerState(&server.StateRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetAggregatePrice(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *oracle.GetAggregatePriceResponse
		expectedErr    error
	}{
		{
			name: "Valid aggregate price",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"median": "123.45",
						"time":   float64(1234567890),
					},
				},
			},
			expected: &oracle.GetAggregatePriceResponse{
				Median: "123.45",
				Time:   1234567890,
			},
			expectedErr: nil,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"error":  "invalidParams",
					"status": "error",
					"type":   "response",
				},
			},
			expected:    nil,
			expectedErr: errors.New("invalidParams"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetAggregatePrice(&oracle.GetAggregatePriceRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *utility.PingResponse
		expectedErr    error
	}{
		{
			name: "successful ping",
			serverMessages: []map[string]any{
				{
					"id": 1,
					"result": map[string]any{
						"role":      "full",
						"unlimited": true,
					},
				},
			},
			expected:    &utility.PingResponse{Role: "full", Unlimited: true},
			expectedErr: nil,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"error":  "invalidParams",
					"status": "error",
					"type":   "response",
				},
			},
			expected:    nil,
			expectedErr: errors.New("invalidParams"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.Ping(&utility.PingRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}

func TestClient_GetRandom(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *utility.RandomResponse
		expectedErr    error
	}{
		{
			name: "successful random",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"result": map[string]any{"random": "123ABC"},
				},
			},
			expected:    &utility.RandomResponse{Random: "123ABC"},
			expectedErr: nil,
		},
		{
			name: "error response",
			serverMessages: []map[string]any{
				{
					"id":     1,
					"error":  "invalidParams",
					"status": "error",
					"type":   "response",
				},
			},
			expected:    nil,
			expectedErr: errors.New("invalidParams"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.serverMessages)
			defer cleanup()

			result, err := cl.GetRandom(&utility.RandomRequest{})

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, but got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("Expected %+v, but got %+v", tt.expected, result)
			}
		})
	}
}
