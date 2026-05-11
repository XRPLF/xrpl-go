## Always Use the Race Detector in Tests

**Impact: HIGH**

Data races are undefined behavior and cause mysterious bugs. Go's race detector (`-race` flag) finds these issues during testing. Not using it means shipping race conditions to production.

**Incorrect:**

```go
// Running tests without race detection
$ go test ./...
```

**Correct:**

```go
// Always run with race detector
$ go test -race ./...

// In CI/CD pipelines
go test -race -timeout 10m ./...

// For specific packages with known goroutines
go test -race -v ./internal/worker
```

Why this matters: Race conditions are non-deterministic—they may not appear in normal test runs but manifest in production under load. The race detector instruments your code to detect concurrent access to shared memory. While it adds overhead (typically 10x slower, uses 10x more memory), catching races during development prevents production incidents.

Best practice: Run `-race` in CI pipelines. Include it in pre-commit hooks for critical packages. Accept the performance hit during testing—it's worth it. Note that the race detector only finds races that actually execute during the test, so good test coverage is essential.
