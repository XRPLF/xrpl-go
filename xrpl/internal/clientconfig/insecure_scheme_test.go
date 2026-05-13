package clientconfig_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig"
	clientconfigtestutil "github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig/testutil"
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
			name:        "unrecognized scheme does not warn",
			rawURL:      "ftp://s1.ripple.com/",
			wantWarning: false,
		},
		{
			name:        "url with credentials is redacted",
			rawURL:      "http://user:secret@s1.ripple.com/",
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
			name:        "bare localhost does not warn",
			rawURL:      "localhost",
			wantWarning: false,
		},
		{
			name:        "bare loopback ip with port does not warn",
			rawURL:      "127.0.0.1:6006",
			wantWarning: false,
		},
		{
			name:        "bare remote host with port warns",
			rawURL:      "s1.ripple.com:6006",
			wantWarning: true,
		},
		{
			name:        "bare remote host warns",
			rawURL:      "s1.ripple.com",
			wantWarning: true,
		},
		{
			name:        "invalid url does not warn",
			rawURL:      "http://[::1",
			wantWarning: false,
		},
		{
			name:        "empty input does not warn",
			rawURL:      "",
			wantWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs := clientconfigtestutil.CaptureLogOutput(t, func() {
				clientconfig.WarnIfInsecureScheme("test", tt.rawURL)
			})

			if tt.wantWarning {
				require.Contains(t, logs, "xrpl-go: warning: test client endpoint")
				require.Contains(t, logs, "is not using a TLS scheme")
				require.NotContains(t, logs, "secret")
			} else {
				require.Empty(t, logs)
			}
		})
	}
}
