package clientconfig

import (
	"net/url"
	"strings"
)

// redactedDisplay returns a user-facing string form of u with all userinfo
// stripped (both username and password) and the leading "//" trimmed for
// network-path references (bare-host inputs reparsed as "//host:port").
//
// url.URL.Redacted() only masks the password component, it leaves the
// username visible, which leaks credentials when the username itself is
// the secret (e.g. http://api-token@host, basic-auth tokens, etc.).
func redactedDisplay(u *url.URL) string {
	displayURL := *u
	displayURL.User = nil
	return strings.TrimPrefix(displayURL.String(), "//")
}
