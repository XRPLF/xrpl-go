## Use Go's Testing Features: Coverage, External Packages, Helpers, and TestMain

**Impact: MEDIUM**

Go provides several testing features that developers often overlook: coverage profiling, external test packages for black-box testing, utility functions that accept `*testing.T`, and `TestMain` for package-level setup and teardown.

**Code coverage:**

```bash
# Generate a coverage profile
go test -coverprofile=coverage.out ./...

# View coverage as HTML in browser
go tool cover -html=coverage.out

# Include coverage from other packages (cross-package coverage)
go test -coverpkg=./... -coverprofile=coverage.out ./...
```

**Testing from an external package (black-box tests):**

```go
// counter.go — package counter
package counter

import "sync/atomic"

var count uint64

func Inc() uint64 {
    atomic.AddUint64(&count, 1)
    return count
}

// counter_test.go — use package counter_test to enforce testing only the public API
package counter_test  // External test package — cannot access unexported count variable

import (
    "testing"
    "myapp/counter"
)

func TestCount(t *testing.T) {
    if counter.Inc() != 1 {
        t.Errorf("expected 1")
    }
}
```

**Utility functions that accept *testing.T:**

```go
// AWKWARD: caller must check errors
func TestCustomer(t *testing.T) {
    customer, err := createCustomer("foo")
    if err != nil {
        t.Fatal(err)
    }
    // ... rest of test
}

// BETTER: utility function calls t.Fatal directly — test body is cleaner
func TestCustomer(t *testing.T) {
    customer := createCustomer(t, "foo")  // Fails the test automatically on error
    // ... rest of test
}

func createCustomer(t *testing.T, someArg string) Customer {
    t.Helper()  // Marks this as a helper — errors point to the caller's line
    customer, err := buildCustomer(someArg)
    if err != nil {
        t.Fatal(err)
    }
    return customer
}
```

**Setup and teardown with t.Cleanup and TestMain:**

```go
// Per-test teardown with t.Cleanup (preferred over defer)
func createConnection(t *testing.T, dsn string) *sql.DB {
    t.Helper()
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        t.FailNow()
    }
    t.Cleanup(func() {
        _ = db.Close()  // Runs at the end of the test automatically
    })
    return db
}

func TestMySQLIntegration(t *testing.T) {
    db := createConnection(t, "tcp(localhost:3306)/db")
    // Use db — no need to manually close it
}

// Per-package setup and teardown with TestMain
func TestMain(m *testing.M) {
    setupMySQL()           // Runs once before all tests in this package
    code := m.Run()        // Run all tests
    teardownMySQL()        // Runs once after all tests
    os.Exit(code)
}
```

Why this matters: (1) `-coverprofile` shows which code paths aren't tested; use `-coverpkg=./...` to see cross-package coverage. (2) External test packages (`package foo_test`) enforce API-level testing: if a refactor doesn't change the API, tests stay green. (3) Utility functions with `*testing.T` reduce boilerplate and centralize error handling; mark them with `t.Helper()` so error lines point to the test, not the helper. (4) `t.Cleanup` is preferable to `defer` in utility functions because it runs at test end regardless of early exits; `TestMain` handles environment setup once per package.
