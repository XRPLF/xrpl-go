package websocket

import (
	"context"
	"sync"

	streamtypes "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types"
)

type lifecycleStream[T any] struct {
	// stateMu protects handlerCh, handler, running, and done.
	stateMu   sync.Mutex
	handlerCh chan lifecycleEvent[T]
	handler   func(T)
	running   bool
	done      chan struct{}
}

type lifecycleEvent[T any] struct {
	value   T
	handler func(T)
}

// On registers handler and starts delivery immediately when ctx is active.
// If ctx is canceled, the handler is still registered so a later Start can
// activate it for a fresh client lifecycle. On atomically replaces any
// previously registered handler on this stream rather than spawning an
// additional runner. Handler replacement is best-effort: an event already
// accepted for delivery may still be dispatched to the previously registered
// handler.
func (s *lifecycleStream[T]) On(ctx context.Context, handler func(T)) {
	s.stateMu.Lock()
	s.handler = handler
	if handler == nil || ctx == nil || ctx.Err() != nil {
		s.stateMu.Unlock()
		return
	}
	ch, done, shouldStart := s.startLocked()
	s.stateMu.Unlock()

	if shouldStart {
		go s.run(ctx, ch, done)
	}
}

func (s *lifecycleStream[T]) Start(ctx context.Context) {
	if ctx == nil || ctx.Err() != nil {
		return
	}

	s.stateMu.Lock()
	if s.handler == nil {
		s.stateMu.Unlock()
		return
	}
	ch, done, shouldStart := s.startLocked()
	s.stateMu.Unlock()

	if shouldStart {
		go s.run(ctx, ch, done)
	}
}

// Report hands value to the per-stream runner over an unbuffered channel. The
// handoff returns when the runner accepts the event, not when the handler
// finishes. A running handler therefore applies backpressure when the next
// event for this stream is reported; if Report is called by readMessages, that
// backpressure stalls all subsequent stream and request dispatch.
func (s *lifecycleStream[T]) Report(ctx context.Context, value T) {
	if ctx == nil || ctx.Err() != nil {
		return
	}

	s.stateMu.Lock()
	handler := s.handler
	ch := s.handlerCh
	s.stateMu.Unlock()
	if handler == nil || ch == nil {
		return
	}
	if ctx.Err() != nil {
		return
	}

	event := lifecycleEvent[T]{
		value:   value,
		handler: handler,
	}
	// Keep stream delivery synchronous so ordering is preserved and callers
	// apply backpressure instead of buffering unbounded events.
	select {
	case ch <- event:
	case <-ctx.Done():
	}
}

// Reset stops accepting reports for the current runner after the caller has
// canceled the lifecycle context passed to On or Start. It returns a channel
// that closes when the detached runner exits.
func (s *lifecycleStream[T]) Reset() <-chan struct{} {
	s.stateMu.Lock()
	done := s.done
	s.handlerCh = nil
	if done == nil {
		s.running = false
		s.stateMu.Unlock()
		return nil
	}
	s.done = nil
	s.running = false
	s.stateMu.Unlock()

	return done
}

func (s *lifecycleStream[T]) startLocked() (chan lifecycleEvent[T], chan struct{}, bool) {
	if s.running {
		return nil, nil, false
	}
	if s.handlerCh == nil {
		s.handlerCh = make(chan lifecycleEvent[T])
	}
	s.done = make(chan struct{})
	s.running = true
	return s.handlerCh, s.done, true
}

func (s *lifecycleStream[T]) run(ctx context.Context, ch <-chan lifecycleEvent[T], done chan<- struct{}) {
	defer close(done)

	for {
		select {
		case event := <-ch:
			// Shutdown wins over delivery when cancellation races with an in-flight report.
			if ctx.Err() != nil {
				return
			}

			if event.handler != nil {
				event.handler(event.value)
			}
		case <-ctx.Done():
			return
		}
	}
}

// registerLifecycleHandler atomically replaces the registered handler on
// stream under streamHandlerStateMu. See Client's godoc for the handler
// concurrency, ordering, and backpressure contract.
func registerLifecycleHandler[T any](c *Client, stream *lifecycleStream[T], handler func(T)) {
	c.streamHandlerStateMu.Lock()
	defer c.streamHandlerStateMu.Unlock()

	stream.On(c.ctx, handler)
}

func (c *Client) reportError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	c.errorStream.Report(ctx, err)
}

// OnError handles asynchronous client errors. It follows the handler contract
// documented on [Client].
func (c *Client) OnError(errHandler func(err error)) {
	registerLifecycleHandler(c, &c.errorStream, errHandler)
}

