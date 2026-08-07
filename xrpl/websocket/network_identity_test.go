package websocket

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	streamtypes "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestClientConnectDiscoversNetworkIdentity(t *testing.T) {
	tests := []struct {
		name                  string
		result                map[string]any
		responseError         string
		override              *uint32
		buildOverride         string
		expectedID            *uint32
		expectedBuild         string
		expectedErr           error
		expectedReportedError string
		expectedRequests      int32
		expectedConnections   int32
		preserveOverride      bool
		configuredIdentity    bool
	}{
		{
			name: "valid mainnet zero",
			result: map[string]any{"info": map[string]any{
				"network_id":    uint32(0),
				"build_version": "1.12.0",
			}},
			expectedID:          uint32Pointer(0),
			expectedBuild:       "1.12.0",
			expectedRequests:    1,
			expectedConnections: 1,
		},
		{
			name: "missing network ID remains unknown",
			result: map[string]any{"info": map[string]any{
				"build_version": "1.12.0",
			}},
			expectedBuild:       "1.12.0",
			expectedRequests:    1,
			expectedConnections: 1,
		},
		{
			name:                  "server_info error does not block Connect",
			responseError:         "noNetwork",
			expectedReportedError: "noNetwork",
			expectedRequests:      1,
			expectedConnections:   1,
		},
		{
			name: "matching override is preserved",
			result: map[string]any{"info": map[string]any{
				"network_id":    uint32(21337),
				"build_version": "1.12.0",
			}},
			override:            uint32Pointer(21337),
			buildOverride:       "1.10.0",
			expectedID:          uint32Pointer(21337),
			expectedBuild:       "1.12.0",
			expectedRequests:    1,
			expectedConnections: 1,
			preserveOverride:    true,
		},
		{
			name: "mismatching override is preserved",
			result: map[string]any{"info": map[string]any{
				"network_id":    uint32(21338),
				"build_version": "1.12.0",
			}},
			override:            uint32Pointer(21337),
			expectedErr:         ErrNetworkIDOverrideMismatch,
			expectedRequests:    1,
			expectedConnections: 1,
			preserveOverride:    true,
		},
		{
			name: "incomplete configured identity performs discovery",
			result: map[string]any{"info": map[string]any{
				"network_id":    uint32(21337),
				"build_version": "1.12.0",
			}},
			override:            uint32Pointer(21337),
			expectedID:          uint32Pointer(21337),
			expectedBuild:       "1.12.0",
			expectedRequests:    1,
			expectedConnections: 1,
			configuredIdentity:  true,
		},
		{
			name:                "complete trusted override bypasses discovery",
			override:            uint32Pointer(21337),
			buildOverride:       "1.12.0",
			expectedID:          uint32Pointer(21337),
			expectedBuild:       "1.12.0",
			expectedRequests:    0,
			expectedConnections: 1,
			configuredIdentity:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestCount atomic.Int32
			var connectionCount atomic.Int32
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
				if tt.expectedRequests > 0 && connectionNumber == 1 {
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
			if tt.configuredIdentity {
				config = config.WithNetworkIdentity(*tt.override, tt.buildOverride)
			}
			cl := NewClient(config)
			reportedErrors := make(chan error, 1)
			cl.OnError(func(reportedErr error) { reportedErrors <- reportedErr })
			if !tt.configuredIdentity {
				setTestNetworkIdentity(cl, tt.override, tt.buildOverride)
			}

			err = cl.Connect()
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.True(t, cl.IsConnected())
				networkID, buildVersion := cl.NetworkIdentity()
				if tt.expectedID == nil {
					require.Nil(t, networkID)
				} else {
					require.NotNil(t, networkID)
					require.Equal(t, *tt.expectedID, *networkID)
				}
				require.Equal(t, tt.expectedBuild, buildVersion)
			}

			if tt.expectedReportedError != "" {
				select {
				case reportedErr := <-reportedErrors:
					require.ErrorContains(t, reportedErr, tt.expectedReportedError)
				case <-time.After(time.Second):
					t.Fatal("Connect did not report the server_info error")
				}
			}

			require.Equal(t, tt.expectedRequests, requestCount.Load())
			require.Eventually(t, func() bool {
				return connectionCount.Load() == tt.expectedConnections
			}, time.Second, time.Millisecond)
			if err != nil {
				require.False(t, cl.IsConnected())
			} else {
				require.NoError(t, cl.Disconnect())
			}
			if tt.preserveOverride {
				networkID, _ := cl.NetworkIdentity()
				require.NotNil(t, networkID)
				require.Equal(t, *tt.override, *networkID)
			}
			select {
			case serverError := <-serverErr:
				require.NoError(t, serverError)
			default:
			}
		})
	}
}

