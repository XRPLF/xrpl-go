## Use errgroup for Parallel Goroutines That Need Error Propagation

**Impact: MEDIUM**

When running multiple goroutines in parallel and collecting errors, rolling your own error handling with channels, mutexes, or error slices quickly becomes complex. `golang.org/x/sync/errgroup` provides a clean API: spin up goroutines with `g.Go()`, wait for all with `g.Wait()` which returns the first non-nil error.

**Incorrect:**

```go
// Manual error handling with WaitGroup + channel: verbose and error-prone
func handler(ctx context.Context, circles []Circle) ([]Result, error) {
    results := make([]Result, len(circles))
    wg := sync.WaitGroup{}
    var firstErr error
    var mu sync.Mutex

    for i, circle := range circles {
        i, circle := i, circle
        wg.Add(1)
        go func() {
            defer wg.Done()
            result, err := foo(ctx, circle)
            if err != nil {
                mu.Lock()
                if firstErr == nil {
                    firstErr = err  // Only capture first error
                }
                mu.Unlock()
                return
            }
            results[i] = result
        }()
    }
    wg.Wait()
    if firstErr != nil {
        return nil, firstErr
    }
    return results, nil
}
```

**Correct:**

```go
import "golang.org/x/sync/errgroup"

func handler(ctx context.Context, circles []Circle) ([]Result, error) {
    results := make([]Result, len(circles))
    g, ctx := errgroup.WithContext(ctx)  // Shared context: canceled on first error

    for i, circle := range circles {
        i, circle := i, circle  // Capture loop variables
        g.Go(func() error {
            result, err := foo(ctx, circle)
            if err != nil {
                return err          // errgroup captures the first error
            }
            results[i] = result
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err             // Returns first non-nil error
    }
    return results, nil
}
// errgroup.WithContext also cancels ctx when the first error occurs,
// stopping other goroutines that check ctx.Err()
```

Why this matters: `errgroup.WithContext` creates a group and a derived context. `g.Go` runs a function in a new goroutine. `g.Wait` blocks until all goroutines complete and returns the first non-nil error. The shared context is automatically canceled when the first goroutine returns an error, allowing other goroutines to stop early if they check the context. Install with: `go get golang.org/x/sync/errgroup`.
