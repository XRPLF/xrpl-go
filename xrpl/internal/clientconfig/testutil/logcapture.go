// Package testutil exposes helpers for tests that exercise clientconfig.
package testutil

import (
	"bytes"
	"log"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/internal/clientconfig"
)

// CaptureLogOutput swaps the clientconfig logger for the duration of fn and
// returns everything it wrote. The previous logger is restored via t.Cleanup.
// Unlike directly mutating log.Default() / log.SetOutput, this does not
// touch the stdlib global logger, so unrelated tests in the same process
// remain unaffected.
func CaptureLogOutput(t *testing.T, fn func()) string {
	t.Helper()

	var buf bytes.Buffer
	prev := clientconfig.SetLogger(log.New(&buf, "", 0))
	t.Cleanup(func() {
		clientconfig.SetLogger(prev)
	})

	fn()

	return buf.String()
}
