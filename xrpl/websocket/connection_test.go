package websocket

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	gorillaws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestConnection_ReadMessageEnforcesMaxResponseSize(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		maxResponseSize int64
		expectedErr     error
	}{
		{
			name:            "fail - rejects message over max size",
			message:         strings.Repeat("a", 33),
			maxResponseSize: 32,
			expectedErr:     gorillaws.ErrReadLimit,
		},
		{
			name:            "pass - allows message at max size",
			message:         strings.Repeat("a", 32),
			maxResponseSize: 32,
		},
		{
			name:            "pass - zero max size disables limit",
			message:         strings.Repeat("a", 33),
			maxResponseSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newMessageServer(t, tt.message)
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
			conn := newConnection(wsURL, tt.maxResponseSize)
			require.NoError(t, conn.Connect())
			defer func() {
				_ = conn.Disconnect()
			}()

			got, err := conn.ReadMessage()

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.message, string(got))
		})
	}
}

func newMessageServer(t *testing.T, msg string) *httptest.Server {
	t.Helper()

	upgrader := gorillaws.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		defer conn.Close()

		if err := conn.WriteMessage(gorillaws.TextMessage, []byte(msg)); err != nil {
			t.Errorf("write websocket message: %v", err)
		}
	}))
}

// Exercises the fix that serializes concurrent ReadMessage calls under readMu.
// Run with -race to expose a missing mutex. A lucky-scheduled run can pass without it.
func TestConnection_ReadMessageSerializesConcurrentReaders(t *testing.T) {
	readyToWrite := make(chan struct{})
	serverErr := make(chan error, 1)

	ws := &testutil.MockWebSocketServer{}
	server := ws.TestWebSocketServer(func(serverConn *gorillaws.Conn) {
		defer serverConn.Close()
		<-readyToWrite

		for _, msg := range []string{"first", "second"} {
			if err := serverConn.WriteMessage(gorillaws.TextMessage, []byte(msg)); err != nil {
				serverErr <- err
				return
			}
		}
	})
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)

	conn := NewConnection(url)
	require.NoError(t, conn.Connect())
	defer conn.Disconnect()

	type readResult struct {
		message []byte
		err     error
	}

	results := make(chan readResult, 2)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			<-start
			message, err := conn.ReadMessage()
			results <- readResult{
				message: message,
				err:     err,
			}
		})
	}

	close(start)
	close(readyToWrite)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case err := <-serverErr:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for concurrent reads")
	}

	close(results)

	messages := make([]string, 0, 2)
	for result := range results {
		require.NoError(t, result.err)
		messages = append(messages, string(result.message))
	}
	require.ElementsMatch(t, []string{"first", "second"}, messages)
}

func TestConnection_DisconnectUnblocksReadMessage(t *testing.T) {
	ws := &testutil.MockWebSocketServer{}
	server := ws.TestWebSocketServer(func(serverConn *gorillaws.Conn) {
		defer serverConn.Close()
		// Block until the client closes the connection.
		_, _, _ = serverConn.ReadMessage()
	})
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)

	conn := NewConnection(url)
	require.NoError(t, conn.Connect())

	done := make(chan error, 1)
	go func() {
		_, err := conn.ReadMessage()
		done <- err
	}()

	// Give the goroutine time to enter the underlying read before disconnecting.
	time.Sleep(50 * time.Millisecond)
	require.NoError(t, conn.Disconnect())

	select {
	case err := <-done:
		require.Error(t, err, "ReadMessage should return an error after Disconnect")
	case <-time.After(time.Second):
		t.Fatal("ReadMessage did not return after Disconnect, possible goroutine leak")
	}
}

