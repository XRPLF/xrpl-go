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
// Run with -race to expose a missing mutex; a lucky-scheduled run can pass without it.
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

type fakeWebsocketConnection struct {
	initialDeadlineErr error
	clearDeadlineErr   error
	writeErr           error
	writeStarted       chan struct{}
	writeRelease       chan struct{}
	closed             chan struct{}
	closeHook          func()
	closeOnce          sync.Once
	startOnce          sync.Once
	closeCount         atomic.Int32
	writeCount         atomic.Int32
}

func newFakeWebsocketConnection() *fakeWebsocketConnection {
	return &fakeWebsocketConnection{
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
	return 0, nil, errors.New("not implemented")
}

func (f *fakeWebsocketConnection) SetWriteDeadline(deadline time.Time) error {
	if deadline.IsZero() {
		return f.clearDeadlineErr
	}
	return f.initialDeadlineErr
}

func (f *fakeWebsocketConnection) WriteMessage(int, []byte) error {
	f.writeCount.Add(1)
	f.startOnce.Do(func() {
		close(f.writeStarted)
	})
	if f.writeRelease != nil {
		select {
		case <-f.writeRelease:
		case <-f.closed:
			return errors.New("socket closed")
		}
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
