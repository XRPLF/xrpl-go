package websocket

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/stretchr/testify/require"
)

func TestReviewRequestContextCustomCause(t *testing.T) {
	c := NewClient(NewClientConfig().WithTimeout(time.Second))
	socket := newFakeWebsocketConnection()
	c.conn.conn = socket
	require.NoError(t, c.conn.acquireWrite(context.Background()))
	defer c.conn.releaseWrite()
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	done := make(chan error, 1)
	go func() { _, err := c.RequestContext(ctx, &server.InfoRequest{}); done <- err }()
	require.Eventually(t, func() bool { return pendingResponseCount(c) == 1 }, time.Second, time.Millisecond)
	reason := errors.New("caller custom reason")
	cancel(reason)
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
		require.ErrorIs(t, err, reason)
	case <-time.After(time.Second):
		t.Fatal("request did not return")
	}
}

func TestReviewCustomCauseDuringRequest(t *testing.T) {
	for _, stage := range []string{"queued", "response"} {
		for _, kind := range []string{"cancel", "deadline"} {
			t.Run(stage+"/"+kind, func(t *testing.T) {
				c := NewClient(NewClientConfig().WithTimeout(2 * time.Second))
				socket := newFakeWebsocketConnection()
				c.conn.conn = socket
				if stage == "queued" {
					require.NoError(t, c.conn.acquireWrite(context.Background()))
					defer c.conn.releaseWrite()
				}
				reason := errors.New("custom caller reason")
				var ctx context.Context
				var cancel context.CancelFunc
				expected := context.Canceled
				if kind == "deadline" {
					ctx, cancel = context.WithDeadlineCause(context.Background(), time.Now().Add(200*time.Millisecond), reason)
					expected = context.DeadlineExceeded
				} else {
					var cc context.CancelCauseFunc
					ctx, cc = context.WithCancelCause(context.Background())
					cancel = func() { cc(reason) }
				}
				defer cancel()
				done := make(chan error, 1)
				go func() { _, err := c.RequestContext(ctx, &server.InfoRequest{}); done <- err }()
				require.Eventually(t, func() bool { return pendingResponseCount(c) == 1 }, time.Second, time.Millisecond)
				if stage == "response" {
					<-socket.writeStarted
					wait, stop := context.WithTimeout(context.Background(), time.Second)
					defer stop()
					require.NoError(t, c.conn.acquireWrite(wait))
					c.conn.releaseWrite()
				}
				require.NoError(t, ctx.Err(), "test must reach stage before cancellation")
				if kind == "cancel" {
					cancel()
				}
				select {
				case err := <-done:
					require.ErrorIs(t, err, expected)
					require.ErrorIs(t, err, reason)
					require.NotErrorIs(t, err, ErrRequestTimedOut)
				case <-time.After(time.Second):
					t.Fatal("request did not return")
				}
				require.Zero(t, pendingResponseCount(c))
				require.Zero(t, socket.closeCount.Load())
			})
		}
	}
}

type reviewSignalRequest struct {
	server.InfoRequest
	validated chan struct{}
}

func (r *reviewSignalRequest) Validate() error { close(r.validated); return nil }

func TestReviewCancellationWaitingHandshake(t *testing.T) {
	c := NewClient(NewClientConfig().WithTimeout(time.Second))
	socket := newFakeWebsocketConnection()
	c.conn.conn = socket
	c.connectionHandshakeMu.Lock()
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	req := &reviewSignalRequest{validated: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := c.RequestContext(ctx, req); done <- err }()
	<-req.validated
	reason := errors.New("cancel during handshake")
	cancel(reason)
	c.connectionHandshakeMu.Unlock()
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
		require.ErrorIs(t, err, reason)
	case <-time.After(time.Second):
		t.Fatal("request did not return after unlock")
	}
	require.Zero(t, pendingResponseCount(c))
	require.Zero(t, socket.writeCount.Load())
	require.Zero(t, socket.closeCount.Load())
}

// Done is not ready in acquireWrite; Err cancels exactly at the post-acquire check.
// This controlled context makes the branch deterministic without a scheduler race.
type reviewPostAcquireContext struct {
	context.Context
	conn    *Connection
	checked bool
}

func (c *reviewPostAcquireContext) Err() error {
	c.checked = true
	if len(c.conn.writeToken) != 0 {
		panic("Err checked before token acquired")
	}
	return context.Canceled
}

func TestReviewPostAcquireCancellation(t *testing.T) {
	c := newConnection("ws://unused", defaultMaxResponseSize)
	socket := newFakeWebsocketConnection()
	c.conn = socket
	ctx := &reviewPostAcquireContext{Context: context.Background(), conn: c}
	err := c.writeRequestMessageTo(ctx, socket, []byte("ignored"), time.Now().Add(time.Second))
	require.ErrorIs(t, err, context.Canceled)
	require.True(t, ctx.checked)
	require.Zero(t, socket.writeCount.Load())
	require.Zero(t, socket.closeCount.Load())
	wait, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, c.acquireWrite(wait))
	c.releaseWrite()
}

