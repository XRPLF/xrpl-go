package websocket

import (
	"sync"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestConnection_ReadMessageSerializesConcurrentReaders(t *testing.T) {
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
