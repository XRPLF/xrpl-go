// Package clientconfig contains shared helpers for XRPL client configuration.
package clientconfig

import (
	"log"
	"net"
	"net/url"
	"strings"
	"sync/atomic"
)

var logger atomic.Pointer[log.Logger]

func init() {
	SetLogger(log.Default())
}

// SetLogger swaps the logger used for warnings and returns the previous logger.
func SetLogger(l *log.Logger) *log.Logger {
	return logger.Swap(l)
}

// WarnIfInsecureScheme logs when a remote endpoint is not using a TLS scheme.
// Recognized schemes are http, https, ws, wss; inputs without a scheme are
// treated as bare host strings (e.g. "s1.ripple.com:6006") and warned about
// since they imply non-TLS. Unrecognized schemes and malformed inputs are
// silently ignored to avoid misleading warnings.
func WarnIfInsecureScheme(clientName, rawURL string) {
	if rawURL == "" {
		return
	}

	l := logger.Load()
	if l == nil {
		return
	}

	parseInput := rawURL
	bareHost := !strings.Contains(rawURL, "://")
	if bareHost {
		// Reparse as a network-path reference so url.Parse extracts the
		// hostname. This avoids Go treating inputs like "s1.ripple.com:6006"
		// as Scheme="s1.ripple.com".
		parseInput = "//" + rawURL
	}

	u, err := url.Parse(parseInput)
	if err != nil || u.Hostname() == "" {
		return
	}

	if !bareHost {
		scheme := strings.ToLower(u.Scheme)
		if scheme != "http" && scheme != "ws" {
			return
		}
	}

	if isLocalHost(u.Hostname()) {
		return
	}

	l.Printf(
		"xrpl-go: warning: %s client endpoint %q is not using a TLS scheme; use https:// or wss:// for production",
		clientName,
		redactedDisplay(u),
	)
}

func isLocalHost(host string) bool {
	if strings.EqualFold(strings.TrimSuffix(host, "."), "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsUnspecified())
}