func TestConnection_WriteMessageHonorsCanceledContext(t *testing.T) {
	t.Run("already canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := newConnection("ws://unused", defaultMaxResponseSize).writeMessage(ctx, []byte("test"), time.Second)
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("canceled while waiting for another writer", func(t *testing.T) {
		connection := newConnection("ws://unused", defaultMaxResponseSize)
		socket := newFakeWebsocketConnection()
		connection.conn = socket
		require.NoError(t, connection.acquireWrite(context.Background()))
		defer connection.releaseWrite()

		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan error, 1)
		started := make(chan struct{})
		go func() {
			close(started)
			result <- connection.writeMessage(ctx, []byte("test"), time.Second)
		}()
		<-started
		cancel()

		select {
		case err := <-result:
			require.ErrorIs(t, err, context.Canceled)
		case <-time.After(time.Second):
			t.Fatal("write did not exit after context cancellation while waiting for the writer token")
		}
		require.Zero(t, socket.closeCount.Load())
		require.True(t, connection.IsConnected())
	})

	t.Run("canceled during socket write clears deadline", func(t *testing.T) {
		connection := newConnection("ws://unused", defaultMaxResponseSize)
		socket := newFakeWebsocketConnection()
		connection.conn = socket

		ctx, cancel := context.WithCancel(context.Background())
		socket.writeHook = cancel

		err := connection.writeMessage(ctx, []byte("test"), time.Second)
		require.ErrorIs(t, err, context.Canceled)
		require.Len(t, socket.writeDeadlines, 2)
		require.False(t, socket.writeDeadlines[0].IsZero())
		require.True(t, socket.writeDeadlines[1].IsZero())
	})

	t.Run("canceled after completed write keeps socket", func(t *testing.T) {
		connection := newConnection("ws://unused", defaultMaxResponseSize)
		socket := newFakeWebsocketConnection()
		connection.conn = socket

		ctx, cancel := context.WithCancel(context.Background())
		require.NoError(t, connection.writeMessage(ctx, []byte("test"), time.Second))
		cancel()

		require.Zero(t, socket.closeCount.Load())
		require.True(t, connection.IsConnected())
	})
}

func TestConnection_WriteFailureInvalidatesSocket(t *testing.T) {
	testErr := errors.New("socket write failure")
	tests := []struct {
		name      string
		configure func(*fakeWebsocketConnection)
	}{
		{
			name: "initial write deadline failure",
			configure: func(socket *fakeWebsocketConnection) {
				socket.initialDeadlineErr = testErr
			},
		},
		{
			name: "write failure",
			configure: func(socket *fakeWebsocketConnection) {
				socket.writeErr = testErr
			},
		},
		{
			name: "deadline clear failure",
			configure: func(socket *fakeWebsocketConnection) {
				socket.clearDeadlineErr = testErr
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connection := newConnection("ws://unused", defaultMaxResponseSize)
			failedSocket := newFakeWebsocketConnection()
			tt.configure(failedSocket)
			connection.conn = failedSocket

			err := connection.writeMessage(context.Background(), []byte("test"), time.Second)
			require.ErrorIs(t, err, testErr)
			require.False(t, connection.IsConnected())
			require.GreaterOrEqual(t, failedSocket.closeCount.Load(), int32(1))

			replacement := newFakeWebsocketConnection()
			connection.mu.Lock()
			connection.conn = replacement
			connection.mu.Unlock()
			require.NoError(t, connection.WriteMessage([]byte("replacement")))
			require.True(t, connection.IsConnected())
			require.Equal(t, int32(1), replacement.writeCount.Load())
		})
	}
}

func TestConnection_StaleWriteFailureDoesNotInvalidateReplacement(t *testing.T) {
	connection := newConnection("ws://unused", defaultMaxResponseSize)
	oldSocket := newFakeWebsocketConnection()
	oldSocket.writeErr = errors.New("old socket failed")
	oldSocket.writeRelease = make(chan struct{})
	connection.conn = oldSocket

	result := make(chan error, 1)
	go func() {
		result <- connection.WriteMessage([]byte("old"))
	}()
	<-oldSocket.writeStarted

	replacement := newFakeWebsocketConnection()
	connection.mu.Lock()
	connection.conn = replacement
	connection.mu.Unlock()
	close(oldSocket.writeRelease)

	require.ErrorIs(t, <-result, oldSocket.writeErr)
	require.True(t, connection.IsConnected())
	require.GreaterOrEqual(t, oldSocket.closeCount.Load(), int32(1))
	require.Zero(t, replacement.closeCount.Load())
	require.NoError(t, connection.WriteMessage([]byte("replacement")))
}

