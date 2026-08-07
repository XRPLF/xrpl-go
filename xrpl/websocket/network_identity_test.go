package websocket

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func uint32Pointer(value uint32) *uint32 {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}

func TestClientCurrentNetworkIdentityBeforeReady(t *testing.T) {
	cl := NewClient(*NewClientConfig())

	identity, err := cl.CurrentNetworkIdentity()

	require.ErrorIs(t, err, ErrNetworkIDUnavailable)
	require.Nil(t, identity.NetworkID)
	require.Empty(t, identity.BuildVersion)
}

func TestClientCurrentNetworkIdentityConcurrentRefresh(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	cl.storeDiscoveredNetworkIdentity(clientNetworkIdentity(1, "1.12.0"))

	const iterations = 1_000
	var invalidSnapshots atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})

	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := range iterations {
			if i%2 == 0 {
				cl.storeDiscoveredNetworkIdentity(clientNetworkIdentity(1, "1.12.0"))
				continue
			}
			cl.storeDiscoveredNetworkIdentity(clientNetworkIdentity(2, "1.13.0"))
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for range iterations {
			identity, err := cl.CurrentNetworkIdentity()
			if err != nil || identity.NetworkID == nil {
				invalidSnapshots.Add(1)
				continue
			}
			if (*identity.NetworkID == 1 && identity.BuildVersion != "1.12.0") ||
				(*identity.NetworkID == 2 && identity.BuildVersion != "1.13.0") ||
				(*identity.NetworkID != 1 && *identity.NetworkID != 2) {
				invalidSnapshots.Add(1)
			}
		}
	}()

	close(start)
	wg.Wait()
	require.Zero(t, invalidSnapshots.Load())
}

func clientNetworkIdentity(networkID uint32, buildVersion string) clientinternal.NetworkIdentity {
	return clientinternal.NetworkIdentity{
		NetworkID:    uint32Pointer(networkID),
		BuildVersion: buildVersion,
	}
}

func setTrustedTestNetworkIdentity(cl *Client, networkID uint32) {
	cl.NetworkID = uint32Pointer(networkID)
	cl.BuildVersion = "1.12.0"
	cl.identity.ready = true
	cl.identity.trusted = true
	cl.identity.current.NetworkID = uint32Pointer(networkID)
	cl.identity.current.BuildVersion = "1.12.0"
}

type reconnectGateRequest struct {
	started chan struct{}
}

func (r *reconnectGateRequest) Method() string {
	return "server_info"
}

func (r *reconnectGateRequest) Validate() error {
	close(r.started)
	return nil
}

func (r *reconnectGateRequest) APIVersion() int {
	return 2
}