func TestClientConnectBuffersFramesBeforeIdentityResponse(t *testing.T) {
	serverErr := make(chan error, 1)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()

		request := make(map[string]any)
		if err := conn.ReadJSON(&request); err != nil {
			serverErr <- err
			return
		}
		messages := []map[string]any{
			{"type": "ledgerClosed", "ledger_index": 1},
			{"id": uint64(999), "result": map[string]any{}},
			{
				"id": request["id"],
				"result": map[string]any{"info": map[string]any{
					"network_id":    uint32(21337),
					"build_version": "1.12.0",
				}},
			},
		}
		for _, message := range messages {
			if err := conn.WriteJSON(message); err != nil {
				serverErr <- err
				return
			}
		}
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithTimeout(time.Second))
	ledgerClosed := make(chan *streamtypes.LedgerStream, 1)
	cl.OnLedgerClosed(func(ledger *streamtypes.LedgerStream) {
		ledgerClosed <- ledger
	})

	require.NoError(t, cl.Connect())
	networkID, buildVersion := cl.NetworkIdentity()
	require.NotNil(t, networkID)
	require.Equal(t, uint32(21337), *networkID)
	require.Equal(t, "1.12.0", buildVersion)
	select {
	case ledger := <-ledgerClosed:
		require.EqualValues(t, 1, ledger.LedgerIndex)
	case <-time.After(time.Second):
		t.Fatal("Connect did not replay the buffered ledger stream")
	}
	require.NoError(t, cl.Disconnect())
	select {
	case err := <-serverErr:
		require.NoError(t, err)
	default:
	}
}

func TestClientNetworkIdentityReturnsSnapshot(t *testing.T) {
	cl := NewClient(NewClientConfig().WithNetworkIdentity(21337, "1.12.0"))

	networkID, buildVersion := cl.NetworkIdentity()
	require.NotNil(t, networkID)
	require.Equal(t, uint32(21337), *networkID)
	require.Equal(t, "1.12.0", buildVersion)

	*networkID = 1
	storedNetworkID, storedBuildVersion := cl.NetworkIdentity()
	require.NotNil(t, storedNetworkID)
	require.Equal(t, uint32(21337), *storedNetworkID)
	require.Equal(t, "1.12.0", storedBuildVersion)
}

func TestClientConnectDiscoveryTimeoutDoesNotBlockConnect(t *testing.T) {
	firstRequestRead := make(chan struct{})
	firstConnectionClosed := make(chan struct{})
	secondConnectionOpened := make(chan struct{})
	var connectionCount atomic.Int32
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		if connectionCount.Add(1) == 1 {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
			close(firstRequestRead)
			if _, _, err := conn.ReadMessage(); err != nil {
				close(firstConnectionClosed)
			}
			return
		}
		close(secondConnectionOpened)
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithTimeout(20 * time.Millisecond))
	reportedErrors := make(chan error, 1)
	cl.OnError(func(reportedErr error) { reportedErrors <- reportedErr })

	require.NoError(t, cl.Connect())
	require.True(t, cl.IsConnected())
	networkID, _ := cl.NetworkIdentity()
	require.Nil(t, networkID)
	select {
	case <-firstRequestRead:
	case <-time.After(time.Second):
		t.Fatal("server did not receive server_info request")
	}
	select {
	case <-firstConnectionClosed:
	case <-time.After(time.Second):
		t.Fatal("Connect did not replace the timed-out websocket")
	}
	select {
	case <-secondConnectionOpened:
	case <-time.After(time.Second):
		t.Fatal("Connect did not open a replacement websocket")
	}
	select {
	case reportedErr := <-reportedErrors:
		require.ErrorIs(t, reportedErr, ErrRequestTimedOut)
		var timeoutErr interface{ Timeout() bool }
		require.ErrorAs(t, reportedErr, &timeoutErr)
		require.True(t, timeoutErr.Timeout())
	case <-time.After(time.Second):
		t.Fatal("Connect did not report the server_info timeout")
	}
	require.NoError(t, cl.Disconnect())
}