func (c *Client) reportLedgerClosed(ctx context.Context, ledger *streamtypes.LedgerStream) {
	c.ledgerClosedStream.Report(ctx, ledger)
}

// OnLedgerClosed handles "ledgerClosed" events. It follows the handler
// contract documented on [Client].
func (c *Client) OnLedgerClosed(handler func(ledger *streamtypes.LedgerStream)) {
	registerLifecycleHandler(c, &c.ledgerClosedStream, handler)
}

func (c *Client) reportValidationReceived(ctx context.Context, validation *streamtypes.ValidationStream) {
	c.validationStream.Report(ctx, validation)
}

// OnValidationReceived handles "validationReceived" events. It follows the
// handler contract documented on [Client].
func (c *Client) OnValidationReceived(handler func(validation *streamtypes.ValidationStream)) {
	registerLifecycleHandler(c, &c.validationStream, handler)
}

func (c *Client) reportTransaction(ctx context.Context, transaction *streamtypes.TransactionStream) {
	c.transactionStream.Report(ctx, transaction)
}

// OnTransactions handles "transaction" messages, including messages produced
// by transaction, account, and order-book subscriptions. It follows the
// handler contract documented on [Client].
func (c *Client) OnTransactions(handler func(transactions *streamtypes.TransactionStream)) {
	registerLifecycleHandler(c, &c.transactionStream, handler)
}

func (c *Client) reportPeerStatusChange(ctx context.Context, peerStatus *streamtypes.PeerStatusStream) {
	c.peerStatusStream.Report(ctx, peerStatus)
}

// OnPeerStatusChange handles "peerStatusChange" events. It follows the handler
// contract documented on [Client].
func (c *Client) OnPeerStatusChange(handler func(peerStatus *streamtypes.PeerStatusStream)) {
	registerLifecycleHandler(c, &c.peerStatusStream, handler)
}

func (c *Client) reportOrderBook(ctx context.Context, orderbook *streamtypes.OrderBookStream) {
	c.orderBookStream.Report(ctx, orderbook)
}

// OnOrderBook handles order-book notifications. The XRPL protocol sends these
// as "transaction" messages without identifying which subscription matched,
// so this handler receives every transaction message, as does OnTransactions.
// It follows the handler contract documented on [Client].
func (c *Client) OnOrderBook(handler func(orderbook *streamtypes.OrderBookStream)) {
	registerLifecycleHandler(c, &c.orderBookStream, handler)
}

func (c *Client) reportBookChanges(ctx context.Context, bookChanges *streamtypes.BookChangesStream) {
	c.bookChangesStream.Report(ctx, bookChanges)
}

// OnBookChanges handles "bookChanges" events. It follows the handler contract
// documented on [Client].
func (c *Client) OnBookChanges(handler func(bookChanges *streamtypes.BookChangesStream)) {
	registerLifecycleHandler(c, &c.bookChangesStream, handler)
}

func (c *Client) reportConsensusPhase(ctx context.Context, consensusPhase *streamtypes.ConsensusStream) {
	c.consensusStream.Report(ctx, consensusPhase)
}

// OnConsensusPhase handles "consensusPhase" events. It follows the handler
// contract documented on [Client].
func (c *Client) OnConsensusPhase(handler func(consensusPhase *streamtypes.ConsensusStream)) {
	registerLifecycleHandler(c, &c.consensusStream, handler)
}

func (c *Client) startRegisteredHandlers(ctx context.Context) {
	c.errorStream.Start(ctx)
	c.ledgerClosedStream.Start(ctx)
	c.validationStream.Start(ctx)
	c.transactionStream.Start(ctx)
	c.peerStatusStream.Start(ctx)
	c.orderBookStream.Start(ctx)
	c.bookChangesStream.Start(ctx)
	c.consensusStream.Start(ctx)
}

func (c *Client) resetHandlerRunners() []<-chan struct{} {
	return []<-chan struct{}{
		c.errorStream.Reset(),
		c.ledgerClosedStream.Reset(),
		c.validationStream.Reset(),
		c.transactionStream.Reset(),
		c.peerStatusStream.Reset(),
		c.orderBookStream.Reset(),
		c.bookChangesStream.Reset(),
		c.consensusStream.Reset(),
	}
}

func waitForHandlerRunners(doneChannels []<-chan struct{}) {
	for _, done := range doneChannels {
		if done == nil {
			continue
		}
		<-done
	}
}
