package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type websocketConnection interface {
	Close() error
	SetReadLimit(limit int64)
	SetReadDeadline(t time.Time) error
	ReadMessage() (messageType int, p []byte, err error)
	SetWriteDeadline(t time.Time) error
	WriteMessage(messageType int, data []byte) error
}

// Connection is a wrapper around a websocket connection.
// It provides a method to read messages from the connection.
// All methods are safe for concurrent use.
type Connection struct {
	conn            websocketConnection
	preparing       websocketConnection
	disconnecting   websocketConnection
	url             string
	maxResponseSize int64

	mu         sync.Mutex
	readMu     sync.Mutex
	writeOnce  sync.Once
	writeToken chan struct{}
}

// NewConnection creates a new Connection.
func NewConnection(url string) *Connection {
	return newConnection(url, defaultMaxResponseSize)
}

func newConnection(url string, maxResponseSize int64) *Connection {
	return &Connection{
		url:             url,
		maxResponseSize: maxResponseSize,
	}
}

// Connect opens a websocket connection to the server.
func (c *Connection) Connect() error {
	ctx := context.Background()
	conn, err := c.beginConnect(ctx)
	if err != nil {
		return err
	}
	return c.publishSocket(ctx, conn)
}

// beginConnect dials a socket and keeps it unavailable to normal reads and
// writes until publishSocket is called.
func (c *Connection) beginConnect(ctx context.Context) (websocketConnection, error) {
	c.mu.Lock()
	alreadyConnected := c.conn != nil || c.preparing != nil || c.disconnecting != nil
	c.mu.Unlock()
	if alreadyConnected {
		return nil, ErrAlreadyConnected
	}

	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, c.url, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if c.maxResponseSize > 0 {
		conn.SetReadLimit(c.maxResponseSize)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil || c.preparing != nil || c.disconnecting != nil {
		_ = conn.Close()
		return nil, ErrAlreadyConnected
	}
	if err := ctx.Err(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	c.preparing = conn
	return conn, nil
}

func (c *Connection) publishSocket(ctx context.Context, conn websocketConnection) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.preparing != conn {
		if err := ctx.Err(); err != nil {
			return err
		}
		return ErrNotConnected
	}
	if err := ctx.Err(); err != nil {
		c.preparing = nil
		_ = conn.Close()
		return err
	}
	c.preparing = nil
	c.conn = conn
	return nil
}

// Disconnect closes the websocket connection and sets the connection to nil.
// It returns an error if the connection is not connected.
func (c *Connection) Disconnect() error {
	return c.disconnect(nil)
}

// disconnect claims the current or preparing socket before lifecycleCancel
// can wake an operation that invalidates it. The claim makes the disconnect
// successful for the socket that was active when this method started.
func (c *Connection) disconnect(lifecycleCancel func()) error {
	c.mu.Lock()
	conn := c.conn
	if conn != nil {
		c.conn = nil
	} else {
		conn = c.preparing
		c.preparing = nil
	}
	if conn != nil {
		c.disconnecting = conn
	}
	c.mu.Unlock()

	if lifecycleCancel != nil {
		lifecycleCancel()
	}
	if conn == nil {
		return ErrNotConnected
	}

	err := conn.Close()
	c.mu.Lock()
	if c.disconnecting == conn {
		c.disconnecting = nil
	}
	c.mu.Unlock()
	return err
}

// invalidateSocket removes a failed socket only if it is still current or is
// being prepared. It closes the exact failed socket unless Disconnect already
// claimed it, so a late failure cannot affect a replacement.
func (c *Connection) invalidateSocket(failed websocketConnection) error {
	c.mu.Lock()
	if c.conn == failed {
		c.conn = nil
	}
	if c.preparing == failed {
		c.preparing = nil
	}
	claimedByDisconnect := c.disconnecting == failed
	c.mu.Unlock()
	if claimedByDisconnect {
		return nil
	}
	return failed.Close()
}

// IsConnected returns true if the connection is connected.
func (c *Connection) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn != nil
}

// ReadMessage reads a message from the connection.
// It returns the message and an error if the message is not read.
// This method is blocking, it will block until a message is read.
func (c *Connection) ReadMessage() ([]byte, error) {
	return c.readMessage(time.Time{})
}

func (c *Connection) readMessage(deadline time.Time) ([]byte, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return nil, ErrNotConnected
	}
	return c.readMessageFrom(conn, deadline)
}

// readMessageFrom reads one message from an exact socket. A non-zero deadline
// applies only to this read and is cleared again on success. Identity discovery
// uses it before the socket is published to normal requests.
func (c *Connection) readMessageFrom(conn websocketConnection, deadline time.Time) ([]byte, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	if err := conn.SetReadDeadline(deadline); err != nil {
		return nil, err
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if !deadline.IsZero() {
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			return nil, err
		}
	}
	return message, nil
}

// WriteMessage writes a message to the connection.
// It returns an error if the message is not written.
func (c *Connection) WriteMessage(message []byte) error {
	return c.writeMessage(context.Background(), message, 0)
}

func (c *Connection) writeMessage(ctx context.Context, message []byte, timeout time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return ErrNotConnected
	}
	return c.writeMessageTo(ctx, conn, message, timeout)
}

func (c *Connection) writeMessageTo(
	ctx context.Context,
	conn websocketConnection,
	message []byte,
	timeout time.Duration,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := c.acquireWrite(ctx); err != nil {
		return err
	}
	defer c.releaseWrite()

	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	if contextDeadline, ok := ctx.Deadline(); ok && (deadline.IsZero() || contextDeadline.Before(deadline)) {
		deadline = contextDeadline
	}
	if err := conn.SetWriteDeadline(deadline); err != nil {
		_ = c.invalidateSocket(conn)
		return err
	}

	var (
		writeDone chan struct{}
		watchDone chan struct{}
	)
	if ctx.Done() != nil {
		writeDone = make(chan struct{})
		watchDone = make(chan struct{})
		go func() {
			defer close(watchDone)
			select {
			case <-ctx.Done():
				select {
				case <-writeDone:
					return
				default:
					_ = c.invalidateSocket(conn)
				}
			case <-writeDone:
			}
		}()
	}

	writeErr := conn.WriteMessage(websocket.TextMessage, message)
	if writeDone != nil {
		close(writeDone)
		<-watchDone
	}
	if writeErr != nil {
		_ = c.invalidateSocket(conn)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if writeErr != nil {
		return writeErr
	}
	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		_ = c.invalidateSocket(conn)
		return err
	}
	return nil
}

func (c *Connection) acquireWrite(ctx context.Context) error {
	c.writeOnce.Do(func() {
		c.writeToken = make(chan struct{}, 1)
		c.writeToken <- struct{}{}
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.writeToken:
		return nil
	}
}

func (c *Connection) releaseWrite() {
	c.writeToken <- struct{}{}
}
