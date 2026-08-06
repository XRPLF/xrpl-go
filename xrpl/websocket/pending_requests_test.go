package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestClientDisconnectRejectsPendingRequest(t *testing.T) {
	requestReceived := make(chan struct{})
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
		close(requestReceived)
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	client := newPendingRequestTestClient(t, server.URL, 0)
	require.NoError(t, client.Connect())

	requestErr := make(chan error, 1)
	go func() {
		_, err := client.Request(newAccountChannelsRequest())
		requestErr <- err
	}()

	select {
	case <-requestReceived:
	case <-time.After(time.Second):
		t.Fatal("server did not receive request")
	}

	started := time.Now()
	require.NoError(t, client.Disconnect())
	select {
	case err := <-requestErr:
		require.ErrorIs(t, err, ErrDisconnected)
		require.Less(t, time.Since(started), time.Second)
	case <-time.After(time.Second):
		t.Fatal("pending request was not rejected")
	}
	require.Zero(t, pendingResponseCount(client))
}

func TestClientUnexpectedDisconnectRejectsAllPendingRequests(t *testing.T) {
	const requestCount = 8

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for range requestCount {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	client := newPendingRequestTestClient(t, server.URL, 0)
	require.NoError(t, client.Connect())
	defer client.Disconnect()

	var waitGroup sync.WaitGroup
	errorsByRequest := make(chan error, requestCount)
	waitGroup.Add(requestCount)
	for range requestCount {
		go func() {
			defer waitGroup.Done()
			_, err := client.Request(newAccountChannelsRequest())
			errorsByRequest <- err
		}()
	}
	waitGroup.Wait()
	close(errorsByRequest)

	for err := range errorsByRequest {
		require.ErrorIs(t, err, ErrDisconnected)
	}
	require.Zero(t, pendingResponseCount(client))
}

func TestClientRequestAfterReconnectUsesNewID(t *testing.T) {
	var connectionCount atomic.Int32
	requestIDs := make(chan uint64, 2)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionNumber := connectionCount.Add(1)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var request struct {
			ID uint64 `json:"id"`
		}
		if err := json.Unmarshal(message, &request); err != nil {
			return
		}
		requestIDs <- request.ID
		if connectionNumber == 1 {
			return
		}
		_ = conn.WriteJSON(map[string]any{
			"id": request.ID, "status": "success", "type": "response", "result": map[string]any{},
		})
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	client := newPendingRequestTestClient(t, server.URL, 2)
	client.cfg.reconnectBaseDelay = time.Millisecond
	client.cfg.reconnectMaxDelay = time.Millisecond
	require.NoError(t, client.Connect())
	defer client.Disconnect()

	_, err := client.Request(newAccountChannelsRequest())
	require.ErrorIs(t, err, ErrDisconnected)
	require.Eventually(t, client.IsConnected, time.Second, time.Millisecond)

	response, err := client.Request(newAccountChannelsRequest())
	require.NoError(t, err)
	require.NotNil(t, response)

	firstID := <-requestIDs
	secondID := <-requestIDs
	require.Greater(t, secondID, firstID)
	require.Zero(t, pendingResponseCount(client))
}

func TestClientWriteFailureReturnsDisconnected(t *testing.T) {
	client := NewClient(NewClientConfig().WithTimeout(time.Second))
	socket := newFakeWebsocketConnection()
	socket.writeErr = errors.New("write failed")
	client.conn.conn = socket

	response, err := client.Request(newAccountChannelsRequest())
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrDisconnected)
	require.Zero(t, pendingResponseCount(client))
}

func TestClientRequestTimeoutCoversWriteAndResponse(t *testing.T) {
	const requestTimeout = 100 * time.Millisecond
	client := NewClient(NewClientConfig().WithTimeout(requestTimeout))
	socket := newFakeWebsocketConnection()
	socket.writeRelease = make(chan struct{})
	client.conn.conn = socket

	go func() {
		time.Sleep(75 * time.Millisecond)
		close(socket.writeRelease)
	}()

	started := time.Now()
	response, err := client.Request(newAccountChannelsRequest())
	duration := time.Since(started)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrRequestTimedOut)
	require.Less(t, duration, 150*time.Millisecond)
	require.Zero(t, pendingResponseCount(client))
}

func TestClientDisconnectWithoutSocketIsIdempotent(t *testing.T) {
	client := NewClient(*NewClientConfig())
	require.NoError(t, client.Disconnect())
	require.NoError(t, client.Disconnect())
}

func TestLateResponseDoesNotReplaceDisconnectError(t *testing.T) {
	client := NewClient(*NewClientConfig())
	responseChan := client.registerPendingResponse(7)
	client.failPendingResponses(ErrDisconnected)
	client.handleRequest(context.Background(), []byte(`{"id":7,"type":"response","status":"success","result":{}}`))

	response, err := client.awaitResponse(context.Background(), responseChan)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrDisconnected)
	require.Zero(t, pendingResponseCount(client))
}

func TestPendingResponseContextAndDisconnectRace(t *testing.T) {
	client := NewClient(NewClientConfig().WithTimeout(time.Second))

	for id := uint64(1); id <= 100; id++ {
		responseChan := client.registerPendingResponse(id)
		ctx, cancel := context.WithCancel(context.Background())
		start := make(chan struct{})
		var waitGroup sync.WaitGroup
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			<-start
			cancel()
		}()
		go func() {
			defer waitGroup.Done()
			<-start
			client.failPendingResponses(ErrDisconnected)
		}()
		close(start)

		_, err := client.awaitResponse(ctx, responseChan)
		require.True(t, errors.Is(err, context.Canceled) || errors.Is(err, ErrDisconnected), "unexpected error: %v", err)
		waitGroup.Wait()
		client.unregisterPendingResponse(id)
	}

	require.Zero(t, pendingResponseCount(client))
}

func newPendingRequestTestClient(t *testing.T, serverURL string, maxReconnects int) *Client {
	t.Helper()
	url, err := testutil.ConvertHTTPToWS(serverURL)
	require.NoError(t, err)
	config := NewClientConfig().
		WithHost(url).
		WithTimeout(5*time.Second).
		WithMaxReconnects(maxReconnects).
		WithNetworkIdentity(0, "1.12.0")
	return NewClient(config)
}

func pendingResponseCount(client *Client) int {
	client.pendingResponsesMu.Lock()
	defer client.pendingResponsesMu.Unlock()
	return len(client.pendingResponses)
}
