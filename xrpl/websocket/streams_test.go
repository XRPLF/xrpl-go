package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	streamtypes "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types"
	"github.com/stretchr/testify/require"
)

func TestClient_ReportError(t *testing.T) {
	tests := []struct {
		name   string
		client func() *Client
	}{
		{
			name: "does not block without handler",
			client: func() *Client {
				return NewClient(*NewClientConfig())
			},
		},
		{
			name: "does not block with zero-value client",
			client: func() *Client {
				return &Client{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := tt.client()
			expectedErr := errors.New("test error")

			done := make(chan struct{})
			go func() {
				cl.reportError(cl.lifecycleContext(), expectedErr)
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("reportError blocked")
			}
		})
	}
}

func TestClient_OnErrorHandlesReportedErrors(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	cl.resetLifecycle()
	defer cl.cancelLifecycle()

	expectedErr := errors.New("test error")
	receivedErr := make(chan error, 1)

	cl.OnError(func(err error) {
		receivedErr <- err
	})

	cl.reportError(cl.lifecycleContext(), expectedErr)

	select {
	case err := <-receivedErr:
		require.Equal(t, expectedErr, err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for error handler")
	}
}

func TestClient_OnErrorReplacesPreviousHandler(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	cl.resetLifecycle()
	defer cl.cancelLifecycle()

	firstCalled := make(chan struct{}, 1)
	secondCalled := make(chan struct{}, 1)

	cl.OnError(func(error) {
		firstCalled <- struct{}{}
	})
	cl.OnError(func(error) {
		secondCalled <- struct{}{}
	})

	cl.reportError(cl.lifecycleContext(), errors.New("test error"))

	select {
	case <-firstCalled:
		t.Fatal("first error handler was called after replacement")
	default:
	}

	select {
	case <-secondCalled:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for replacement error handler")
	}
}

func TestClient_OnErrorCanBeDisabledWithNil(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	called := make(chan struct{}, 1)

	cl.OnError(func(error) {
		called <- struct{}{}
	})
	cl.OnError(nil)

	cl.reportError(cl.lifecycleContext(), errors.New("test error"))

	select {
	case <-called:
		t.Fatal("error handler was called after being disabled")
	default:
	}
}

func TestClient_StreamHandlersReceiveReportedStreams(t *testing.T) {
	tests := []struct {
		name     string
		register func(*Client, chan struct{})
		report   func(*Client)
	}{
		{
			name: "ledgerClosed",
			register: func(c *Client, received chan struct{}) {
				c.OnLedgerClosed(func(*streamtypes.LedgerStream) {
					received <- struct{}{}
				})
			},
			report: func(c *Client) {
				c.reportLedgerClosed(c.lifecycleContext(), &streamtypes.LedgerStream{})
			},
		},
		{
			name: "validationReceived",
			register: func(c *Client, received chan struct{}) {
				c.OnValidationReceived(func(*streamtypes.ValidationStream) {
					received <- struct{}{}
				})
			},
			report: func(c *Client) {
				c.reportValidationReceived(c.lifecycleContext(), &streamtypes.ValidationStream{})
			},
		},
		{
			name: "transaction",
			register: func(c *Client, received chan struct{}) {
				c.OnTransactions(func(*streamtypes.TransactionStream) {
					received <- struct{}{}
				})
			},
			report: func(c *Client) {
				c.reportTransaction(c.lifecycleContext(), &streamtypes.TransactionStream{})
			},
		},
		{
			name: "peerStatusChange",
			register: func(c *Client, received chan struct{}) {
				c.OnPeerStatusChange(func(*streamtypes.PeerStatusStream) {
					received <- struct{}{}
				})
			},
			report: func(c *Client) {
				c.reportPeerStatusChange(c.lifecycleContext(), &streamtypes.PeerStatusStream{})
			},
		},
		{
			// TODO: handleStream has no OrderBookStreamType case (it aliases
			// TransactionStreamType in xrpl/queries/subscription/types), so this
			// case exercises the handler runner via reportOrderBook directly and
			// does not cover wire dispatch.
			name: "orderBook",
			register: func(c *Client, received chan struct{}) {
				c.OnOrderBook(func(*streamtypes.OrderBookStream) {
					received <- struct{}{}
				})
			},
			report: func(c *Client) {
				c.reportOrderBook(c.lifecycleContext(), &streamtypes.OrderBookStream{})
			},
		},
		{
			// TODO: handleStream has no BookChangesStreamType case (the type is
			// not defined in xrpl/queries/subscription/types), so this case
			// exercises the handler runner via reportBookChanges directly and
			// does not cover wire dispatch.
			name: "bookChanges",
			register: func(c *Client, received chan struct{}) {
				c.OnBookChanges(func(*streamtypes.BookChangesStream) {
					received <- struct{}{}
				})
			},
			report: func(c *Client) {
				c.reportBookChanges(c.lifecycleContext(), &streamtypes.BookChangesStream{})
			},
		},
		{
			name: "consensusPhase",
			register: func(c *Client, received chan struct{}) {
				c.OnConsensusPhase(func(*streamtypes.ConsensusStream) {
					received <- struct{}{}
				})
			},
			report: func(c *Client) {
				c.reportConsensusPhase(c.lifecycleContext(), &streamtypes.ConsensusStream{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewClient(*NewClientConfig())
			cl.resetLifecycle()
			defer cl.cancelLifecycle()

			received := make(chan struct{}, 1)

			tt.register(cl, received)
			tt.report(cl)

			select {
			case <-received:
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for stream handler")
			}
		})
	}
}

func TestClient_StreamHandlersReplacePreviousHandler(t *testing.T) {
	tests := []struct {
		name     string
		register func(*Client, func())
		report   func(*Client)
	}{
		{
			name: "ledgerClosed",
			register: func(c *Client, handler func()) {
				c.OnLedgerClosed(func(*streamtypes.LedgerStream) {
					handler()
				})
			},
			report: func(c *Client) {
				c.reportLedgerClosed(c.lifecycleContext(), &streamtypes.LedgerStream{})
			},
		},
		{
			name: "validationReceived",
			register: func(c *Client, handler func()) {
				c.OnValidationReceived(func(*streamtypes.ValidationStream) {
					handler()
				})
			},
			report: func(c *Client) {
				c.reportValidationReceived(c.lifecycleContext(), &streamtypes.ValidationStream{})
			},
		},
		{
			name: "transaction",
			register: func(c *Client, handler func()) {
				c.OnTransactions(func(*streamtypes.TransactionStream) {
					handler()
				})
			},
			report: func(c *Client) {
				c.reportTransaction(c.lifecycleContext(), &streamtypes.TransactionStream{})
			},
		},
		{
			name: "peerStatusChange",
			register: func(c *Client, handler func()) {
				c.OnPeerStatusChange(func(*streamtypes.PeerStatusStream) {
					handler()
				})
			},
			report: func(c *Client) {
				c.reportPeerStatusChange(c.lifecycleContext(), &streamtypes.PeerStatusStream{})
			},
		},
		{
			// TODO: handleStream has no OrderBookStreamType case (it aliases
			// TransactionStreamType in xrpl/queries/subscription/types), so this
			// case exercises the handler runner via reportOrderBook directly and
			// does not cover wire dispatch.
			name: "orderBook",
			register: func(c *Client, handler func()) {
				c.OnOrderBook(func(*streamtypes.OrderBookStream) {
					handler()
				})
			},
			report: func(c *Client) {
				c.reportOrderBook(c.lifecycleContext(), &streamtypes.OrderBookStream{})
			},
		},
		{
			// TODO: handleStream has no BookChangesStreamType case (the type is
			// not defined in xrpl/queries/subscription/types), so this case
			// exercises the handler runner via reportBookChanges directly and
			// does not cover wire dispatch.
			name: "bookChanges",
			register: func(c *Client, handler func()) {
				c.OnBookChanges(func(*streamtypes.BookChangesStream) {
					handler()
				})
			},
			report: func(c *Client) {
				c.reportBookChanges(c.lifecycleContext(), &streamtypes.BookChangesStream{})
			},
		},
		{
			name: "consensusPhase",
			register: func(c *Client, handler func()) {
				c.OnConsensusPhase(func(*streamtypes.ConsensusStream) {
					handler()
				})
			},
			report: func(c *Client) {
				c.reportConsensusPhase(c.lifecycleContext(), &streamtypes.ConsensusStream{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewClient(*NewClientConfig())
			cl.resetLifecycle()
			defer cl.cancelLifecycle()

			firstCalled := make(chan struct{}, 1)
			secondCalled := make(chan struct{}, 1)

			tt.register(cl, func() {
				firstCalled <- struct{}{}
			})
			tt.register(cl, func() {
				secondCalled <- struct{}{}
			})

			tt.report(cl)

			select {
			case <-firstCalled:
				t.Fatal("first stream handler was called after replacement")
			default:
			}

			select {
			case <-secondCalled:
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for replacement stream handler")
			}
		})
	}
}

func TestClient_StreamHandlersCanBeDisabledWithNil(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	called := make(chan struct{}, 1)

	cl.OnLedgerClosed(func(*streamtypes.LedgerStream) {
		called <- struct{}{}
	})
	cl.OnLedgerClosed(nil)

	cl.reportLedgerClosed(cl.lifecycleContext(), &streamtypes.LedgerStream{})

	select {
	case <-called:
		t.Fatal("stream handler was called after being disabled")
	default:
	}
}

func TestClient_StreamHandlersSurviveLifecycleReset(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	received := make(chan struct{}, 1)

	cl.OnLedgerClosed(func(*streamtypes.LedgerStream) {
		received <- struct{}{}
	})

	cl.resetLifecycle()
	defer cl.cancelLifecycle()

	cl.reportLedgerClosed(cl.lifecycleContext(), &streamtypes.LedgerStream{})

	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler after lifecycle reset")
	}
}

func TestClient_ReportStreamSkipsWhenChannelUnset(t *testing.T) {
	tests := []struct {
		name   string
		report func(*Client)
	}{
		{
			name: "ledgerClosed",
			report: func(c *Client) {
				c.reportLedgerClosed(c.lifecycleContext(), &streamtypes.LedgerStream{})
			},
		},
		{
			name: "validationReceived",
			report: func(c *Client) {
				c.reportValidationReceived(c.lifecycleContext(), &streamtypes.ValidationStream{})
			},
		},
		{
			name: "transaction",
			report: func(c *Client) {
				c.reportTransaction(c.lifecycleContext(), &streamtypes.TransactionStream{})
			},
		},
		{
			name: "peerStatusChange",
			report: func(c *Client) {
				c.reportPeerStatusChange(c.lifecycleContext(), &streamtypes.PeerStatusStream{})
			},
		},
		{
			// TODO: handleStream has no OrderBookStreamType case (it aliases
			// TransactionStreamType in xrpl/queries/subscription/types), so this
			// case exercises the handler runner via reportOrderBook directly and
			// does not cover wire dispatch.
			name: "orderBook",
			report: func(c *Client) {
				c.reportOrderBook(c.lifecycleContext(), &streamtypes.OrderBookStream{})
			},
		},
		{
			// TODO: handleStream has no BookChangesStreamType case (the type is
			// not defined in xrpl/queries/subscription/types), so this case
			// exercises the handler runner via reportBookChanges directly and
			// does not cover wire dispatch.
			name: "bookChanges",
			report: func(c *Client) {
				c.reportBookChanges(c.lifecycleContext(), &streamtypes.BookChangesStream{})
			},
		},
		{
			name: "consensusPhase",
			report: func(c *Client) {
				c.reportConsensusPhase(c.lifecycleContext(), &streamtypes.ConsensusStream{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewClient(*NewClientConfig())
			done := make(chan struct{})

			go func() {
				tt.report(cl)
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("stream report blocked without channel")
			}
		})
	}
}

func TestClient_DisconnectCancelsLifecycle(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	ctx := cl.lifecycleContext()

	err := cl.Disconnect()
	require.ErrorIs(t, err, ErrNotConnected)

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for lifecycle cancellation")
	}
}

func TestClient_ResetLifecycleCreatesFreshContext(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	firstCtx := cl.lifecycleContext()

	cl.cancelLifecycle()
	secondCtx := cl.resetLifecycle()

	select {
	case <-firstCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first lifecycle cancellation")
	}

	select {
	case <-secondCtx.Done():
		t.Fatal("new lifecycle context was already canceled")
	default:
	}

	cl.cancelLifecycle()
	select {
	case <-secondCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for second lifecycle cancellation")
	}
}

func TestClient_ResetLifecycleSerializesStreamHandlerState(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	firstCtx := cl.lifecycleContext()

	cl.errorStream.stateMu.Lock()
	streamLocked := true
	defer func() {
		if streamLocked {
			cl.errorStream.stateMu.Unlock()
		}
	}()

	resetDone := make(chan struct{})
	go func() {
		cl.resetLifecycle()
		cl.cancelLifecycle()
		close(resetDone)
	}()

	select {
	case <-firstCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first lifecycle cancellation")
	}

	requireStreamHandlerStateMutexHeld(t, cl, 50*time.Millisecond)

	cl.errorStream.stateMu.Unlock()
	streamLocked = false

	select {
	case <-resetDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for lifecycle reset")
	}
}

func TestClient_CancelLifecycleSerializesStreamHandlerState(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	firstCtx := cl.lifecycleContext()

	cl.errorStream.stateMu.Lock()
	streamLocked := true
	defer func() {
		if streamLocked {
			cl.errorStream.stateMu.Unlock()
		}
	}()

	cancelDone := make(chan struct{})
	go func() {
		cl.cancelLifecycle()
		close(cancelDone)
	}()

	select {
	case <-firstCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for lifecycle cancellation")
	}

	requireStreamHandlerStateMutexHeld(t, cl, 50*time.Millisecond)

	cl.errorStream.stateMu.Unlock()
	streamLocked = false

	select {
	case <-cancelDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for lifecycle cancellation")
	}
}

func TestClient_HandlerRegistrationSerializesStreamHandlerState(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	defer cl.cancelLifecycle()

	cl.ledgerClosedStream.stateMu.Lock()
	streamLocked := true
	defer func() {
		if streamLocked {
			cl.ledgerClosedStream.stateMu.Unlock()
		}
	}()

	registerDone := make(chan struct{})
	go func() {
		cl.OnLedgerClosed(func(*streamtypes.LedgerStream) {})
		close(registerDone)
	}()

	waitForStreamHandlerStateMutexHeld(t, cl, time.Second)
	requireStreamHandlerStateMutexHeld(t, cl, 50*time.Millisecond)

	cl.ledgerClosedStream.stateMu.Unlock()
	streamLocked = false

	select {
	case <-registerDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for handler registration")
	}
}

func TestClient_HandleMessageUsesReaderLifecycle(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	defer cl.cancelLifecycle()

	received := make(chan struct{}, 1)
	cl.OnLedgerClosed(func(*streamtypes.LedgerStream) {
		received <- struct{}{}
	})
	oldCtx := cl.lifecycleContext()

	newCtx := cl.resetLifecycle()

	select {
	case <-oldCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for old lifecycle cancellation")
	}

	ledgerMessage := []byte(`{"type":"ledgerClosed","ledger_index":1}`)
	cl.handleMessage(oldCtx, ledgerMessage)

	select {
	case <-received:
		t.Fatal("stale lifecycle message was delivered to fresh handler")
	case <-time.After(50 * time.Millisecond):
	}

	cl.handleMessage(newCtx, ledgerMessage)

	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for fresh lifecycle message")
	}
}

func TestClient_ResetLifecycleWaitsForOldStreamRunner(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	cl.resetLifecycle()
	defer cl.cancelLifecycle()

	handlerStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	handlerDone := make(chan struct{})
	cl.OnLedgerClosed(func(*streamtypes.LedgerStream) {
		close(handlerStarted)
		<-releaseHandler
		close(handlerDone)
	})

	cl.reportLedgerClosed(cl.lifecycleContext(), &streamtypes.LedgerStream{})
	select {
	case <-handlerStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for old stream handler")
	}

	resetDone := make(chan struct{})
	go func() {
		cl.resetLifecycle()
		cl.cancelLifecycle()
		close(resetDone)
	}()

	select {
	case <-resetDone:
		t.Fatal("lifecycle reset completed before old stream runner exited")
	case <-time.After(50 * time.Millisecond):
	}

	close(releaseHandler)

	select {
	case <-handlerDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for old stream handler to finish")
	}

	select {
	case <-resetDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for lifecycle reset")
	}
}

func TestClient_DisconnectFromStreamHandlerDoesNotDeadlock(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	cl.resetLifecycle()

	disconnected := make(chan struct{})
	cl.OnLedgerClosed(func(*streamtypes.LedgerStream) {
		_ = cl.Disconnect()
		close(disconnected)
	})

	cl.reportLedgerClosed(cl.lifecycleContext(), &streamtypes.LedgerStream{})

	select {
	case <-disconnected:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for disconnect from stream handler")
	}
}

func TestLifecycleStreamReportUsesHandlerSnapshot(t *testing.T) {
	ctx := t.Context()

	var stream lifecycleStream[int]
	firstCalled := make(chan int, 1)
	secondCalled := make(chan int, 1)
	firstHandler := func(value int) {
		firstCalled <- value
	}
	stream.handler = firstHandler
	stream.handlerCh = make(chan lifecycleEvent[int])

	ch := stream.handlerCh
	reported := make(chan struct{})
	go func() {
		stream.Report(ctx, 1)
		close(reported)
	}()

	select {
	case <-reported:
		t.Fatal("report completed before the event was received")
	case <-time.After(50 * time.Millisecond):
	}

	stream.stateMu.Lock()
	stream.handler = func(value int) {
		secondCalled <- value
	}
	stream.stateMu.Unlock()

	event := <-ch
	event.handler(event.value)

	select {
	case <-reported:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for report to complete")
	}

	select {
	case value := <-firstCalled:
		require.Equal(t, 1, value)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for original handler")
	}

	select {
	case <-secondCalled:
		t.Fatal("replacement handler received in-flight event")
	default:
	}
}

func waitForStreamHandlerStateMutexHeld(t *testing.T, cl *Client, timeout time.Duration) {
	t.Helper()

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline.C:
			t.Fatal("timed out waiting for stream handler state mutex to be held")
		case <-ticker.C:
			if cl.streamHandlerStateMu.TryLock() {
				cl.streamHandlerStateMu.Unlock()
				continue
			}
			return
		}
	}
}

func requireStreamHandlerStateMutexHeld(t *testing.T, cl *Client, duration time.Duration) {
	t.Helper()

	deadline := time.NewTimer(duration)
	defer deadline.Stop()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline.C:
			return
		case <-ticker.C:
			if cl.streamHandlerStateMu.TryLock() {
				cl.streamHandlerStateMu.Unlock()
				t.Fatal("stream handler state mutex was released while a lifecycle transition was blocked")
			}
		}
	}
}

func TestLifecycleStreamStartRunsRegisteredHandler(t *testing.T) {
	ctx := t.Context()

	var stream lifecycleStream[int]
	received := make(chan int, 1)
	stream.handler = func(value int) {
		received <- value
	}

	stream.Start(ctx)
	stream.Report(ctx, 1)

	select {
	case value := <-received:
		require.Equal(t, 1, value)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for lifecycle stream handler")
	}
}

func TestLifecycleStreamStartSkipsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var stream lifecycleStream[int]
	stream.handler = func(int) {}

	stream.Start(ctx)

	require.False(t, stream.running)
	require.Nil(t, stream.handlerCh)
}

