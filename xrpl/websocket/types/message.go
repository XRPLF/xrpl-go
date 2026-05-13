// Package types contains data structures for websocket message types.
// revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types"
)

// Message is a struct that represents a message from the websocket.
// It contains every field that can be found in a websocket message.
type Message struct {
	// Type field from all streams
	Type types.Type `json:"type"`
	// ID field from all websocket requests
	ID uint64 `json:"id"`
}

// IsRequest returns true if the message has an ID.
// This is true for all websocket requests.
func (m *Message) IsRequest() bool {
	return m.ID != 0
}

// IsStream returns true if the message has a Type.
// This is true for all websocket streams.
func (m *Message) IsStream() bool {
	return m.Type != ""
}