func TestClientConnectRecoveryDialUsesTimeout(t *testing.T) {
	var connectionCount atomic.Int32
	firstRequestRead := make(chan struct{})
	secondHandshakeStarted := make(chan struct{})
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if connectionCount.Add(1) == 2 {
			close(secondHandshakeStarted)
			<-r.Context().Done()
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
		close(firstRequestRead)
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithTimeout(20 * time.Millisecond))

	err = cl.Connect()

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	select {
	case <-firstRequestRead:
	case <-time.After(time.Second):
		t.Fatal("server did not receive server_info request")
	}
	select {
	case <-secondHandshakeStarted:
	case <-time.After(time.Second):
		t.Fatal("Connect did not start the replacement handshake")
	}
	require.False(t, cl.IsConnected())
}

func TestClientReconnectKeepsDiscoveredNetworkIdentity(t *testing.T) {
	var connectionCount atomic.Int32
	secondConnectionOpened := make(chan struct{})
	unexpectedRequest := make(chan struct{}, 1)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		if connectionCount.Add(1) == 1 {
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
			return
		}

		close(secondConnectionOpened)
		if _, _, err := conn.ReadMessage(); err == nil {
			unexpectedRequest <- struct{}{}
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	t.Cleanup(swapReconnectDelays(time.Millisecond, time.Millisecond))
	cl := NewClient(NewClientConfig().WithHost(url).WithMaxReconnects(1).WithTimeout(time.Second))

	require.NoError(t, cl.Connect())
	select {
	case <-secondConnectionOpened:
	case <-time.After(time.Second):
		t.Fatal("client did not reconnect")
	}
	networkID, buildVersion := cl.NetworkIdentity()
	require.NotNil(t, networkID)
	require.Equal(t, uint32(1), *networkID)
	require.Equal(t, "1.12.0", buildVersion)
	require.Equal(t, int32(2), connectionCount.Load())
	select {
	case <-unexpectedRequest:
		t.Fatal("reconnect sent another server_info request")
	case <-time.After(20 * time.Millisecond):
	}
	require.NoError(t, cl.Disconnect())
}

func TestClientExplicitReconnectRefreshesUnknownNetworkIdentity(t *testing.T) {
	var connectionCount atomic.Int32
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
		info := map[string]any{"build_version": "1.12.0"}
		if connectionNumber == 2 {
			info["network_id"] = uint32(21337)
		}
		if err := conn.WriteJSON(map[string]any{
			"id":     request["id"],
			"result": map[string]any{"info": info},
		}); err != nil {
			return
		}
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithTimeout(time.Second))

	require.NoError(t, cl.Connect())
	networkID, buildVersion := cl.NetworkIdentity()
	require.Nil(t, networkID)
	require.Equal(t, "1.12.0", buildVersion)
	require.NoError(t, cl.Disconnect())

	require.NoError(t, cl.Connect())
	networkID, buildVersion = cl.NetworkIdentity()
	require.NotNil(t, networkID)
	require.Equal(t, uint32(21337), *networkID)
	require.Equal(t, "1.12.0", buildVersion)
	require.NoError(t, cl.Disconnect())
	require.Equal(t, int32(2), connectionCount.Load())
}

func TestClientExplicitReconnectRejectsNetworkIdentityChange(t *testing.T) {
	var connectionCount atomic.Int32
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
				"network_id":    uint32(connectionNumber),
				"build_version": "1.12.0",
			}},
		}); err != nil {
			return
		}
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithTimeout(time.Second))

	require.NoError(t, cl.Connect())
	networkID, buildVersion := cl.NetworkIdentity()
	require.NotNil(t, networkID)
	require.Equal(t, uint32(1), *networkID)
	require.Equal(t, "1.12.0", buildVersion)
	require.NoError(t, cl.Disconnect())

	err = cl.Connect()
	require.ErrorIs(t, err, ErrNetworkIDOverrideMismatch)
	require.False(t, cl.IsConnected())
	networkID, buildVersion = cl.NetworkIdentity()
	require.NotNil(t, networkID)
	require.Equal(t, uint32(1), *networkID)
	require.Equal(t, "1.12.0", buildVersion)
	require.Equal(t, int32(2), connectionCount.Load())
}

