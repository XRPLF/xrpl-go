package clientconfig

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedactedDisplay(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   string
	}{
		{
			name:   "no userinfo",
			rawURL: "http://s1.ripple.com/",
			want:   "http://s1.ripple.com/",
		},
		{
			name:   "username only (token-as-user)",
			rawURL: "http://token@s1.ripple.com/",
			want:   "http://s1.ripple.com/",
		},
		{
			name:   "user and password",
			rawURL: "http://user:secret@s1.ripple.com/",
			want:   "http://s1.ripple.com/",
		},
		{
			name:   "username only with port and path",
			rawURL: "http://api-token@s1.ripple.com:51234/path",
			want:   "http://s1.ripple.com:51234/path",
		},
		{
			name:   "user and password with port",
			rawURL: "ws://user:secret@s1.ripple.com:6006",
			want:   "ws://s1.ripple.com:6006",
		},
		{
			name:   "bare host with port (network-path reference)",
			rawURL: "//s1.ripple.com:6006",
			want:   "s1.ripple.com:6006",
		},
		{
			name:   "bare host with username (network-path reference)",
			rawURL: "//token@s1.ripple.com:6006",
			want:   "s1.ripple.com:6006",
		},
		{
			name:   "bare host with user and password (network-path reference)",
			rawURL: "//user:secret@s1.ripple.com:6006",
			want:   "s1.ripple.com:6006",
		},
		{
			name:   "ipv6 host preserves brackets",
			rawURL: "http://user@[2001:db8::1]:51234/",
			want:   "http://[2001:db8::1]:51234/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.rawURL)
			require.NoError(t, err)

			got := redactedDisplay(u)

			require.Equal(t, tt.want, got)
			require.NotContains(t, got, "token")
			require.NotContains(t, got, "secret")
			require.NotContains(t, got, "api-token")
		})
	}
}
