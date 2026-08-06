//revive:disable var-naming
package common

import "time"

const (
	// LedgerOffset is the number of ledgers to offset when querying ledger data.
	LedgerOffset uint32 = 20

	// DefaultHost is the default host for the XRPL server.
	DefaultHost = "localhost"
	// DefaultMaxRetries is the default bounded retry/polling limit for client operations.
	DefaultMaxRetries = 10
	// DefaultMaxReconnects is the default maximum number of reconnect attempts for websocket.
	DefaultMaxReconnects = 3
	// DefaultRetryDelay is the default delay between retry or polling attempts.
	DefaultRetryDelay = 1 * time.Second
	// DefaultFeeCushion is the default fee cushion multiplier.
	DefaultFeeCushion float64 = 1.2
	// DefaultMaxFeeXRP is the default maximum fee in XRP.
	DefaultMaxFeeXRP = "2"

	// DefaultTimeout is the default timeout for RPC calls (5 seconds).
	DefaultTimeout = 5 * time.Second
)
