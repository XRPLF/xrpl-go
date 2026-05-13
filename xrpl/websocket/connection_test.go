package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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
			message:         strings.Repeat("a", int(defaultMaxResponseSize)+1),
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
func TestConnection_ReadMessageAllowsConcurrentCallers(t *testing.T) {
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
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			message, err := conn.ReadMessage()
			results <- readResult{
				message: message,
				err:     err,
			}
		}()
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
