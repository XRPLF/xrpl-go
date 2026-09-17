//go:build integration_localnet

package common

import "time"

// DefaultRetryDelay polls localnet finality at the local ledger-close rate.
const DefaultRetryDelay = 100 * time.Millisecond