func TestConnection_SimultaneousCancellationAndWriteFailureInvalidatesExactSocket(t *testing.T) {
	connection := newConnection("ws://unused", defaultMaxResponseSize)
	failedSocket := newFakeWebsocketConnection()
	failedSocket.writeRelease = make(chan struct{})
	replacement := newFakeWebsocketConnection()
	failedSocket.closeHook = func() {
		connection.mu.Lock()
		connection.conn = replacement
		connection.mu.Unlock()
	}
	connection.conn = failedSocket

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- connection.writeMessage(ctx, []byte("test"), time.Second)
	}()
	<-failedSocket.writeStarted
	cancel()

	require.ErrorIs(t, <-result, context.Canceled)
	require.GreaterOrEqual(t, failedSocket.closeCount.Load(), int32(2))
	require.Zero(t, replacement.closeCount.Load())
	require.True(t, connection.IsConnected())
}

func TestConnection_CanceledActiveWriteInvalidatesSocket(t *testing.T) {
	connection := newConnection("ws://unused", defaultMaxResponseSize)
	socket := newFakeWebsocketConnection()
	socket.writeRelease = make(chan struct{})
	connection.conn = socket

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- connection.writeMessage(ctx, []byte("test"), time.Second)
	}()
	<-socket.writeStarted
	cancel()

	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("active write did not exit after context cancellation")
	}
	require.False(t, connection.IsConnected())
	require.GreaterOrEqual(t, socket.closeCount.Load(), int32(1))
}

