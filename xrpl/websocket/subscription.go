package websocket

import (
	subscribe "github.com/Peersyst/xrpl-go/xrpl/queries/subscription"
	streamtypes "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types"
)

// Subscribe subscribes to the streams and accounts specified in the request.
// It returns a response from the server.
func (c *Client) Subscribe(req *subscribe.Request) (*subscribe.Response, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr subscribe.Response
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// Unsubscribe unsubscribes from the streams and accounts specified in the request.
// It returns a response from the server.
func (c *Client) Unsubscribe(req *subscribe.UnsubscribeRequest) (*subscribe.UnsubscribeResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr subscribe.UnsubscribeResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// Handle errors

// OnError handles "error" events.
// It returns a stream of error streams. Creates a new channel and a goroutine to handle the stream.
func (c *Client) OnError(
	errHandler func(err error),
) {
	go func() {
		defer close(c.errChan)
		for err := range c.errChan {
			errHandler(err)
		}
	}()
}

// Ledger streams

// OnLedgerClosed handles "ledgerClosed" events.
// It returns a stream of ledger streams. Creates a new channel and a goroutine to handle the stream.
func (c *Client) OnLedgerClosed(
	handler func(ledger *streamtypes.LedgerStream),
) {
	c.ledgerClosedChan = make(chan *streamtypes.LedgerStream)
	go func() {
		defer close(c.ledgerClosedChan)
		for ledger := range c.ledgerClosedChan {
			handler(ledger)
		}
	}()
}

// Validation streams

// OnValidationReceived handles "validationReceived" events.
// It returns a stream of validation streams. Creates a new channel and a goroutine to handle the stream.
func (c *Client) OnValidationReceived(
	handler func(validation *streamtypes.ValidationStream),
) {
	c.validationChan = make(chan *streamtypes.ValidationStream)
	go func() {
		defer close(c.validationChan)
		for validation := range c.validationChan {
			handler(validation)
		}
	}()
}

// Transaction streams

// OnTransactions handles "transactions" events.
// It returns a stream of transaction streams. Creates a new channel and a goroutine to handle the stream.
func (c *Client) OnTransactions(
	handler func(transactions *streamtypes.TransactionStream),
) {
	c.transactionChan = make(chan *streamtypes.TransactionStream)
	go func() {
		defer close(c.transactionChan)
		for transaction := range c.transactionChan {
			handler(transaction)
		}
	}()
}

// Peer status streams

// OnPeerStatusChange handles "peerStatus" events.
// It returns a stream of peer status streams. Creates a new channel and a goroutine to handle the stream.
func (c *Client) OnPeerStatusChange(
	handler func(peerStatus *streamtypes.PeerStatusStream),
) {
	c.peerStatusChan = make(chan *streamtypes.PeerStatusStream)
	go func() {
		defer close(c.peerStatusChan)
		for peerStatus := range c.peerStatusChan {
			handler(peerStatus)
		}
	}()
}

// Orderbook streams

// OnOrderBook handles "orderbook" events.
// It returns a stream of orderbook streams. Creates a new channel and a goroutine to handle the stream.
func (c *Client) OnOrderBook(
	handler func(orderbook *streamtypes.OrderBookStream),
) {
	c.orderBookChan = make(chan *streamtypes.OrderBookStream)
	go func() {
		defer close(c.orderBookChan)
		for orderbook := range c.orderBookChan {
			handler(orderbook)
		}
	}()
}

// Book changes streams

// OnBookChanges handles "bookChanges" events.
// It returns a stream of book changes streams. Creates a new channel and a goroutine to handle the stream.
func (c *Client) OnBookChanges(
	handler func(bookChanges *streamtypes.BookChangesStream),
) {
	c.bookChangesChan = make(chan *streamtypes.BookChangesStream)
	go func() {
		defer close(c.bookChangesChan)
		for bookChanges := range c.bookChangesChan {
			handler(bookChanges)
		}
	}()
}

// Consensus streams

// OnConsensusPhase handles "consensusPhase" events.
// It returns a stream of consensus phase streams. Creates a new channel and a goroutine to handle the stream.
func (c *Client) OnConsensusPhase(
	handler func(consensusPhase *streamtypes.ConsensusStream),
) {
	c.consensusChan = make(chan *streamtypes.ConsensusStream)
	go func() {
		defer close(c.consensusChan)
		for consensusPhase := range c.consensusChan {
			handler(consensusPhase)
		}
	}()
}
