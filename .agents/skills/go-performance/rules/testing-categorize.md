## Categorize Tests With Build Tags, Environment Variables, or Short Mode

**Impact: MEDIUM**

Running all tests together — unit, integration, and end-to-end — slows CI and forces slow external-dependency tests to run where they shouldn't. Categorize tests so developers can run only the relevant subset: unit tests during development, integration tests in CI, etc.

**Approach 1: Build Tags**

```go
//go:build integration

package db_test

import (
    "testing"
)

// This file is only compiled (and run) when the "integration" build tag is active
func TestConnectToDatabase(t *testing.T) {
    db := openRealDatabase(t)
    defer db.Close()
    // ...
}
```

```bash
# Run only unit tests (default — no integration tag)
go test ./...

# Run only integration tests
go test -tags=integration ./...

# Run both
go test -tags=integration,e2e ./...
```

**Approach 2: Environment Variables**

```go
func TestWriteToDatabase(t *testing.T) {
    if os.Getenv("INTEGRATION") != "1" {
        t.Skip("skipping integration test; set INTEGRATION=1 to run")
    }
    // ... test using real database
}
```

```bash
# Run integration tests
INTEGRATION=1 go test ./...
```

**Approach 3: Short Mode**

```go
func TestHeavyProcessing(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping heavy test in short mode")
    }
    // ... slow computation
}
```

```bash
# Skip long-running tests
go test -short ./...

# Run all tests (including slow ones)
go test ./...
```

**Combining approaches for a complete test strategy:**

```go
//go:build integration

package service_test

import (
    "os"
    "testing"
)

func TestServiceIntegration(t *testing.T) {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        t.Skip("DATABASE_URL not set")
    }
    // ...
}
```

Why this matters: Unit tests should run in milliseconds; integration tests may take seconds or minutes and require external services. Mixing them wastes developer time and creates flaky pipelines. Build tags are compile-time (cleanest separation); environment variables are runtime (more flexible); short mode is a convention for quick runs. Use build tags for integration/e2e categories and short mode for optionally-skippable slow unit tests.