func TestClient_ActiveWriteInvalidationReconnects(t *testing.T) {
	var dialCount atomic.Int32
	upgrader := gorillaws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dialCount.Add(1)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			var request struct {
				ID uint64 `json:"id"`
			}
			if err := conn.ReadJSON(&request); err != nil {
				return
			}
			if err := conn.WriteJSON(map[string]any{
				"id":     request.ID,
				"status": "success",
				"type":   "response",
				"result": map[string]any{},
			}); err != nil {
				return
			}
		}
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

	failedSocket := newFakeWebsocketConnection()
	failedSocket.readResults = make(chan fakeReadResult)
	failedSocket.writeRelease = make(chan struct{})
	client.conn.conn = failedSocket

	ctx := client.resetLifecycle()
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		client.readMessages(ctx)
	}()
	<-failedSocket.readStarted
	defer func() {
		require.NoError(t, client.Disconnect())
		select {
		case <-readDone:
		case <-time.After(time.Second):
			t.Fatal("read loop did not stop")
		}
	}()

	writeCtx, cancelWrite := context.WithCancel(t.Context())
	writeDone := make(chan error, 1)
	go func() {
		writeDone <- client.conn.writeMessage(writeCtx, []byte("request"), time.Second)
	}()
	<-failedSocket.writeStarted
	cancelWrite()
	require.ErrorIs(t, <-writeDone, context.Canceled)

	require.Eventually(t, func() bool {
		return dialCount.Load() == 1 && client.IsConnected()
	}, time.Second, time.Millisecond)

	response, err := client.Request(newAccountChannelsRequest())
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestClient_ManualReplacementFailsPendingOldSocket(t *testing.T) {
	upgrader := gorillaws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			var request struct {
				ID uint64 `json:"id"`
			}
			if err := conn.ReadJSON(&request); err != nil {
				return
			}
			if err := conn.WriteJSON(map[string]any{
				"id":     request.ID,
				"status": "success",
				"type":   "response",
				"result": map[string]any{},
			}); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	client := NewClient(NewClientConfig().WithHost(url).WithTimeout(time.Second))
	setTrustedTestNetworkIdentity(client, 0)

	oldSocket := newFakeWebsocketConnection()
	oldSocket.readRelease = make(chan struct{})
	oldSocket.readCloseObserved = make(chan struct{})
	oldSocket.readCloseRelease = make(chan struct{})
	oldSocket.writeRelease = make(chan struct{})
	client.conn.conn = oldSocket

	oldCtx := client.resetLifecycle()
	oldReaderDone := make(chan struct{})
	go func() {
		defer close(oldReaderDone)
		client.readMessages(oldCtx)
	}()
	<-oldSocket.readStarted

	var releaseOldReader sync.Once
	defer func() {
		releaseOldReader.Do(func() {
			close(oldSocket.readCloseRelease)
		})
		require.NoError(t, client.Disconnect())
	}()

	const oldResponseID = uint64(901)
	oldPending := client.registerPendingResponse(oldResponseID, oldSocket)
	defer client.unregisterPendingResponse(oldResponseID)

	writeCtx, cancelWrite := context.WithCancel(t.Context())
	writeDone := make(chan error, 1)
	go func() {
		writeDone <- client.conn.writeMessage(writeCtx, []byte("request"), time.Second)
	}()
	<-oldSocket.writeStarted
	cancelWrite()
	require.ErrorIs(t, <-writeDone, context.Canceled)
	select {
	case <-oldSocket.readCloseObserved:
	case <-time.After(time.Second):
		t.Fatal("old reader did not observe the closed socket")
	}

	require.NoError(t, client.Connect())

	responseCtx, cancelResponse := context.WithTimeout(t.Context(), time.Second)
	defer cancelResponse()
	response, err := client.awaitResponse(responseCtx, oldPending)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrDisconnected)

	releaseOldReader.Do(func() {
		close(oldSocket.readCloseRelease)
	})
	select {
	case <-oldReaderDone:
	case <-time.After(time.Second):
		t.Fatal("old reader did not exit")
	}

	response, err = client.Request(newAccountChannelsRequest())
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestClient_ConcurrentFailurePathsUseSingleReconnect(t *testing.T) {
	var dialCount atomic.Int32
	unexpectedDial := make(chan struct{}, 1)
	upgrader := gorillaws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if dialCount.Add(1) > 1 {
			select {
			case unexpectedDial <- struct{}{}:
			default:
			}
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		messageType, message, readErr := conn.ReadMessage()
		if readErr == nil {
			t.Errorf("unexpected replacement message type %d: %s", messageType, message)
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cfg := withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithMaxReconnects(2).
			WithTimeout(time.Second),
		20*time.Millisecond,
		20*time.Millisecond,
	)
	client := NewClient(cfg)
	setTrustedTestNetworkIdentity(client, 0)

	failedSocket := newFakeWebsocketConnection()
	failedSocket.readErr = errors.New("socket failed")
	failedSocket.readRelease = make(chan struct{})
	client.conn.conn = failedSocket
	ctx := client.resetLifecycle()

	var readers sync.WaitGroup
	for range 2 {
		readers.Go(func() {
			client.readMessages(ctx)
		})
	}
	<-failedSocket.readStarted
	close(failedSocket.readRelease)

	require.Eventually(t, func() bool {
		return dialCount.Load() == 1 && client.IsConnected()
	}, time.Second, time.Millisecond)

	// The reconnect delay is 20ms. This window gives the second failure path
	// time to pass through connectionHandshakeMu after the first path publishes
	// the replacement.
	select {
	case <-unexpectedDial:
		t.Fatal("concurrent failure path opened a second replacement connection")
	case <-time.After(100 * time.Millisecond):
	}
	require.Equal(t, int32(1), dialCount.Load())

	require.NoError(t, client.Disconnect())
	readersDone := make(chan struct{})
	go func() {
		readers.Wait()
		close(readersDone)
	}()
	select {
	case <-readersDone:
	case <-time.After(time.Second):
		t.Fatal("read loops did not stop")
	}
}

func TestClient_ExplicitDisconnectDoesNotReconnect(t *testing.T) {
	dialAttempted := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case dialAttempted <- struct{}{}:
		default:
		}
		http.Error(w, "unexpected reconnect", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cfg := withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithMaxReconnects(1),
		time.Millisecond,
		time.Millisecond,
	)
	client := NewClient(cfg)
	setTrustedTestNetworkIdentity(client, 0)

	socket := newFakeWebsocketConnection()
	socket.readResults = make(chan fakeReadResult)
	client.conn.conn = socket
	ctx := client.resetLifecycle()
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		client.readMessages(ctx)
	}()
	<-socket.readStarted

	require.NoError(t, client.Disconnect())
	select {
	case <-readDone:
	case <-time.After(time.Second):
		t.Fatal("read loop did not stop")
	}
	select {
	case <-dialAttempted:
		t.Fatal("explicit disconnect started a reconnect")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestClient_ManualConnectCancelsOldReaderBeforePublishing(t *testing.T) {
	serverConnected := make(chan struct{})
	writeResponse := make(chan struct{})
	responseWritten := make(chan struct{})
	stopServer := make(chan struct{})
	upgrader := gorillaws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		close(serverConnected)

		select {
		case <-writeResponse:
		case <-stopServer:
			return
		}
		if err := conn.WriteJSON(map[string]any{
			"id":     uint64(77),
			"status": "success",
			"type":   "response",
			"result": map[string]any{},
		}); err != nil {
			t.Errorf("write replacement response: %v", err)
			return
		}
		close(responseWritten)
		messageType, message, readErr := conn.ReadMessage()
		if readErr == nil {
			t.Errorf("unexpected client message type %d: %s", messageType, message)
		}
	}))
	defer server.Close()
	defer close(stopServer)

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cfg := withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithMaxReconnects(1).
			WithTimeout(time.Second),
		200*time.Millisecond,
		200*time.Millisecond,
	)
	client := NewClient(cfg)
	setTrustedTestNetworkIdentity(client, 0)

	oldSocket := newFakeWebsocketConnection()
	oldSocket.readErr = errors.New("old socket failed")
	oldSocket.readRelease = make(chan struct{})
	client.conn.conn = oldSocket
	client.OnError(func(error) {})
	oldCtx := client.resetLifecycle()
	oldReaderDone := make(chan struct{})
	go func() {
		defer close(oldReaderDone)
		client.readMessages(oldCtx)
	}()
	<-oldSocket.readStarted

	client.streamHandlerResetMu.Lock()
	resetLocked := true
	defer func() {
		if resetLocked {
			client.streamHandlerResetMu.Unlock()
		}
		require.NoError(t, client.Disconnect())
	}()
	close(oldSocket.readRelease)
	require.Eventually(t, func() bool {
		return client.conn.currentSocket() == nil
	}, time.Second, time.Millisecond)

	connectResult := make(chan error, 1)
	go func() {
		connectResult <- client.Connect()
	}()
	select {
	case <-serverConnected:
	case <-time.After(time.Second):
		t.Fatal("manual Connect did not reach the replacement server")
	}
	require.Eventually(t, client.IsConnected, time.Second, time.Millisecond)

	// Connect is blocked in resetLifecycle. The old reader must already have
	// stopped because its lifecycle was canceled before socket publication.
	select {
	case <-oldReaderDone:
	case <-time.After(time.Second):
		t.Fatal("old reader remained active after replacement publication")
	}
	oldRunnerTracked := func() bool {
		client.errorStream.stateMu.Lock()
		defer client.errorStream.stateMu.Unlock()
		return client.errorStream.done != nil
	}()
	require.True(t, oldRunnerTracked, "manual Connect detached old handler runners before lifecycle reset")

	const responseID = uint64(77)
	pending := client.registerPendingResponse(responseID, client.conn.currentSocket())
	defer client.unregisterPendingResponse(responseID)
	close(writeResponse)
	select {
	case <-responseWritten:
	case <-time.After(time.Second):
		t.Fatal("server did not write the first replacement response")
	}

	client.streamHandlerResetMu.Unlock()
	resetLocked = false
	require.NoError(t, <-connectResult)

	responseCtx, cancelResponse := context.WithTimeout(t.Context(), time.Second)
	defer cancelResponse()
	response, err := client.awaitResponse(responseCtx, pending)
	require.NoError(t, err)
	require.Equal(t, responseID, response.ID)
}

