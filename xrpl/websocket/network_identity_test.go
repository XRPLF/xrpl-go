package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func uint32Pointer(value uint32) *uint32 {
	return &value
}

func setTrustedTestNetworkIdentity(cl *Client, networkID uint32) {
	cl.NetworkID = uint32Pointer(networkID)
	cl.BuildVersion = "1.12.0"
	cl.identity.ready = true
	cl.identity.trusted = true
	cl.identity.current.NetworkID = uint32Pointer(networkID)
	cl.identity.current.BuildVersion = "1.12.0"
}

func TestClientConnectDiscoversNetworkIdentity(t *testing.T) {
	tests := []struct {
		name             string
		result           map[string]any
		responseError    string
		override         *uint32
		buildOverride    string
		expectedID       *uint32
		expectedBuild    string
		expectedErr      error
		expectedErrText  string
		expectedRequests int32
		preserveOverride bool
		trustedConfig    bool
	}{
		{
			name: "valid mainnet zero",
			result: map[string]any{"info": map[string]any{
				"network_id":    uint32(0),
				"build_version": "1.12.0",
			}},
			expectedID:       uint32Pointer(0),
			expectedBuild:    "1.12.0",
			expectedRequests: 1,
		},
		{
			name: "missing network ID",
			result: map[string]any{"info": map[string]any{
				"build_version": "1.12.0",
			}},
			expectedErr:      ErrNetworkIDUnavailable,
			expectedRequests: 1,
		},
		{
			name:             "server_info error",
			responseError:    "noNetwork",
			expectedErrText:  "noNetwork",
			expectedRequests: 1,
		},
		{
			name: "matching override is preserved",
			result: map[string]any{"info": map[string]any{
				"network_id":    uint32(21337),
				"build_version": "1.12.0",
			}},
			override:         uint32Pointer(21337),
			buildOverride:    "1.10.0",
			expectedID:       uint32Pointer(21337),
			expectedBuild:    "1.12.0",
			expectedRequests: 1,
			preserveOverride: true,
		},
		{
			name: "mismatching override is preserved",
			result: map[string]any{"info": map[string]any{
				"network_id":    uint32(21338),
				"build_version": "1.12.0",
			}},
			override:         uint32Pointer(21337),
			expectedErr:      ErrNetworkIDOverrideMismatch,
			expectedRequests: 1,
			preserveOverride: true,
		},
		{
			name:             "complete trusted override bypasses discovery",
			override:         uint32Pointer(21337),
			buildOverride:    "1.12.0",
			expectedID:       uint32Pointer(21337),
			expectedBuild:    "1.12.0",
			expectedRequests: 0,
			trustedConfig:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestCount atomic.Int32
			serverErr := make(chan error, 1)
			upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					serverErr <- err
					return
				}
				defer conn.Close()

				if tt.expectedRequests > 0 {
					request := make(map[string]any)
					if err := conn.ReadJSON(&request); err != nil {
						serverErr <- err
						return
					}
					requestCount.Add(1)
					response := map[string]any{"id": request["id"]}
					if tt.responseError != "" {
						response["error"] = tt.responseError
					} else {
						response["result"] = tt.result
					}
					if err := conn.WriteJSON(response); err != nil {
						serverErr <- err
						return
					}
				}

				for {
					if _, _, err := conn.ReadMessage(); err != nil {
						return
					}
				}
			}))
			defer server.Close()

			url, err := testutil.ConvertHTTPToWS(server.URL)
			require.NoError(t, err)
			config := NewClientConfig().WithHost(url).WithTimeout(200 * time.Millisecond)
			if tt.trustedConfig {
				config = config.WithNetworkIdentity(*tt.override, tt.buildOverride)
			}
			cl := NewClient(config)
			if !tt.trustedConfig {
				cl.NetworkID = tt.override
				cl.BuildVersion = tt.buildOverride
			}

			err = cl.Connect()
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else if tt.expectedErrText != "" {
				require.EqualError(t, err, tt.expectedErrText)
			} else {
				require.NoError(t, err)
				require.True(t, cl.IsConnected())
				require.NotNil(t, cl.NetworkID)
				require.Equal(t, *tt.expectedID, *cl.NetworkID)
				require.Equal(t, tt.expectedBuild, cl.BuildVersion)
			}

			if err != nil {
				require.False(t, cl.IsConnected())
			} else {
				require.NoError(t, cl.Disconnect())
			}
			require.Equal(t, tt.expectedRequests, requestCount.Load())
			if tt.preserveOverride {
				require.Same(t, tt.override, cl.NetworkID)
			}
			select {
			case serverError := <-serverErr:
				require.NoError(t, serverError)
			default:
			}
		})
	}
}

