package clientconfig

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWarnIfInsecureScheme(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		wantWarning bool
	}{
		{
			name:        "remote http warns",
			rawURL:      "http://s1.ripple.com:51234/",
			wantWarning: true,
		},
		{
			name:        "remote ws warns",
			rawURL:      "ws://s1.ripple.com:6006/",
			wantWarning: true,
		},
		{
			name:        "other non tls scheme warns",
			rawURL:      "ftp://s1.ripple.com/",
			wantWarning: true,
		},
		{
			name:        "https does not warn",
			rawURL:      "https://s1.ripple.com:51234/",
			wantWarning: false,
		},
		{
			name:        "wss does not warn",
			rawURL:      "wss://s1.ripple.com:6006/",
			wantWarning: false,
		},
		{
			name:        "localhost http does not warn",
			rawURL:      "http://localhost:51234/",
			wantWarning: false,
		},
		{
			name:        "loopback websocket does not warn",
			rawURL:      "ws://127.0.0.1:6006/",
			wantWarning: false,
		},
		{
			name:        "ipv6 loopback does not warn",
			rawURL:      "http://[::1]:51234/",
			wantWarning: false,
		},
		{
			name:        "missing scheme does not warn",
			rawURL:      "localhost",
			wantWarning: false,
		},
		{
			name:        "invalid url does not warn",
			rawURL:      "http://[::1",
			wantWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs := captureLogOutput(t, func() {
				WarnIfInsecureScheme("test", tt.rawURL)
			})

			if tt.wantWarning {
				require.Contains(t, logs, "xrpl-go: warning: test client endpoint")
				require.Contains(t, logs, "uses non-TLS scheme")
			} else {
				require.Empty(t, logs)
			}
		})
	}
}

func captureLogOutput(t *testing.T, fn func()) string {
	t.Helper()

	var buf bytes.Buffer
	previousOutput := log.Writer()
	previousFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(previousOutput)
		log.SetFlags(previousFlags)
	}()

	fn()

	return buf.String()
}