func TestClient_AutomaticReconnectWinsWithoutLifecycleCancellation(t *testing.T) {
	upgrader := gorillaws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			var request struct {
				ID uint64 `json:"id"`
			}
			if err := conn.ReadJSON(&request); err != nil {
				return
			}
			if err := conn.WriteJSON(map[string]any{
				"id":     request.ID,
				"status": "success",
				"type":   "response",
				"result": map[string]any{},
			}); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cfg := withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithMaxReconnects(1).
			WithTimeout(time.Second),
		time.Millisecond,
		time.Millisecond,
	)
	client := NewClient(cfg)
	setTrustedTestNetworkIdentity(client, 0)

	failedSocket := newFakeWebsocketConnection()
	failedSocket.readErr = errors.New("old socket failed")
	client.conn.conn = failedSocket
	ctx := client.resetLifecycle()
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		client.readMessages(ctx)
	}()
	defer func() {
		require.NoError(t, client.Disconnect())
		select {
		case <-readDone:
		case <-time.After(time.Second):
			t.Fatal("read loop did not stop")
		}
	}()

	require.Eventually(t, client.IsConnected, time.Second, time.Millisecond)
	require.ErrorIs(t, client.Connect(), ErrAlreadyConnected)
	require.NoError(t, ctx.Err())

	response, err := client.Request(newAccountChannelsRequest())
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestClient_ConnectTimeoutBoundsWebSocketHandshake(t *testing.T) {
	const connectTimeout = 50 * time.Millisecond
	requestStarted := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requestStarted <- struct{}{}
		time.Sleep(time.Second)
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	client := NewClient(
		NewClientConfig().
			WithHost(url).
			WithTimeout(connectTimeout).
			WithNetworkIdentity(0, "2.0.0"),
	)

	startedAt := time.Now()
	err = client.Connect()
	elapsed := time.Since(startedAt)

	require.Error(t, err)
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("WebSocket handshake did not start")
	}
	require.Less(t, elapsed, 5*connectTimeout)
	require.False(t, client.IsConnected())
}

