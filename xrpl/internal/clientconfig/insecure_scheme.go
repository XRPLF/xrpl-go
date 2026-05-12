// Package clientconfig contains shared helpers for XRPL client configuration.
package clientconfig

import (
	"log"
	"net"
	"net/url"
	"strings"
)

// WarnIfInsecureScheme logs when a remote endpoint URL uses a scheme that does not imply TLS.
func WarnIfInsecureScheme(clientName, rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme == "" || isTLSScheme(scheme) || isLocalHost(u.Hostname()) {
		return
	}

	log.Printf(
		"xrpl-go: warning: %s client endpoint %q uses non-TLS scheme %q; use a TLS endpoint for production",
		clientName,
		rawURL,
		scheme,
	)
}

func isTLSScheme(scheme string) bool {
	return scheme == "https" || scheme == "wss"
}

func isLocalHost(host string) bool {
	if strings.EqualFold(strings.TrimSuffix(host, "."), "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