func TestReviewConfiguredBudgetIncludesQueue(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewClient(NewClientConfig().WithTimeout(time.Second))
		socket := newFakeWebsocketConnection()
		socket.writeRelease = make(chan struct{})
		c.conn.conn = socket
		require.NoError(t, c.conn.acquireWrite(context.Background()))
		start := time.Now()
		done := make(chan error, 1)
		go func() { _, err := c.RequestContext(context.Background(), &server.InfoRequest{}); done <- err }()
		synctest.Wait()
		require.Equal(t, 1, pendingResponseCount(c))
		time.Sleep(600 * time.Millisecond)
		c.conn.releaseWrite()
		<-socket.writeStarted
		require.ErrorIs(t, <-done, ErrRequestTimedOut)
		require.Equal(t, start.Add(time.Second), time.Now())
		synctest.Wait()
		require.Equal(t, start.Add(time.Second), socket.writeDeadlines[0], "queue wait must not reset write deadline")
		require.Zero(t, pendingResponseCount(c))
		require.False(t, c.IsConnected())
	})
}

func TestReviewFirstCompletionCause(t *testing.T) {
	for _, callerFirst := range []bool{true, false} {
		name := "configured-first"
		if callerFirst {
			name = "caller-first"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				c := NewClient(NewClientConfig().WithTimeout(time.Second))
				socket := newFakeWebsocketConnection()
				c.conn.conn = socket
				require.NoError(t, c.conn.acquireWrite(context.Background()))
				defer c.conn.releaseWrite()
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				done := make(chan error, 1)
				go func() { _, err := c.RequestContext(ctx, &server.InfoRequest{}); done <- err }()
				synctest.Wait() // request context created; blocked on writer token
				reason := errors.New("caller reason")
				if callerFirst {
					cancel(reason)
				}
				time.Sleep(2 * time.Second)
				if !callerFirst {
					cancel(reason)
				}

				err := <-done
				if callerFirst {
					require.ErrorIs(t, err, context.Canceled)
					require.ErrorIs(t, err, reason)
					require.NotErrorIs(t, err, ErrRequestTimedOut)
				} else {
					require.ErrorIs(t, err, ErrRequestTimedOut)
					require.NotErrorIs(t, err, context.Canceled)
					require.NotErrorIs(t, err, reason)
				}
				require.Zero(t, pendingResponseCount(c))
				require.Zero(t, socket.writeCount.Load())
			})
		})
	}
}

func TestReviewDetachedWriteFailureAndLifecycle(t *testing.T) {
	c := NewClient(NewClientConfig().WithTimeout(time.Second))
	socket := newFakeWebsocketConnection()
	socket.writeRelease = make(chan struct{})
	socket.writeErr = errors.New("physical write failed")
	c.conn.conn = socket
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	// Exercise the private request path shared by typed/context-aware helpers too.
	go func() { _, err := c.request(ctx, &server.InfoRequest{}); done <- err }()
	<-socket.writeStarted
	cancel()
	require.ErrorIs(t, <-done, context.Canceled)
	close(socket.writeRelease)
	wait, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	require.NoError(t, c.conn.acquireWrite(wait))
	c.conn.releaseWrite()
	require.False(t, c.IsConnected())
	require.Positive(t, socket.closeCount.Load())
	require.Zero(t, pendingResponseCount(c))
	replacement := newFakeWebsocketConnection()
	c.conn.mu.Lock()
	c.conn.conn = replacement
	c.conn.mu.Unlock()
	replacement.writeHook = func() {
		c.handleRequest(context.Background(), []byte(`{"id":2,"type":"response","status":"success","result":{}}`))
	}
	_, err := c.request(context.Background(), &server.InfoRequest{})
	require.NoError(t, err)
	require.Zero(t, replacement.closeCount.Load())
	require.NoError(t, c.Disconnect())
	require.Equal(t, int32(1), replacement.closeCount.Load())
}

func TestReviewDisconnectAfterEarlyReturn(t *testing.T) {
	c := NewClient(NewClientConfig().WithTimeout(time.Second))
	socket := newFakeWebsocketConnection()
	socket.writeRelease = make(chan struct{})
	c.conn.conn = socket
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := c.request(ctx, &server.InfoRequest{}); done <- err }()
	<-socket.writeStarted
	cancel()
	require.ErrorIs(t, <-done, context.Canceled)
	require.NoError(t, c.Disconnect())
	wait, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	require.NoError(t, c.conn.acquireWrite(wait))
	c.conn.releaseWrite()
	require.Zero(t, pendingResponseCount(c))
	require.False(t, c.IsConnected())
}

func TestReviewWriteFailureSimultaneousCallerCancel(t *testing.T) {
	c := NewClient(NewClientConfig().WithTimeout(time.Second))
	socket := newFakeWebsocketConnection()
	c.conn.conn = socket
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	socket.writeHook = cancel
	socket.writeErr = errors.New("physical write failure concurrent with caller cancellation")
	_, err := c.RequestContext(ctx, &server.InfoRequest{})
	require.ErrorIs(t, err, context.Canceled)
	wait, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	require.NoError(t, c.conn.acquireWrite(wait))
	c.conn.releaseWrite()
	require.Positive(t, socket.closeCount.Load())
	require.False(t, c.IsConnected())
	require.Zero(t, pendingResponseCount(c))
}