func TestClientConnectDiscoversNetworkIdentity(t *testing.T) {
	tests := []struct {
		name                      string
		result                    map[string]any
		responseError             string
		override                  *uint32
		buildOverride             string
		expectedID                *uint32
		expectedBuild             string
		expectedErr               error
		expectedErrText           string
		expectedRequests          int32
		expectedNetworkIDRequired *bool
		preserveOverride          bool
		trustedConfig             bool
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
			name: "Clio rippled version fallback",
			result: map[string]any{"info": map[string]any{
				"network_id":      uint32(21337),
				"rippled_version": "1.12.0",
			}},
			expectedID:                uint32Pointer(21337),
			expectedBuild:             "1.12.0",
			expectedRequests:          1,
			expectedNetworkIDRequired: boolPointer(true),
		},
		{
			name: "build version preferred over rippled version",
			result: map[string]any{"info": map[string]any{
				"network_id":      uint32(21337),
				"build_version":   "1.10.0",
				"rippled_version": "1.12.0",
			}},
			expectedID:                uint32Pointer(21337),
			expectedBuild:             "1.10.0",
			expectedRequests:          1,
			expectedNetworkIDRequired: boolPointer(false),
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

				identity, identityErr := cl.networkIdentity()
				require.NoError(t, identityErr)
				require.Equal(t, *tt.expectedID, *identity.NetworkID)
				require.Equal(t, tt.expectedBuild, identity.BuildVersion)
				if tt.expectedNetworkIDRequired != nil {
					required, policyErr := clientinternal.NetworkIDRequired(identity)
					require.NoError(t, policyErr)
					require.Equal(t, *tt.expectedNetworkIDRequired, required)
				}
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
		networkID := uint32(1)
		buildVersion := "1.12.0"
		if connectionNumber == 2 {
			buildVersion = "1.13.0"
		}
		if err := conn.WriteJSON(map[string]any{
			"id": request["id"],
			"result": map[string]any{"info": map[string]any{
				"network_id":    networkID,
				"build_version": buildVersion,
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
	t.Cleanup(swapReconnectDelays(time.Millisecond, time.Millisecond))
	cl := NewClient(NewClientConfig().WithHost(url).WithMaxReconnects(1).WithTimeout(time.Second))

	require.NoError(t, cl.Connect())
	defer cl.Disconnect()
	select {
	case <-secondDiscovery:
	case <-time.After(time.Second):
		t.Fatal("reconnect did not rediscover network identity")
	}
	require.Eventually(t, func() bool {
		identity, err := cl.CurrentNetworkIdentity()
		return err == nil && identity.NetworkID != nil && identity.BuildVersion == "1.13.0"
	}, time.Second, time.Millisecond)
	require.Equal(t, int32(2), connectionCount.Load())

	// Exported fields retain the first discovery value for compatibility and to
	// keep concurrent reads safe after Connect.
	require.NotNil(t, cl.NetworkID)
	require.Equal(t, uint32(1), *cl.NetworkID)
	require.Equal(t, "1.12.0", cl.BuildVersion)

	identity, err := cl.CurrentNetworkIdentity()
	require.NoError(t, err)
	require.NotNil(t, identity.NetworkID)
	require.Equal(t, uint32(1), *identity.NetworkID)
	require.Equal(t, "1.13.0", identity.BuildVersion)

	// The public snapshot must not alias the client's internal pointer.
	*identity.NetworkID = 99
	identity, err = cl.CurrentNetworkIdentity()
	require.NoError(t, err)
	require.Equal(t, uint32(1), *identity.NetworkID)
}

func TestClientReconnectReportsLastNetworkIdentityFailure(t *testing.T) {
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
		if connectionNumber == 1 {
			info["network_id"] = uint32(1)
		}
		if err := conn.WriteJSON(map[string]any{
			"id":     request["id"],
			"result": map[string]any{"info": info},
		}); err != nil {
			return
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	t.Cleanup(swapReconnectDelays(time.Millisecond, time.Millisecond))
	cl := NewClient(NewClientConfig().WithHost(url).WithMaxReconnects(1).WithTimeout(time.Second))

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
		require.Equal(t, 1, maxErr.Attempts)
		require.ErrorIs(t, got, ErrNetworkIDUnavailable)
		require.ErrorIs(t, maxErr.Err, ErrNetworkIDUnavailable)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for reconnect identity failure")
	}
}

func TestClientReconnectBlocksRequestsDuringNetworkIdentityDiscovery(t *testing.T) {
	var connectionCount atomic.Int32
	closeFirstConnection := make(chan struct{})
	reconnectDiscoveryStarted := make(chan struct{})
	checkForEarlyRequest := make(chan struct{})
	earlyRequest := make(chan bool, 1)
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
		discoveryRequest := make(map[string]any)
		if err := conn.ReadJSON(&discoveryRequest); err != nil {
			serverErr <- err
			return
		}
		identityResponse := map[string]any{
			"id": discoveryRequest["id"],
			"result": map[string]any{"info": map[string]any{
				"network_id":    uint32(1),
				"build_version": "1.12.0",
			}},
		}

		if connectionNumber == 1 {
			if err := conn.WriteJSON(identityResponse); err != nil {
				serverErr <- err
				return
			}
			<-closeFirstConnection
			return
		}
		if connectionNumber != 2 {
			return
		}

		close(reconnectDiscoveryStarted)
		normalRequest := make(chan map[string]any, 1)
		normalRequestErr := make(chan error, 1)
		go func() {
			request := make(map[string]any)
			if err := conn.ReadJSON(&request); err != nil {
				normalRequestErr <- err
				return
			}
			normalRequest <- request
		}()

		<-checkForEarlyRequest
		var request map[string]any
		select {
		case request = <-normalRequest:
			earlyRequest <- true
			if err := conn.WriteJSON(map[string]any{"id": request["id"], "result": map[string]any{}}); err != nil {
				serverErr <- err
				return
			}
			if err := conn.WriteJSON(identityResponse); err != nil {
				serverErr <- err
			}
			return
		case err := <-normalRequestErr:
			serverErr <- err
			return
		case <-time.After(100 * time.Millisecond):
			earlyRequest <- false
		}

		if err := conn.WriteJSON(identityResponse); err != nil {
			serverErr <- err
			return
		}
		select {
		case request = <-normalRequest:
		case err := <-normalRequestErr:
			serverErr <- err
			return
		case <-time.After(time.Second):
			serverErr <- ErrRequestTimedOut
			return
		}
		if err := conn.WriteJSON(map[string]any{"id": request["id"], "result": map[string]any{}}); err != nil {
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
	t.Cleanup(swapReconnectDelays(time.Millisecond, time.Millisecond))
	cl := NewClient(NewClientConfig().WithHost(url).WithMaxReconnects(1).WithTimeout(time.Second))

	require.NoError(t, cl.Connect())
	defer cl.Disconnect()
	close(closeFirstConnection)
	select {
	case <-reconnectDiscoveryStarted:
	case <-time.After(time.Second):
		t.Fatal("reconnect identity discovery did not start")
	}

	requestStarted := make(chan struct{})
	requestDone := make(chan error, 1)
	go func() {
		_, err := cl.Request(&reconnectGateRequest{started: requestStarted})
		requestDone <- err
	}()
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("normal request did not start")
	}
	close(checkForEarlyRequest)

	select {
	case arrivedEarly := <-earlyRequest:
		require.False(t, arrivedEarly, "normal request reached the socket before identity discovery completed")
	case <-time.After(time.Second):
		t.Fatal("server did not check for an early request")
	}
	select {
	case err := <-requestDone:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("normal request did not complete after identity discovery")
	}
	require.True(t, cl.IsConnected())
	require.Equal(t, int32(2), connectionCount.Load())
	select {
	case err := <-serverErr:
		require.NoError(t, err)
	default:
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

func TestClientConcurrentConnectKeepsOneSocket(t *testing.T) {
	var connectionCount atomic.Int32
	identityRequestReceived := make(chan struct{})
	releaseIdentityResponse := make(chan struct{})
	var releaseIdentityResponseOnce sync.Once
	releaseIdentityHandshake := func() {
		releaseIdentityResponseOnce.Do(func() {
			close(releaseIdentityResponse)
		})
	}
	t.Cleanup(releaseIdentityHandshake)
	serverErr := make(chan error, 1)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()

		if connectionCount.Add(1) != 1 {
			return
		}
		request := make(map[string]any)
		if err := conn.ReadJSON(&request); err != nil {
			serverErr <- err
			return
		}
		close(identityRequestReceived)
		<-releaseIdentityResponse
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
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithTimeout(time.Second))

	firstResult := make(chan error, 1)
	go func() {
		firstResult <- cl.Connect()
	}()
	select {
	case <-identityRequestReceived:
	case <-time.After(time.Second):
		t.Fatal("first Connect did not start identity discovery")
	}

	secondStarted := make(chan struct{})
	secondResult := make(chan error, 1)
	go func() {
		close(secondStarted)
		secondResult <- cl.Connect()
	}()
	<-secondStarted
	select {
	case connectErr := <-secondResult:
		t.Fatalf("second Connect returned before the first identity handshake completed: %v", connectErr)
	default:
	}

	releaseIdentityHandshake()
	var firstErr error
	select {
	case firstErr = <-firstResult:
	case <-time.After(time.Second):
		t.Fatal("first Connect did not complete")
	}
	var secondErr error
	select {
	case secondErr = <-secondResult:
	case <-time.After(time.Second):
		t.Fatal("second Connect did not complete")
	}

	require.NoError(t, firstErr)
	require.ErrorIs(t, secondErr, ErrAlreadyConnected)
	require.True(t, cl.IsConnected())
	require.Equal(t, int32(1), connectionCount.Load())
	require.NoError(t, cl.Disconnect())
	select {
	case err := <-serverErr:
		require.NoError(t, err)
	default:
	}
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