func TestClient_AutomaticReconnectTimeoutBoundsWebSocketHandshake(t *testing.T) {
	const connectTimeout = 50 * time.Millisecond
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(time.Second)
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	client := NewClient(withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithMaxReconnects(1).
			WithTimeout(connectTimeout).
			WithNetworkIdentity(0, "2.0.0"),
		time.Millisecond,
		time.Millisecond,
	))
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	retryCount := 0

	startedAt := time.Now()
	connected := client.reconnectWithBackoff(ctx, &retryCount, 1)
	elapsed := time.Since(startedAt)

	require.False(t, connected)
	require.Equal(t, 1, retryCount)
	require.Less(t, elapsed, 5*connectTimeout)
	require.False(t, client.IsConnected())
}

func TestClient_ConnectFailureDoesNotCancelLifecycleBeforePublication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "connection rejected", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	client := NewClient(NewClientConfig().WithHost(url))
	setTrustedTestNetworkIdentity(client, 0)
	ctx := client.resetLifecycle()
	beforePublishCalled := false

	bufferedMessages, err := client.connect(ctx, func() {
		beforePublishCalled = true
	})
	require.Error(t, err)
	require.Empty(t, bufferedMessages)
	require.False(t, beforePublishCalled)
	require.NoError(t, ctx.Err())
	require.False(t, client.IsConnected())
}

func TestClient_StaleReaderReturnsBeforeReplacementResponse(t *testing.T) {
	client := NewClient(NewClientConfig().WithMaxReconnects(0))
	oldSocket := newFakeWebsocketConnection()
	oldSocket.readErr = errors.New("old socket failed")
	oldSocket.readRelease = make(chan struct{})
	client.conn.conn = oldSocket

	oldReaderDone := make(chan struct{})
	go func() {
		defer close(oldReaderDone)
		client.readMessages(context.Background())
	}()
	<-oldSocket.readStarted

	replacement := newFakeWebsocketConnection()
	replacement.readResults = make(chan fakeReadResult)
	client.conn.mu.Lock()
	client.conn.conn = replacement
	client.conn.mu.Unlock()
	close(oldSocket.readRelease)

	select {
	case <-oldReaderDone:
	case <-time.After(time.Second):
		t.Fatal("stale reader attempted to read from the replacement socket")
	}
	require.Zero(t, replacement.closeCount.Load())

	const responseID = uint64(91)
	pending := client.registerPendingResponse(responseID, replacement)
	defer client.unregisterPendingResponse(responseID)
	newReaderCtx, cancelNewReader := context.WithCancel(context.Background())
	newReaderDone := make(chan struct{})
	go func() {
		defer close(newReaderDone)
		client.readMessages(newReaderCtx)
	}()
	<-replacement.readStarted
	replacement.readResults <- fakeReadResult{message: []byte(`{"id":91,"type":"response","status":"success","result":{}}`)}

	responseCtx, cancelResponse := context.WithTimeout(context.Background(), time.Second)
	defer cancelResponse()
	response, err := client.awaitResponse(responseCtx, pending)
	require.NoError(t, err)
	require.Equal(t, responseID, response.ID)

	cancelNewReader()
	require.NoError(t, replacement.Close())
	select {
	case <-newReaderDone:
	case <-time.After(time.Second):
		t.Fatal("replacement reader did not exit")
	}
}

