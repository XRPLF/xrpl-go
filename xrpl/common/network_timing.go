//go:build !integration_localnet

package common

import "time"

// DefaultRetryDelay is the default delay between retry or polling attempts.
const DefaultRetryDelay = 1 * time.Second