func TestClientConnectDiscoveryTimeoutIsAtomic(t *testing.T) {
	requestRead := make(chan struct{})
	connectionClosed := make(chan struct{})
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
		close(requestRead)
		if _, _, err := conn.ReadMessage(); err != nil {
			close(connectionClosed)
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithTimeout(20 * time.Millisecond))

	err = cl.Connect()
	require.ErrorIs(t, err, ErrRequestTimedOut)
	require.False(t, cl.IsConnected())
	require.Error(t, cl.lifecycleContext().Err())
	select {
	case <-requestRead:
	case <-time.After(time.Second):
		t.Fatal("server did not receive server_info request")
	}
	select {
	case <-connectionClosed:
	case <-time.After(time.Second):
		t.Fatal("failed Connect did not close the websocket")
	}
	var timeoutErr interface{ Timeout() bool }
	require.ErrorAs(t, err, &timeoutErr)
	require.True(t, timeoutErr.Timeout())
}

func TestClientReconnectRediscoversNetworkIdentity(t *testing.T) {
	var connectionCount atomic.Int32
	secondDiscovery := make(chan struct{})
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		connectionNumber := connectionCount.Add(1)
		request := make(map[string]any)
		if err := conn.ReadJSON(&request); err != nil {
			return
		}
		if err := conn.WriteJSON(map[string]any{
			"id": request["id"],
			"result": map[string]any{"info": map[string]any{
				"network_id":    uint32(1),
				"build_version": "1.12.0",
			}},
		}); err != nil {
			return
		}
		if connectionNumber == 1 {
			return
		}
		close(secondDiscovery)
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(withReconnectDelays(
		NewClientConfig().WithHost(url).WithMaxReconnects(1).WithTimeout(time.Second),
		time.Millisecond,
		time.Millisecond,
	))

	require.NoError(t, cl.Connect())
	defer cl.Disconnect()
	select {
	case <-secondDiscovery:
	case <-time.After(time.Second):
		t.Fatal("reconnect did not rediscover network identity")
	}
	require.Equal(t, int32(2), connectionCount.Load())
}

func TestClientReconnectKeepsSocketPrivateDuringIdentityDiscovery(t *testing.T) {
	var connectionCount atomic.Int32
	secondDiscoveryStarted := make(chan struct{})
	allowSecondDiscovery := make(chan struct{})
	serverErr := make(chan error, 1)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()

		connectionNumber := connectionCount.Add(1)
		request := make(map[string]any)
		if err := conn.ReadJSON(&request); err != nil {
			serverErr <- err
			return
		}
		if connectionNumber == 2 {
			close(secondDiscoveryStarted)
			<-allowSecondDiscovery
		}
		if err := conn.WriteJSON(map[string]any{
			"id": request["id"],
			"result": map[string]any{"info": map[string]any{
				"network_id":    uint32(1),
				"build_version": "1.12.0",
			}},
		}); err != nil {
			serverErr <- err
			return
		}
		if connectionNumber == 1 {
			return
		}

		request = make(map[string]any)
		if err := conn.ReadJSON(&request); err != nil {
			serverErr <- err
			return
		}
		if err := conn.WriteJSON(map[string]any{
			"id":     request["id"],
			"status": "success",
			"type":   "response",
			"result": map[string]any{},
		}); err != nil {
			serverErr <- err
			return
		}
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(withReconnectDelays(
		NewClientConfig().WithHost(url).WithMaxReconnects(1).WithTimeout(time.Second),
		time.Millisecond,
		time.Millisecond,
	))

	require.NoError(t, cl.Connect())
	defer func() {
		if cl.IsConnected() {
			_ = cl.Disconnect()
		}
	}()
	select {
	case <-secondDiscoveryStarted:
	case <-time.After(time.Second):
		t.Fatal("reconnect identity discovery did not start")
	}

	response, err := cl.Request(newAccountChannelsRequest())
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrNotConnectedToServer)

	close(allowSecondDiscovery)
	require.Eventually(t, cl.IsConnected, time.Second, time.Millisecond)
	response, err = cl.Request(newAccountChannelsRequest())
	require.NoError(t, err)
	require.NotNil(t, response)
	select {
	case err := <-serverErr:
		require.NoError(t, err)
	default:
	}
}

func TestClientDisconnectClaimsReconnectSocketBeforeCancellationInvalidation(t *testing.T) {
	cl := NewClient(NewClientConfig().WithHost("ws://unused"))
	socket := newFakeWebsocketConnection()
	cl.conn.mu.Lock()
	cl.conn.preparing = socket
	cl.conn.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	var invalidationErr error
	cl.streamHandlerStateMu.Lock()
	cl.ctx = ctx
	cl.cancel = func() {
		cancel()
		// This reproduces the identity write watcher invalidation before the
		// lifecycle cancellation call returns.
		invalidationErr = cl.conn.invalidateSocket(socket)
	}
	cl.streamHandlerStateMu.Unlock()

	require.NoError(t, cl.Disconnect())
	require.NoError(t, invalidationErr)
	require.ErrorIs(t, ctx.Err(), context.Canceled)
	require.False(t, cl.IsConnected())
	require.Equal(t, int32(1), socket.closeCount.Load())
	cl.conn.mu.Lock()
	require.Nil(t, cl.conn.preparing)
	require.Nil(t, cl.conn.disconnecting)
	cl.conn.mu.Unlock()
}