func TestClient_ReportStreamAfterDisconnectDoesNotBlock(t *testing.T) {
	tests := []struct {
		name     string
		register func(*Client)
		report   func(*Client)
	}{
		{
			name: "ledgerClosed",
			register: func(c *Client) {
				c.OnLedgerClosed(func(*streamtypes.LedgerStream) {})
			},
			report: func(c *Client) {
				c.reportLedgerClosed(c.lifecycleContext(), &streamtypes.LedgerStream{})
			},
		},
		{
			name: "validationReceived",
			register: func(c *Client) {
				c.OnValidationReceived(func(*streamtypes.ValidationStream) {})
			},
			report: func(c *Client) {
				c.reportValidationReceived(c.lifecycleContext(), &streamtypes.ValidationStream{})
			},
		},
		{
			name: "transaction",
			register: func(c *Client) {
				c.OnTransactions(func(*streamtypes.TransactionStream) {})
			},
			report: func(c *Client) {
				c.reportTransaction(c.lifecycleContext(), &streamtypes.TransactionStream{})
			},
		},
		{
			name: "peerStatusChange",
			register: func(c *Client) {
				c.OnPeerStatusChange(func(*streamtypes.PeerStatusStream) {})
			},
			report: func(c *Client) {
				c.reportPeerStatusChange(c.lifecycleContext(), &streamtypes.PeerStatusStream{})
			},
		},
		{
			// TODO: handleStream has no OrderBookStreamType case (it aliases
			// TransactionStreamType in xrpl/queries/subscription/types), so this
			// case exercises the handler runner via reportOrderBook directly and
			// does not cover wire dispatch.
			name: "orderBook",
			register: func(c *Client) {
				c.OnOrderBook(func(*streamtypes.OrderBookStream) {})
			},
			report: func(c *Client) {
				c.reportOrderBook(c.lifecycleContext(), &streamtypes.OrderBookStream{})
			},
		},
		{
			// TODO: handleStream has no BookChangesStreamType case (the type is
			// not defined in xrpl/queries/subscription/types), so this case
			// exercises the handler runner via reportBookChanges directly and
			// does not cover wire dispatch.
			name: "bookChanges",
			register: func(c *Client) {
				c.OnBookChanges(func(*streamtypes.BookChangesStream) {})
			},
			report: func(c *Client) {
				c.reportBookChanges(c.lifecycleContext(), &streamtypes.BookChangesStream{})
			},
		},
		{
			name: "consensusPhase",
			register: func(c *Client) {
				c.OnConsensusPhase(func(*streamtypes.ConsensusStream) {})
			},
			report: func(c *Client) {
				c.reportConsensusPhase(c.lifecycleContext(), &streamtypes.ConsensusStream{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewClient(*NewClientConfig())
			tt.register(cl)
			_ = cl.Disconnect()

			done := make(chan struct{})
			go func() {
				tt.report(cl)
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("stream report blocked after disconnect")
			}
		})
	}
}
