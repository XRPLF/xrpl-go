package websocket

import (
	"context"
	"errors"
)

// A private cause distinguishes our timer from any cause supplied by the caller.
var errConfiguredRequestTimeout = errors.New("configured websocket request timeout")

// requestContextError preserves the first completion recorded by this context.
// A custom caller cause supplements rather than replaces the standard ctx.Err.
func requestContextError(ctx context.Context) error {
	err := ctx.Err()
	if err == nil {
		return nil
	}
	cause := context.Cause(ctx)
	if errors.Is(cause, errConfiguredRequestTimeout) {
		return ErrRequestTimedOut
	}
	if errors.Is(cause, err) {
		return cause
	}
	return errors.Join(err, cause)
}
