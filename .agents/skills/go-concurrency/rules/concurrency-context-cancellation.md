## Always Handle Context Cancellation

**Impact: CRITICAL**

Context cancellation provides a way to stop operations when they're no longer needed. Ignoring `ctx.Done()` causes goroutine leaks and wasted resources.

**Incorrect:**

```go
func FetchData(ctx context.Context, urls []string) []Result {
    results := make([]Result, len(urls))
    for i, url := range urls {
        // Ignores context cancellation!
        resp, _ := http.Get(url)
        results[i] = process(resp)
    }
    return results
}
```

**Correct:**

```go
func FetchData(ctx context.Context, urls []string) ([]Result, error) {
    results := make([]Result, len(urls))
    for i, url := range urls {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()  // Stop early if context cancelled
        default:
        }

        req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            return nil, err
        }
        results[i] = process(resp)
    }
    return results, nil
}
```

Why this matters: When an HTTP request is cancelled (user navigates away) or times out, continuing to fetch remaining URLs wastes resources. Context cancellation propagates through `http.NewRequestWithContext()` and checking `ctx.Done()` lets you stop immediately. Ignoring this causes slow responses and resource leaks.
