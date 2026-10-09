package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/stretchr/testify/require"
)

func TestRequestContextAlreadyCanceled(t *testing.T) {
	c := NewClient(*NewClientConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.RequestContext(ctx, &server.InfoRequest{})
	require.ErrorIs(t, err, context.Canceled)
	require.Zero(t, pendingResponseCount(c))
}

func TestRequestContextCanceledActiveWriteKeepsSocket(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name := "cancel"
		if deadline {
			name = "deadline"
		}
		t.Run(name, func(t *testing.T) {
			c := NewClient(NewClientConfig().WithTimeout(3 * time.Second))
			socket := newFakeWebsocketConnection()
			socket.writeRelease = make(chan struct{})
			c.conn.conn = socket
			ctx, cancel := context.WithCancel(context.Background())
			expected := context.Canceled
			if deadline {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
				expected = context.DeadlineExceeded
			}
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := c.RequestContext(ctx, &server.InfoRequest{}); done <- err }()
			select {
			case <-socket.writeStarted:
			case <-time.After(time.Second):
				t.Fatal("write did not start")
			}
			if !deadline {
				cancel()
			}
			select {
			case err := <-done:
				require.ErrorIs(t, err, expected)
			case <-time.After(time.Second):
				t.Fatal("caller did not return")
			}
			require.Zero(t, pendingResponseCount(c))
			require.Zero(t, socket.closeCount.Load())
			require.True(t, c.IsConnected())
			// A queued canceled request must neither start a write nor close A's socket.
			queued, cancelQueued := context.WithCancel(context.Background())
			cancelQueued()
			_, err := c.RequestContext(queued, &server.InfoRequest{})
			require.ErrorIs(t, err, context.Canceled)
			require.Equal(t, int32(1), socket.writeCount.Load())
			close(socket.writeRelease)
			// Taking the token proves A's writer (including deadline cleanup) finished.
			wait, stop := context.WithTimeout(context.Background(), time.Second)
			defer stop()
			require.NoError(t, c.conn.acquireWrite(wait))
			c.conn.releaseWrite()
			require.Zero(t, socket.closeCount.Load())
			c.handleRequest(context.Background(), []byte(`{"id":1,"type":"response","status":"success","result":{}}`))
			require.Zero(t, pendingResponseCount(c))
			// A subsequent request receives its own response on the same fake socket.
			socket.writeHook = func() {
				c.handleRequest(context.Background(), []byte(`{"id":2,"type":"response","status":"success","result":{}}`))
			}
			res, err := c.RequestContext(context.Background(), &server.InfoRequest{})
			require.NoError(t, err)
			require.Equal(t, uint64(2), res.ID)
			require.Zero(t, pendingResponseCount(c))
			require.Zero(t, socket.closeCount.Load())
		})
	}
}

func TestRequestContextDetachedWriteStillTimesOut(t *testing.T) {
	c := NewClient(NewClientConfig().WithTimeout(100 * time.Millisecond))
	socket := newFakeWebsocketConnection()
	socket.writeRelease = make(chan struct{})
	c.conn.conn = socket
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := c.RequestContext(ctx, &server.InfoRequest{}); done <- err }()
	<-socket.writeStarted
	cancel()
	require.ErrorIs(t, <-done, context.Canceled)
	select {
	case <-socket.closed:
	case <-time.After(time.Second):
		t.Fatal("configured timeout did not stop detached writer")
	}
	wait, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	require.NoError(t, c.conn.acquireWrite(wait))
	c.conn.releaseWrite()
	require.Zero(t, pendingResponseCount(c))
	require.False(t, c.IsConnected())
}

func TestRequestContextCleanupOnResponseCancellation(t *testing.T) {
	for range 50 {
		c := NewClient(NewClientConfig().WithTimeout(time.Second))
		socket := newFakeWebsocketConnection()
		c.conn.conn = socket
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { _, err := c.RequestContext(ctx, &server.InfoRequest{}); done <- err }()
		<-socket.writeStarted
		go c.handleRequest(context.Background(), []byte(`{"id":1,"type":"response","status":"success","result":{}}`))
		cancel()
		err := <-done
		require.True(t, err == nil || errors.Is(err, context.Canceled), "unexpected error %v", err)
		require.Zero(t, pendingResponseCount(c))
		require.Zero(t, socket.closeCount.Load())
	}
}

func TestRequestContextCanceledWhileQueuedDoesNotWrite(t *testing.T) {
	c := NewClient(NewClientConfig().WithTimeout(time.Second))
	socket := newFakeWebsocketConnection()
	c.conn.conn = socket
	require.NoError(t, c.conn.acquireWrite(context.Background()))
	defer c.conn.releaseWrite()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := c.RequestContext(ctx, &server.InfoRequest{}); done <- err }()
	require.Eventually(t, func() bool { return pendingResponseCount(c) == 1 }, time.Second, time.Millisecond)
	cancel()
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("queued request did not cancel")
	}
	require.Zero(t, pendingResponseCount(c))
	require.Zero(t, socket.writeCount.Load())
	require.Zero(t, socket.closeCount.Load())
}

func TestRequestContextConfiguredTimeoutDuringWrite(t *testing.T) {
	c := NewClient(NewClientConfig().WithTimeout(50 * time.Millisecond))
	socket := newFakeWebsocketConnection()
	socket.writeRelease = make(chan struct{})
	c.conn.conn = socket
	_, err := c.RequestContext(context.Background(), &server.InfoRequest{})
	require.ErrorIs(t, err, ErrRequestTimedOut)
	require.Zero(t, pendingResponseCount(c))
	select {
	case <-socket.closed:
	case <-time.After(time.Second):
		t.Fatal("configured write timeout did not close failed socket")
	}
}
