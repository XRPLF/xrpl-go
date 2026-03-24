// Package testutil provides testing utilities for websocket functionality.
package testutil

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/gorilla/websocket"
)

// MockWebSocketServer is a test utility for simulating a WebSocket server and capturing sent messages.
type MockWebSocketServer struct {
	Msgs []map[string]any
}

type connFn func(*websocket.Conn)

// TestWebSocketServer starts an HTTP test server that upgrades requests to WebSocket and invokes writeFunc.
func (ms *MockWebSocketServer) TestWebSocketServer(writeFunc connFn) *httptest.Server {
	upgrader := websocket.Upgrader{}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(_ *http.Request) bool { return true }
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade:", err)
		}

		writeFunc(c)
	}))

	return s
}

// ConvertHTTPToWS converts an HTTP(S) URL to its WebSocket (ws:// or wss://) equivalent.
func ConvertHTTPToWS(u string) (string, error) {
	s, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	switch s.Scheme {
	case "http":
		s.Scheme = "ws"
	case "https":
		s.Scheme = "wss"
	}

	return s.String(), nil
}
