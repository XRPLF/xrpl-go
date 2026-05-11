## Don't Propagate a Context That May Already Be Canceled

**Impact: MEDIUM**

HTTP request contexts are canceled when the response is written back to the client. If you propagate this context to an asynchronous goroutine (e.g., publishing to a message queue), the goroutine may receive an already-canceled context, causing it to fail silently. Create a detached context for work that must outlive the request.

**Incorrect:**

```go
func handler(w http.ResponseWriter, r *http.Request) {
    response, err := doSomeTask(r.Context(), r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    go func() {
        // BUG: r.Context() is canceled once writeResponse returns
        // If the response is written before publish completes, this will fail
        err := publish(r.Context(), response)
        // ...
    }()

    writeResponse(w, response)  // Context cancels after this
}
```

**Correct:**

```go
// Option 1: Use context.Background() — loses parent values
go func() {
    err := publish(context.Background(), response)
}()

// Option 2: Custom detached context — inherits values but not cancellation
type detach struct{ ctx context.Context }

func (d detach) Deadline() (time.Time, bool) { return time.Time{}, false }
func (d detach) Done() <-chan struct{}        { return nil }
func (d detach) Err() error                  { return nil }
func (d detach) Value(key any) any           { return d.ctx.Value(key) }  // Inherit values

go func() {
    // Detached context: no cancellation signal, but carries parent values (e.g., trace IDs)
    ctx := detach{ctx: r.Context()}
    err := publish(ctx, response)
}()
```

Why this matters: An HTTP request context (`r.Context()`) is canceled in three situations: the client disconnects, the HTTP/2 request is canceled, or the response has been written back to the client. Passing this context to an async goroutine races with response writing. Use `context.Background()` for fire-and-forget operations, or implement a custom "detach" wrapper that inherits values but not the cancellation signal.