func TestClientAutofillOmitsNetworkIDWhenIdentityIsMissing(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	tx := transaction.FlatTransaction{
		"Account":            "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		"TransactionType":    "AccountSet",
		"Sequence":           uint32(1),
		"Fee":                "10",
		"LastLedgerSequence": uint32(100),
	}

	require.NoError(t, cl.Autofill(&tx))
	require.NotContains(t, tx, "NetworkID")
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

func TestClientConcurrentConnectKeepsOneSocket(t *testing.T) {
	var arrivals atomic.Int32
	var activeConnections atomic.Int32
	bothArrived := make(chan struct{})
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if arrivals.Add(1) == 2 {
			close(bothArrived)
		}
		select {
		case <-bothArrived:
		case <-time.After(time.Second):
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		activeConnections.Add(1)
		defer activeConnections.Add(-1)
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithNetworkIdentity(0, "1.12.0"))

	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			results <- cl.Connect()
		}()
	}
	close(start)

	var successCount int
	var alreadyConnectedCount int
	for range 2 {
		connectErr := <-results
		switch {
		case connectErr == nil:
			successCount++
		case errors.Is(connectErr, ErrAlreadyConnected):
			alreadyConnectedCount++
		default:
			t.Fatalf("unexpected concurrent Connect error: %v", connectErr)
		}
	}
	require.Equal(t, 1, successCount)
	require.Equal(t, 1, alreadyConnectedCount)
	require.True(t, cl.IsConnected())
	require.Eventually(t, func() bool {
		return activeConnections.Load() == 1
	}, time.Second, time.Millisecond)

	require.NoError(t, cl.Disconnect())
	require.Eventually(t, func() bool {
		return activeConnections.Load() == 0
	}, time.Second, time.Millisecond)
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
	t.Cleanup(swapReconnectDelays(time.Millisecond, time.Millisecond))
	cl := NewClient(NewClientConfig().
		WithHost(url).
		WithMaxReconnects(1).
		WithNetworkIdentity(0, "1.12.0"))

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

func TestClientGetSignedTxSkipsNetworkPolicyWhenAutofillDisabled(t *testing.T) {
	cl := NewClient(NewClientConfig().WithNetworkIdentity(21337, "1.12.0"))
	signer, err := wallet.FromSeed("sEd7io6yt5dFJrcePgRiFVHvmkJhJD1", "")
	require.NoError(t, err)
	tx := transaction.FlatTransaction{
		"Account":         signer.ClassicAddress.String(),
		"TransactionType": "AccountSet",
		"Fee":             "10",
		"Sequence":        uint32(1),
		"NetworkID":       uint32(1),
	}

	blob, err := cl.getSignedTx(tx, false, &signer)

	require.NoError(t, err)
	require.NotEmpty(t, blob)
	signedTx, err := binarycodec.Decode(blob)
	require.NoError(t, err)
	require.EqualValues(t, uint32(1), signedTx["NetworkID"])
	require.Equal(t, uint32(1), tx["NetworkID"])
}

func uint32Pointer(value uint32) *uint32 {
	return &value
}

func setTestNetworkIdentity(cl *Client, networkID *uint32, buildVersion string) {
	cl.identity.mu.Lock()
	defer cl.identity.mu.Unlock()
	cl.identity.current.NetworkID = nil
	if networkID != nil {
		cl.identity.current.NetworkID = uint32Pointer(*networkID)
	}
	cl.identity.current.BuildVersion = buildVersion
}

func setTrustedTestNetworkIdentity(cl *Client, networkID uint32) {
	cl.identity.mu.Lock()
	defer cl.identity.mu.Unlock()
	cl.identity.ready = true
	cl.identity.trusted = true
	cl.identity.current.NetworkID = uint32Pointer(networkID)
	cl.identity.current.BuildVersion = "1.12.0"
}
