package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestConnection_ReadMessageEnforcesMaxResponseSize(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		maxResponseSize int64
		expectedErr     error
	}{
		{
			name:            "fail - rejects message over max size",
			message:         strings.Repeat("a", 33),
			maxResponseSize: 32,
			expectedErr:     gorillawebsocket.ErrReadLimit,
		},
		{
			name:            "pass - allows message at max size",
			message:         strings.Repeat("a", 32),
			maxResponseSize: 32,
		},
		{
			name:            "pass - zero max size disables limit",
			message:         strings.Repeat("a", int(defaultMaxResponseSize)+1),
			maxResponseSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newMessageServer(t, tt.message)
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
			conn := newConnection(wsURL, tt.maxResponseSize)
			require.NoError(t, conn.Connect())
			defer func() {
				_ = conn.Disconnect()
			}()

			got, err := conn.ReadMessage()

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.message, string(got))
		})
	}
}

func newMessageServer(t *testing.T, msg string) *httptest.Server {
	t.Helper()

	upgrader := gorillawebsocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		require.NoError(t, conn.WriteMessage(gorillawebsocket.TextMessage, []byte(msg)))
	}))
}
