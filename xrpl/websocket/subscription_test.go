package websocket

import (
	"reflect"
	"testing"

	subscribe "github.com/Peersyst/xrpl-go/xrpl/queries/subscription"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
)

func TestClient_Subscribe(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *subscribe.Response
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"fee_base":          10,
					"fee_ref":           10,
					"ledger_hash":       "60EBABF55F6AB58864242CADA0B24FBEA027F2426917F39CA56576B335C0065A",
					"ledger_index":      14380380,
					"reserve_base":      20000000,
					"reserve_inc":       5000000,
					"server_status":     "full",
					"validated_ledgers": "32570-6595042",
				},
			}},
			expected: &subscribe.Response{
				FeeBase:          10,
				FeeRef:           10,
				LedgerHash:       "60EBABF55F6AB58864242CADA0B24FBEA027F2426917F39CA56576B335C0065A",
				LedgerIndex:      14380380,
				ReserveBase:      20000000,
				ReserveInc:       5000000,
				ServerStatus:     "full",
				ValidatedLedgers: "32570-6595042",
			},
			expectedErr: nil,
		},
		{
			name:           "Error response",
			serverMessages: []map[string]any{{"id": 1, "error": ErrIncorrectID.Error()}},
			expected:       nil,
			expectedErr:    ErrIncorrectID,
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

			if err := cl.Connect(); err != nil {
				t.Errorf("Error connecting to server: %v", err)
			}

			result, err := cl.Subscribe(&subscribe.Request{})

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

func TestClient_Unsubscribe(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages []map[string]any
		expected       *subscribe.UnsubscribeResponse
		expectedErr    error
	}{
		{
			name: "Successful response",
			serverMessages: []map[string]any{{
				"id": 1,
				"result": map[string]any{
					"status":  "success",
					"streams": []string{"ledger", "transactions"},
				},
			}},
			expected:    &subscribe.UnsubscribeResponse{},
			expectedErr: nil,
		},
		{
			name:           "Error response",
			serverMessages: []map[string]any{{"id": 1, "error": ErrIncorrectID.Error()}},
			expected:       nil,
			expectedErr:    ErrIncorrectID,
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

			if err := cl.Connect(); err != nil {
				t.Errorf("Error connecting to server: %v", err)
			}

			result, err := cl.Unsubscribe(&subscribe.UnsubscribeRequest{})

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