func TestClientDisconnectClosesSocketDuringReconnectIdentityDiscovery(t *testing.T) {
	var connectionCount atomic.Int32
	secondDiscoveryStarted := make(chan struct{})
	secondConnectionClosed := make(chan struct{})
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		connectionNumber := connectionCount.Add(1)
		request := make(map[string]any)
		if err := conn.ReadJSON(&request); err != nil {
			return
		}
		if connectionNumber == 1 {
			_ = conn.WriteJSON(map[string]any{
				"id": request["id"],
				"result": map[string]any{"info": map[string]any{
					"network_id":    uint32(1),
					"build_version": "1.12.0",
				}},
			})
			return
		}

		close(secondDiscoveryStarted)
		if _, _, err := conn.ReadMessage(); err != nil {
			close(secondConnectionClosed)
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(withReconnectDelays(
		NewClientConfig().WithHost(url).WithMaxReconnects(1).WithTimeout(time.Second),
		time.Millisecond,
		time.Millisecond,
	))

	require.NoError(t, cl.Connect())
	select {
	case <-secondDiscoveryStarted:
	case <-time.After(time.Second):
		t.Fatal("reconnect identity discovery did not start")
	}
	require.NoError(t, cl.Disconnect())
	require.False(t, cl.IsConnected())
	select {
	case <-secondConnectionClosed:
	case <-time.After(time.Second):
		t.Fatal("Disconnect did not close the reconnecting socket")
	}
}

func TestClientConnectRejectsAlreadyConnected(t *testing.T) {
	var connectionCount atomic.Int32
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		connectionCount.Add(1)
		defer conn.Close()
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithNetworkIdentity(0, "1.12.0"))

	require.NoError(t, cl.Connect())
	require.ErrorIs(t, cl.Connect(), ErrAlreadyConnected)
	require.True(t, cl.IsConnected())
	require.Equal(t, int32(1), connectionCount.Load())
	require.NoError(t, cl.Disconnect())
}

func TestClientDisconnectCancelsInFlightReconnectDial(t *testing.T) {
	var requestCount atomic.Int32
	reconnectStarted := make(chan struct{})
	allowReconnectResponse := make(chan struct{})
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestCount.Add(1) == 2 {
			close(reconnectStarted)
			<-allowReconnectResponse
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.Close()
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithMaxReconnects(1).
			WithNetworkIdentity(0, "1.12.0"),
		time.Millisecond,
		time.Millisecond,
	))

	require.NoError(t, cl.Connect())
	select {
	case <-reconnectStarted:
	case <-time.After(time.Second):
		t.Fatal("reconnect dial did not start")
	}
	disconnectErr := cl.Disconnect()
	if disconnectErr != nil {
		require.ErrorIs(t, disconnectErr, ErrNotConnected)
	}
	close(allowReconnectResponse)
	require.Eventually(t, func() bool { return !cl.IsConnected() }, time.Second, time.Millisecond)
}

func TestClientGetSignedTxFailsClosedWithoutAutofill(t *testing.T) {
	cl := NewClient(*NewClientConfig())

	_, err := cl.getSignedTx(
		transaction.FlatTransaction{"TransactionType": "AccountSet"},
		false,
		&wallet.Wallet{},
	)
	require.ErrorIs(t, err, ErrNetworkIDUnavailable)
}

func TestClientAutofillRejectsUnverifiedPublicNetworkIdentity(t *testing.T) {
	networkID := uint32(21337)
	cl := NewClient(*NewClientConfig())
	cl.NetworkID = &networkID
	cl.BuildVersion = "1.12.0"
	tx := transaction.FlatTransaction{
		"TransactionType":    "AccountSet",
		"Account":            "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		"Sequence":           uint32(1),
		"Fee":                "10",
		"LastLedgerSequence": uint32(100),
	}

	err := cl.Autofill(&tx)

	require.ErrorIs(t, err, ErrNetworkIDUnavailable)
	require.NotContains(t, tx, "NetworkID")
}

func TestClientAutofillAcceptsTrustedNetworkIdentity(t *testing.T) {
	cl := NewClient(NewClientConfig().WithNetworkIdentity(21337, "1.12.0"))
	tx := transaction.FlatTransaction{
		"TransactionType":    "AccountSet",
		"Account":            "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		"Sequence":           uint32(1),
		"Fee":                "10",
		"LastLedgerSequence": uint32(100),
	}

	err := cl.Autofill(&tx)

	require.NoError(t, err)
	require.Equal(t, uint32(21337), tx["NetworkID"])
}