type fakeReadResult struct {
	message []byte
	err     error
}

type fakeWebsocketConnection struct {
	initialDeadlineErr   error
	clearDeadlineErr     error
	writeErr             error
	readErr              error
	readStarted          chan struct{}
	readRelease          chan struct{}
	readResults          chan fakeReadResult
	readCloseObserved    chan struct{}
	readCloseRelease     chan struct{}
	writeStarted         chan struct{}
	writeRelease         chan struct{}
	closed               chan struct{}
	closeHook            func()
	writeHook            func()
	writeDeadlines       []time.Time
	closeOnce            sync.Once
	readStartOnce        sync.Once
	readCloseObserveOnce sync.Once
	writeStartOnce       sync.Once
	closeCount           atomic.Int32
	writeCount           atomic.Int32
}

func newFakeWebsocketConnection() *fakeWebsocketConnection {
	return &fakeWebsocketConnection{
		readStarted:  make(chan struct{}),
		writeStarted: make(chan struct{}),
		closed:       make(chan struct{}),
	}
}

func (f *fakeWebsocketConnection) Close() error {
	f.closeCount.Add(1)
	f.closeOnce.Do(func() {
		if f.closeHook != nil {
			f.closeHook()
		}
		close(f.closed)
	})
	return nil
}

func (f *fakeWebsocketConnection) SetReadLimit(int64) {}

func (f *fakeWebsocketConnection) SetReadDeadline(time.Time) error { return nil }

func (f *fakeWebsocketConnection) ReadMessage() (int, []byte, error) {
	f.readStartOnce.Do(func() {
		close(f.readStarted)
	})
	if f.readRelease != nil {
		select {
		case <-f.readRelease:
		case <-f.closed:
			f.waitForReadCloseRelease()
			return 0, nil, errors.New("socket closed")
		}
	}
	if f.readResults != nil {
		select {
		case result := <-f.readResults:
			return gorillaws.TextMessage, result.message, result.err
		case <-f.closed:
			f.waitForReadCloseRelease()
			return 0, nil, errors.New("socket closed")
		}
	}
	if f.readErr != nil {
		return 0, nil, f.readErr
	}
	return 0, nil, errors.New("not implemented")
}

func (f *fakeWebsocketConnection) waitForReadCloseRelease() {
	if f.readCloseObserved != nil {
		f.readCloseObserveOnce.Do(func() {
			close(f.readCloseObserved)
		})
	}
	if f.readCloseRelease != nil {
		<-f.readCloseRelease
	}
}

func (f *fakeWebsocketConnection) SetWriteDeadline(deadline time.Time) error {
	f.writeDeadlines = append(f.writeDeadlines, deadline)
	if deadline.IsZero() {
		return f.clearDeadlineErr
	}
	return f.initialDeadlineErr
}

func (f *fakeWebsocketConnection) WriteMessage(int, []byte) error {
	f.writeCount.Add(1)
	f.writeStartOnce.Do(func() {
		close(f.writeStarted)
	})
	if f.writeRelease != nil {
		select {
		case <-f.writeRelease:
		case <-f.closed:
			return errors.New("socket closed")
		}
	}
	if f.writeHook != nil {
		f.writeHook()
	}
	return f.writeErr
}

func TestConnection_DisconnectStopsConcurrentWriteMessage(t *testing.T) {
	ws := &testutil.MockWebSocketServer{}
	server := ws.TestWebSocketServer(func(serverConn *gorillaws.Conn) {
		defer serverConn.Close()
		for {
			if _, _, err := serverConn.ReadMessage(); err != nil {
				return
			}
		}
	})
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)

	conn := NewConnection(url)
	require.NoError(t, conn.Connect())

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for {
			if err := conn.WriteMessage([]byte("ping")); err != nil {
				return
			}
		}
	}()

	// Give the goroutine time to issue several writes before disconnecting.
	time.Sleep(50 * time.Millisecond)
	require.NoError(t, conn.Disconnect())

	select {
	case <-writerDone:
	case <-time.After(time.Second):
		t.Fatal("WriteMessage goroutine did not exit after Disconnect, possible goroutine leak")
	}
}
