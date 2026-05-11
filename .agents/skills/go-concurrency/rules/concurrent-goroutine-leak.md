## Prevent Goroutine Leaks with Context or Channels

**Impact: CRITICAL**

Goroutines that run indefinitely without a way to stop them will leak memory and resources. Every goroutine should have a clear termination condition using context cancellation or channel closure.

**Incorrect:**

```go
func StartWorker(jobs <-chan Job) {
    go func() {
        for job := range jobs {
            process(job)
        }
    }()
    // Goroutine has no way to be stopped
    // If jobs channel is never closed, this leaks forever
}
```

**Correct:**

```go
func StartWorker(ctx context.Context, jobs <-chan Job) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                // Graceful shutdown when context is cancelled
                return
            case job, ok := <-jobs:
                if !ok {
                    // Channel closed, exit cleanly
                    return
                }
                process(job)
            }
        }
    }()
}
```

Why this matters: In a long-running server, goroutine leaks cause memory growth that eventually crashes your application. Each leaked goroutine holds its stack (minimum 2KB) and any referenced memory. With thousands of requests, this adds up fast.

Pattern: Use `context.Context` for cancellation, always check channel closure with `job, ok := <-ch`, and ensure every goroutine has a clear exit path. Test with `go test -race` to catch goroutine issues.
