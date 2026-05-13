package websocket

import (
	"sync"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// Exercises the fix that serializes concurrent ReadMessage calls under readMu.
// Run with -race to expose a missing mutex; a lucky-scheduled run can pass without it.
func TestConnection_ReadMessageAllowsConcurrentCallers(t *testing.T) {
	readyToWrite := make(chan struct{})
	serverErr := make(chan error, 1)

	ws := &testutil.MockWebSocketServer{}
	server := ws.TestWebSocketServer(func(serverConn *websocket.Conn) {
		defer serverConn.Close()
		<-readyToWrite

		for _, msg := range []string{"first", "second"} {
			if err := serverConn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
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
	server := ws.TestWebSocketServer(func(serverConn *websocket.Conn) {
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
		t.Fatal("ReadMessage did not return after Disconnect — possible goroutine leak")
	}
}

func TestConnection_DisconnectStopsConcurrentWriteMessage(t *testing.T) {
	ws := &testutil.MockWebSocketServer{}
	server := ws.TestWebSocketServer(func(serverConn *websocket.Conn) {
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
		t.Fatal("WriteMessage goroutine did not exit after Disconnect — possible goroutine leak")
	}
}
