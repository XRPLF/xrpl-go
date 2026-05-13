package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Connection is a wrapper around a websocket connection.
// It provides a method to read messages from the connection.
// All methods are safe for concurrent use.
type Connection struct {
	conn *websocket.Conn
	url  string

	mu      sync.Mutex
	readMu  sync.Mutex
	writeMu sync.Mutex
}

// NewConnection creates a new Connection.
func NewConnection(url string) *Connection {
	return &Connection{
		url: url,
	}
}

// Connect opens a websocket connection to the server.
func (c *Connection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// Disconnect closes the websocket connection and sets the connection to nil.
// It returns an error if the connection is not connected.
func (c *Connection) Disconnect() error {
	c.mu.Lock()
	conn := c.conn
	if conn == nil {
		c.mu.Unlock()
		return ErrNotConnected
	}
	c.conn = nil
	c.mu.Unlock()

	if err := conn.Close(); err != nil {
		return err
	}
	return nil
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
	c.readMu.Lock()
	defer c.readMu.Unlock()

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return nil, ErrNotConnected
	}
	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return message, nil
}

// WriteMessage writes a message to the connection.
// It returns an error if the message is not written.
func (c *Connection) WriteMessage(message []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return ErrNotConnected
	}
	return conn.WriteMessage(websocket.TextMessage, message)
}
