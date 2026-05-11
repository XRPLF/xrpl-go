# Go Performance & Quality

**Version 0.1.0**  
Agent Skills  
March 2026

> **Note:**  
> This document is mainly for agents and LLMs to follow when maintaining,  
> generating, or refactoring Go standard library usage, testing patterns, and performance optimization. Humans  
> may also find it useful, but guidance here is optimized for automation  
> and consistency by AI-assisted workflows.

---

## Abstract

Guidelines for Go standard library usage, testing best practices, and performance optimization. Covers HTTP timeouts, JSON/SQL pitfalls, table-driven tests, race detector, CPU caches, memory allocation, and the garbage collector.

---

## Table of Contents

1. [Standard Library](#1-standard-library) — **HIGH**
   - 1.1 [Always Close HTTP Response Bodies and Other Transient Resources](#11-always-close-http-response-bodies-and-other-transient-resources)
   - 1.2 [Always Return After http.Error — It Does Not Stop Handler Execution](#12-always-return-after-httperror--it-does-not-stop-handler-execution)
   - 1.3 [Always Set HTTP Client Timeouts](#13-always-set-http-client-timeouts)
   - 1.4 [Common JSON Handling Mistakes: Embedding, time.Time, and map[string]any](#14-common-json-handling-mistakes-embedding-timetime-and-mapstringany)
   - 1.5 [Common SQL Mistakes: Open, Pool, Prepared Statements, Null, Rows.Err](#15-common-sql-mistakes-open-pool-prepared-statements-null-rowserr)
   - 1.6 [time.After in a Loop Leaks Memory — Use time.NewTimer Instead](#16-timeafter-in-a-loop-leaks-memory--use-timenewtimer-instead)
   - 1.7 [time.Duration Is in Nanoseconds, Not Milliseconds](#17-timeduration-is-in-nanoseconds-not-milliseconds)
2. [Testing](#2-testing) — **MEDIUM**
   - 2.1 [Always Use the Race Detector in Tests](#21-always-use-the-race-detector-in-tests)
   - 2.2 [Avoid time.Sleep in Tests — Use Retry or Channel Synchronization Instead](#22-avoid-timesleep-in-tests--use-retry-or-channel-synchronization-instead)
   - 2.3 [Categorize Tests With Build Tags, Environment Variables, or Short Mode](#23-categorize-tests-with-build-tags-environment-variables-or-short-mode)
   - 2.4 [Make Time-Dependent Code Testable by Injecting or Passing Time](#24-make-time-dependent-code-testable-by-injecting-or-passing-time)
   - 2.5 [Use Go's Testing Features: Coverage, External Packages, Helpers, and TestMain](#25-use-gos-testing-features-coverage-external-packages-helpers-and-testmain)
   - 2.6 [Use net/http/httptest and testing/iotest for HTTP and I/O Tests](#26-use-nethttphttptest-and-testingiotest-for-http-and-io-tests)
   - 2.7 [Use t.Parallel() for Parallel Tests and -shuffle to Detect Order Dependencies](#27-use-tparallel-for-parallel-tests-and--shuffle-to-detect-order-dependencies)
   - 2.8 [Use Table-Driven Tests for Multiple Cases](#28-use-table-driven-tests-for-multiple-cases)
   - 2.9 [Write Accurate Benchmarks: Reset Timer, Prevent Compiler Optimizations](#29-write-accurate-benchmarks-reset-timer-prevent-compiler-optimizations)
3. [Optimizations](#3-optimizations) — **HIGH**
   - 3.1 [Design Data Structures for CPU Cache Efficiency](#31-design-data-structures-for-cpu-cache-efficiency)
   - 3.2 [Order Struct Fields by Size Descending to Reduce Padding and Memory Usage](#32-order-struct-fields-by-size-descending-to-reduce-padding-and-memory-usage)
   - 3.3 [Preallocate Slices When Size is Known](#33-preallocate-slices-when-size-is-known)
   - 3.4 [Prevent False Sharing in Concurrent Code With Padding or Local Variables](#34-prevent-false-sharing-in-concurrent-code-with-padding-or-local-variables)
   - 3.5 [Profile and Trace Go Applications Before Optimizing](#35-profile-and-trace-go-applications-before-optimizing)
   - 3.6 [Reduce Data Hazards With Instruction-Level Parallelism](#36-reduce-data-hazards-with-instruction-level-parallelism)
   - 3.7 [Reduce Heap Allocations With sync.Pool and API Design](#37-reduce-heap-allocations-with-syncpool-and-api-design)
   - 3.8 [Set GOMAXPROCS to Match Container CPU Quota](#38-set-gomaxprocs-to-match-container-cpu-quota)
   - 3.9 [Tune the GC With GOGC to Reduce GC Pressure Under Load](#39-tune-the-gc-with-gogc-to-reduce-gc-pressure-under-load)
   - 3.10 [Understand Stack vs. Heap: Returning Pointers Forces Heap Allocation](#310-understand-stack-vs-heap-returning-pointers-forces-heap-allocation)
   - 3.11 [Use Fast-Path Inlining to Optimize Hot Code Paths](#311-use-fast-path-inlining-to-optimize-hot-code-paths)
   - 3.12 [Use strings.Builder for Efficient String Concatenation](#312-use-stringsbuilder-for-efficient-string-concatenation)

---

## 1. Standard Library

**Impact: HIGH**

The Go standard library has subtle behaviors in time handling, HTTP client/server, JSON marshaling, and SQL operations. Misusing time.After, forgetting HTTP timeouts, and SQL connection pooling mistakes impact production. These patterns cover standard library best practices.

### 1.1 Always Close HTTP Response Bodies and Other Transient Resources

**Impact: HIGH**

HTTP response bodies must be closed after use, even if you don't read them. Failing to close leaks memory and may prevent TCP connection reuse. Always `defer resp.Body.Close()` immediately after a successful HTTP response, and read the body before closing if you want keep-alive connections.

**Incorrect:**

```go
func getBody(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    // BUG: resp.Body is never closed — memory leak, TCP connection never reused
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}
```

**Correct:**

```go
func getBody(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer func() {
        if err := resp.Body.Close(); err != nil {
            log.Printf("failed to close response body: %v\n", err)
        }
    }()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}

// Even when you don't need the body, close it AND read it for connection reuse:
func getStatusCode(url string, body io.Reader) (int, error) {
    resp, err := http.Post(url, "application/json", body)
    if err != nil {
        return 0, err
    }
    defer func() {
        if err := resp.Body.Close(); err != nil {
            log.Printf("failed to close: %v\n", err)
        }
    }()

    // Read body to enable TCP connection reuse (even if we discard it)
    _, _ = io.Copy(io.Discard, resp.Body)
    return resp.StatusCode, nil
}
```

Why this matters: `http.Response.Body` implements `io.ReadCloser`. Closing the body returns the underlying TCP connection to the pool for reuse. If you close without reading, the default transport may close the connection entirely (preventing keep-alive). If you close after reading, the connection can be reused. Use `io.Copy(io.Discard, resp.Body)` to efficiently drain the body without storing it. Apply the same pattern to `sql.Rows` (defer `rows.Close()`), files (defer `file.Close()`), and any other `io.Closer`.

### 1.2 Always Return After http.Error — It Does Not Stop Handler Execution

**Impact: HIGH**

`http.Error` writes an error response and sets the status code, but it does **not** stop the handler from running. If you omit `return` after `http.Error`, the handler continues executing and may write a second response, leading to garbled output, incorrect status codes, or "superfluous response.WriteHeader call" warnings.

**Incorrect:**

```go
func handler(w http.ResponseWriter, r *http.Request) {
    if err := validate(r); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        // BUG: execution continues — handler writes a second response body
    }
    // This runs even after the error response was sent
    fmt.Fprintln(w, "OK")
}
```

**Correct:**

```go
func handler(w http.ResponseWriter, r *http.Request) {
    if err := validate(r); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return  // Stop execution after sending the error response
    }

    fmt.Fprintln(w, "OK")
}

// Same applies when redirecting:
func redirectHandler(w http.ResponseWriter, r *http.Request) {
    if !authorized(r) {
        http.Redirect(w, r, "/login", http.StatusFound)
        return  // Must return — http.Redirect does not stop execution
    }

    renderDashboard(w, r)
}
```

Why this matters: `http.Error`, `http.Redirect`, and `http.NotFound` all write to the `ResponseWriter` and set headers, but they are ordinary function calls — they don't panic or halt the goroutine. The HTTP response body is the concatenation of everything written to `w`. Without `return`, you send two responses into the same connection, causing protocol errors or double-writing. Always follow any response-sending function with an immediate `return` (or `else` block) to guard subsequent code.

### 1.3 Always Set HTTP Client Timeouts

**Impact: CRITICAL**

The default HTTP client has no timeout—it waits forever for responses. This causes goroutine leaks and resource exhaustion when servers are slow or unresponsive.

**Incorrect:**

```go
func FetchData(url string) ([]byte, error) {
    resp, err := http.Get(url)  // Uses default client - no timeout!
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}
```

**Correct:**

```go
var httpClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   5 * time.Second,
        ResponseHeaderTimeout: 5 * time.Second,
    },
}

func FetchData(url string) ([]byte, error) {
    resp, err := httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}
```

Why this matters: Without timeouts, slow servers cause goroutines to hang indefinitely. In production, this leads to goroutine leaks, memory exhaustion, and cascading failures. Set `Client.Timeout` for overall request timeout and `Transport` timeouts for finer control over connection, TLS handshake, and response header phases.

### 1.4 Common JSON Handling Mistakes: Embedding, time.Time, and map[string]any

**Impact: MEDIUM**

Three common JSON mistakes: (1) embedding a type that implements `json.Marshaler` hijacks the parent's marshaling; (2) marshaling/unmarshaling `time.Time` loses the monotonic clock component; (3) `json.Unmarshal` into `map[string]any` converts all numbers to `float64`.

**Mistake 1: Embedded type hijacks JSON marshaling**

**Mistake 2: time.Time loses monotonic part after marshal/unmarshal**

**Mistake 3: map[string]any converts all numbers to float64**

Why this matters: (1) Embedded types promote all methods including `MarshalJSON`; the embedded type's marshaler overrides the parent's default. (2) `time.Time.Equal()` ignores monotonic time, while `==` includes it — always use `Equal()` to compare times. (3) JSON numbers without decimal points are still parsed as `float64` in untyped maps — use typed structs when possible.

### 1.5 Common SQL Mistakes: Open, Pool, Prepared Statements, Null, Rows.Err

**Impact: HIGH**

Five common mistakes when using `database/sql`: (1) assuming `sql.Open` establishes a connection; (2) not configuring the connection pool; (3) not using prepared statements; (4) mishandling NULL values; (5) not checking `rows.Err()` after iteration.

**Mistake 1: sql.Open doesn't guarantee a connection**

**Mistake 2: Configure the connection pool for production**

**Mistake 3: Use prepared statements for repeated queries**

**Mistake 4: Handle NULL values with sql.NullXXX or pointers**

**Mistake 5: Check rows.Err() after iteration**

Why this matters: Each of these mistakes has real production consequences: silent connection failures at startup, database overload from unconstrained connections, performance and security issues from unprepared statements, panics from NULL values, and silent data loss from unchecked iteration errors.

### 1.6 time.After in a Loop Leaks Memory — Use time.NewTimer Instead

**Impact: HIGH**

`time.After(d)` creates a new timer each call. The resources (including the channel) are only released when the timer fires. In a loop or frequently-called function, calling `time.After` without the timer ever firing leads to memory accumulation. Use `time.NewTimer` with `Reset()` to reuse the timer.

**Incorrect:**

```go
func consumer(ch <-chan Event) {
    for {
        select {
        case event := <-ch:
            handle(event)
        case <-time.After(time.Hour):  // BUG: creates a new timer every iteration
            log.Println("warning: no messages received")
            // Each iteration that handles an event leaves a 1-hour timer
            // consuming ~200 bytes until it fires
            // At 5M events/hour: consumes 1 GB of memory
        }
    }
}
```

**Correct:**

```go
func consumer(ch <-chan Event) {
    timerDuration := 1 * time.Hour
    timer := time.NewTimer(timerDuration)   // Create once
    defer timer.Stop()                      // Stop on exit

    for {
        timer.Reset(timerDuration)          // Reset each iteration — no allocation
        select {
        case event := <-ch:
            handle(event)
        case <-timer.C:
            log.Println("warning: no messages received")
        }
    }
}
```

Why this matters: `time.After` is implemented as `time.NewTimer(d).C` — it creates a full timer object but returns only the channel. Since the caller has no reference to the timer, it can't stop it. The timer's resources are held for the full duration, even if the channel is never read. In a loop processing millions of events, this creates millions of timers. `time.NewTimer` with `Reset()` reuses a single timer with no new heap allocation. This applies to loops, HTTP handlers (called repeatedly), and Kafka consumer functions.

### 1.7 time.Duration Is in Nanoseconds, Not Milliseconds

**Impact: MEDIUM**

`time.Duration` is an alias for `int64`, representing a duration in **nanoseconds**. Passing a bare integer like `1000` creates a 1,000 nanosecond (1 microsecond) duration, not 1 second or 1 millisecond as developers from Java/JavaScript backgrounds might expect. Always use time constants.

**Incorrect:**

```go
// WRONG: creates a 1000-nanosecond ticker, not 1-second
ticker := time.NewTicker(1000)  // 1000 ns = 1 microsecond!

// WRONG: 5000 ms? No — 5000 nanoseconds = 5 microseconds
time.Sleep(5000)
```

**Correct:**

```go
// Use time constants for clarity and correctness
ticker := time.NewTicker(time.Second)          // 1 second
ticker := time.NewTicker(500 * time.Millisecond)  // 500 ms
ticker := time.NewTicker(time.Microsecond)     // 1 microsecond (if that's what you want)

time.Sleep(5 * time.Second)
time.Sleep(100 * time.Millisecond)

// Available time constants:
// time.Nanosecond  = 1
// time.Microsecond = 1000
// time.Millisecond = 1000000
// time.Second      = 1000000000
// time.Minute      = 60000000000
// time.Hour        = 3600000000000

// Building durations programmatically:
duration := time.Duration(n) * time.Second  // n seconds
timeout := 30 * time.Second                 // 30 seconds
```

Why this matters: `time.Duration(1000)` is 1000 nanoseconds — 1 microsecond. In Java, `Thread.sleep(1000)` sleeps for 1 second (milliseconds). Go uses nanoseconds as the base unit. The `time` package provides named constants (`time.Second`, `time.Millisecond`, etc.) precisely to avoid this confusion. Always multiply by a time constant rather than passing a raw integer.

---

## 2. Testing

**Impact: MEDIUM**

Effective testing in Go requires understanding table-driven tests, race detector usage, test execution modes, and benchmarking. Poor testing patterns lead to brittle tests, false confidence, and missed bugs. These patterns ensure robust test suites.

### 2.1 Always Use the Race Detector in Tests

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

### 2.2 Avoid time.Sleep in Tests — Use Retry or Channel Synchronization Instead

**Impact: HIGH**

Using `time.Sleep` in tests to wait for goroutines or async operations makes tests flaky: there is no guarantee the sleep duration will be sufficient on a loaded machine. Replace sleeps with either a **retry loop** (poll until condition or timeout) or **channel synchronization** (block until the goroutine signals completion).

**Incorrect — passive sleep:**

```go
func TestGetBestFoo(t *testing.T) {
    mock := &publisherMock{}
    h := Handler{publisher: mock, n: 2}

    foo := h.getBestFoo(42)
    _ = foo

    time.Sleep(10 * time.Millisecond)  // Flaky: may not be enough on a slow machine
    published := mock.Get()
    if len(published) != 2 {
        t.Fatalf("expected 2, got %d", len(published))
    }
}
```

**Correct option 1 — retry loop:**

```go
// assert polls the assertion function up to maxRetry times, sleeping waitTime between tries.
func assert(t *testing.T, assertion func() bool, maxRetry int, waitTime time.Duration) {
    t.Helper()
    for i := 0; i < maxRetry; i++ {
        if assertion() {
            return
        }
        time.Sleep(waitTime)
    }
    t.Fail()
}

func TestGetBestFoo(t *testing.T) {
    mock := &publisherMock{}
    h := Handler{publisher: mock, n: 2}
    _ = h.getBestFoo(42)

    // Poll up to 30 times, 1ms apart — fast when it succeeds, bounded when it fails
    assert(t, func() bool {
        return len(mock.Get()) == 2
    }, 30, time.Millisecond)
}
```

**Correct option 2 — channel synchronization (preferred):**

```go
type publisherMock struct {
    ch chan []Foo
}

func (p *publisherMock) Publish(got []Foo) {
    p.ch <- got  // Signal the test goroutine immediately
}

func TestGetBestFoo(t *testing.T) {
    mock := &publisherMock{ch: make(chan []Foo, 1)}
    defer close(mock.ch)

    h := Handler{publisher: mock, n: 2}
    _ = h.getBestFoo(42)

    // Block until the goroutine publishes, with a timeout to avoid hanging
    select {
    case got := <-mock.ch:
        if len(got) != 2 {
            t.Fatalf("expected 2, got %d", len(got))
        }
    case <-time.After(time.Second):
        t.Fatal("timeout waiting for Publish")
    }
}
```

Why this matters: A passive `time.Sleep` introduces a fixed wait that is both too long (slows the test suite on fast machines) and too short (flaky on slow CI machines). A retry loop checks the condition as soon as it's ready, reducing wait time when things go right and bounding total wait when they go wrong. Channel synchronization is even better: it makes the test fully deterministic by blocking exactly until the work is done, with a timeout guard. Use the testing library `testify/assert`'s `Eventually` function as a ready-made retry helper.

### 2.3 Categorize Tests With Build Tags, Environment Variables, or Short Mode

**Impact: MEDIUM**

Running all tests together — unit, integration, and end-to-end — slows CI and forces slow external-dependency tests to run where they shouldn't. Categorize tests so developers can run only the relevant subset: unit tests during development, integration tests in CI, etc.

**Approach 1: Build Tags**

**Approach 2: Environment Variables**

**Approach 3: Short Mode**

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

### 2.4 Make Time-Dependent Code Testable by Injecting or Passing Time

**Impact: MEDIUM**

Functions that call `time.Now()` directly are hard to test deterministically: the result changes with the real wall clock. Make time a dependency — either as a function field on a struct, or as a parameter passed by callers — so tests can inject a fixed, predictable time.

**Incorrect — calls time.Now() directly:**

```go
func (c *Cache) TrimOlderThan(since time.Duration) {
    t := time.Now().Add(-since)  // Flaky: depends on real clock; test may fail on slow machine
    for i := 0; i < len(c.events); i++ {
        if c.events[i].Timestamp.After(t) {
            c.events = c.events[i:]
            return
        }
    }
}

// Test is flaky because it relies on real time offsets
func TestCache_TrimOlderThan(t *testing.T) {
    events := []Event{
        {Timestamp: time.Now().Add(-20 * time.Millisecond)},
        {Timestamp: time.Now().Add(-10 * time.Millisecond)},
        {Timestamp: time.Now().Add(10 * time.Millisecond)},
    }
    cache := &Cache{}
    cache.Add(events)
    cache.TrimOlderThan(15 * time.Millisecond)  // Flaky on a loaded machine
    // ...
}
```

**Correct option 1 — inject time as a struct field:**

```go
type now func() time.Time

type Cache struct {
    mu     sync.RWMutex
    events []Event
    now    now  // Unexported; set via factory function
}

func NewCache() *Cache {
    return &Cache{
        events: make([]Event, 0),
        now:    time.Now,  // Production: use real clock
    }
}

func (c *Cache) TrimOlderThan(since time.Duration) {
    t := c.now().Add(-since)  // Uses injected clock
    // ...
}

// Test injects a fixed time — fully deterministic
func TestCache_TrimOlderThan(t *testing.T) {
    fixedTime := parseTime(t, "2020-01-01T12:00:00.06Z")
    events := []Event{
        {Timestamp: parseTime(t, "2020-01-01T12:00:00.04Z")},
        {Timestamp: parseTime(t, "2020-01-01T12:00:00.05Z")},
        {Timestamp: parseTime(t, "2020-01-01T12:00:00.06Z")},
    }
    cache := &Cache{now: func() time.Time { return fixedTime }}
    cache.Add(events)
    cache.TrimOlderThan(15 * time.Millisecond)
    // Result is deterministic regardless of machine speed
}
```

**Correct option 2 — pass current time as a parameter (simpler, but changes the API):**

```go
// Caller provides the current time — no internal clock dependency
func (c *Cache) TrimOlderThan(t time.Time) {
    for i := 0; i < len(c.events); i++ {
        if c.events[i].Timestamp.After(t) {
            c.events = c.events[i:]
            return
        }
    }
}

// Production usage:
cache.TrimOlderThan(time.Now().Add(-since))

// Test usage — completely deterministic:
cache.TrimOlderThan(parseTime(t, "2020-01-01T12:00:00.06Z").Add(-15 * time.Millisecond))
```

Why this matters: When a function embeds `time.Now()`, it couples business logic to the real wall clock. Tests must add sleeps or timing offsets that are inherently fragile on CI. Injecting time as a dependency (function field or parameter) lets tests pin the clock to a known value. Prefer option 2 (pass time explicitly) when it doesn't make the API awkward — it requires no stubs and no unexported fields. Use option 1 (struct field) when the function is called many times and passing time each time would be impractical.

### 2.5 Use Go's Testing Features: Coverage, External Packages, Helpers, and TestMain

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

**Testing from an external package: black-box tests**

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

### 2.6 Use net/http/httptest and testing/iotest for HTTP and I/O Tests

**Impact: MEDIUM**

Go's standard library includes two underused testing utilities: `net/http/httptest` for testing HTTP handlers and clients without a real network, and `testing/iotest` for testing custom `io.Reader`/`io.Writer` implementations and error resilience.

**httptest — testing an HTTP handler:**

```go
func Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("X-API-VERSION", "1.0")
    b, _ := io.ReadAll(r.Body)
    _, _ = w.Write(append([]byte("hello "), b...))
    w.WriteHeader(http.StatusCreated)
}

func TestHandler(t *testing.T) {
    // Build a fake request — no network needed
    req := httptest.NewRequest(http.MethodGet, "http://localhost", strings.NewReader("foo"))
    // Record the response
    w := httptest.NewRecorder()

    Handler(w, req)

    // Assert on headers, body, and status code
    if got := w.Result().Header.Get("X-API-VERSION"); got != "1.0" {
        t.Errorf("api version: expected 1.0, got %s", got)
    }
    body, _ := io.ReadAll(w.Result().Body)
    if got := string(body); got != "hello foo" {
        t.Errorf("body: expected hello foo, got %s", got)
    }
    if w.Result().StatusCode != http.StatusCreated {
        t.FailNow()
    }
}
```

**httptest — testing an HTTP client:**

```go
func TestDurationClientGet(t *testing.T) {
    // Start a local HTTP server — no Docker, no external services
    srv := httptest.NewServer(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            _, _ = w.Write([]byte(`{"duration": 314}`))
        },
    ))
    defer srv.Close()  // Shuts down the server after the test

    client := NewDurationClient()
    duration, err := client.GetDuration(srv.URL, 51.55, -0.12, 51.57, -0.13)
    if err != nil {
        t.Fatal(err)
    }
    if duration != 314*time.Second {
        t.Errorf("expected 314 seconds, got %v", duration)
    }
}
```

**iotest — testing a custom io.Reader:**

```go
// Test that LowerCaseReader correctly transforms input
func TestLowerCaseReader(t *testing.T) {
    err := iotest.TestReader(
        &LowerCaseReader{reader: strings.NewReader("aBcDeFgHiJ")},
        []byte("abcdefghij"),  // Expected output
    )
    if err != nil {
        t.Fatal(err)
    }
}

// Test that a function is resilient to read errors
func TestFoo_ToleratesReadErrors(t *testing.T) {
    // iotest.TimeoutReader fails on the second Read call, then succeeds
    err := foo(iotest.TimeoutReader(strings.NewReader(randomString(1024))))
    if err != nil {
        t.Fatal(err)
    }
}
```

**Available iotest helpers:**

- `iotest.TestReader` — validates a custom `io.Reader` behaves correctly (byte count, fills slice, etc.)

- `iotest.ErrReader` — returns a specified error on every read

- `iotest.HalfReader` — reads only half the requested bytes each call

- `iotest.OneByteReader` — reads one byte per call

- `iotest.TimeoutReader` — fails on the second read, then succeeds

- `iotest.TruncateWriter` — writes to an `io.Writer` but stops silently after n bytes

Why this matters: Spinning up real HTTP servers or writing ad-hoc test mocks for I/O interfaces wastes effort and is slower than necessary. `httptest.NewServer` starts a real (local) HTTP server in milliseconds and tears it down automatically. `httptest.NewRecorder` lets you inspect headers, body, and status without any network. `iotest.TestReader` validates all the subtle contracts of `io.Reader` (partial reads, EOF handling) that are easy to miss. Use these before reaching for Docker or hand-rolled mocks.

### 2.7 Use t.Parallel() for Parallel Tests and -shuffle to Detect Order Dependencies

**Impact: MEDIUM**

By default, Go runs tests within a package sequentially. Slow test suites benefit from parallelism via `t.Parallel()`. Additionally, tests that pass only in a specific order hide shared state bugs — use `-shuffle=on` to detect these dependencies.

**Using t.Parallel():**

```go
// WRONG: shared mutable state in parallel tests causes data races
var counter int  // package-level

func TestIncrement(t *testing.T) {
    t.Parallel()
    counter++  // Data race!
}

// CORRECT: each test owns its state
func TestIncrement(t *testing.T) {
    t.Parallel()
    counter := 0  // Local variable
    counter++
    // ...
}
```

**Important: parallel tests and shared resources**

**Using -shuffle to detect order-dependent tests:**

```go
// This test hides an order dependency — it only passes if TestInit runs first
func TestProcess(t *testing.T) {
    // BUG: relies on globalState set by TestInit
    if globalState == nil {
        t.Fatal("globalState is nil")  // Fails when order is shuffled
    }
}

// CORRECT: each test initializes what it needs
func TestProcess(t *testing.T) {
    state := newState()  // No dependency on other tests
    result := process(state)
    // ...
}
```

Why this matters: Sequential tests in a large package can take minutes; parallel tests with `t.Parallel()` can reduce that to seconds on modern hardware. `-shuffle=on` is essential for finding tests that silently depend on execution order or shared global state — these tests will pass in CI but randomly fail when the suite grows or order changes. Run with `-shuffle=on` regularly to keep your test suite isolated and reliable.

### 2.8 Use Table-Driven Tests for Multiple Cases

**Impact: MEDIUM**

When testing a function with multiple input/output combinations, use table-driven tests instead of duplicating test code. This makes tests more maintainable and easier to extend.

**Incorrect:**

```go
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }

    result = Add(0, 0)
    if result != 0 {
        t.Errorf("Add(0, 0) = %d; want 0", result)
    }

    result = Add(-1, 1)
    if result != 0 {
        t.Errorf("Add(-1, 1) = %d; want 0", result)
    }
    // Lots of duplication...
}
```

**Correct:**

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {name: "positive numbers", a: 2, b: 3, want: 5},
        {name: "zeros", a: 0, b: 0, want: 0},
        {name: "negative and positive", a: -1, b: 1, want: 0},
        {name: "large numbers", a: 1000, b: 2000, want: 3000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

Why this matters: Table-driven tests reduce duplication, make it easy to add new test cases (just add a struct), and provide clear test names. When a test fails, you immediately know which case failed. They also work well with `t.Run()` for subtests.

Benefits:

- Adding new test cases is trivial (one line)

- Test output shows which specific case failed

- Easy to run individual cases: `go test -run TestAdd/zeros`

- Less code to maintain

Pattern: Define a slice of test case structs with inputs, expected outputs, and descriptive names. Loop through and run each as a subtest with `t.Run()`.

### 2.9 Write Accurate Benchmarks: Reset Timer, Prevent Compiler Optimizations

**Impact: MEDIUM**

Go benchmarks are easy to write incorrectly. Three common mistakes produce misleading results: (1) not resetting the timer after expensive setup; (2) allowing the compiler to inline and eliminate the function under test; (3) ignoring environmental variance in micro-benchmarks.

**Mistake 1: Not resetting or pausing the timer**

**Mistake 2: Compiler inlines the function, making the benchmark empty**

**Mistake 3: Micro-benchmark order affects results**

**Running benchmarks:**

```bash
# Run benchmarks (not regular tests)
go test -bench=. ./...

# Increase benchmark time for more stable results
go test -bench=. -benchtime=5s ./...

# Run N times for statistical comparison
go test -bench=. -count=10 | tee stats.txt
benchstat stats.txt  # golang.org/x/perf/cmd/benchstat
```

Why this matters: (1) Setup code before the benchmark loop inflates results if the timer isn't reset. (2) Go's compiler can inline small functions and then eliminate calls with no observable side effects — benchmarks that look fast may actually be benchmarking nothing. Assigning to a local variable prevents elimination; assigning the local to a global prevents the local from being optimized away. (3) Micro-benchmark results can flip based on which benchmark runs first due to CPU state. Use `benchstat` with `-count=10` to compute stable statistics and detect when apparent differences are within noise.

---

## 3. Optimizations

**Impact: HIGH**

Go performance optimization requires understanding CPU caches, memory allocation, inlining, escape analysis, and the garbage collector. Premature optimization wastes time, but knowing performance fundamentals prevents costly mistakes. These patterns cover optimization techniques.

### 3.1 Design Data Structures for CPU Cache Efficiency

**Impact: MEDIUM**

Modern CPUs use multi-level caches (L1/L2/L3) that are orders of magnitude faster than RAM. Code that accesses memory in predictable, contiguous patterns gets most data from cache (fast); code that jumps around memory gets frequent cache misses (slow). Understanding this allows data structure choices that can yield 20-70% performance improvements.

**Cache basics:**

- L1: ~1 ns, L2: ~4 ns, L3: ~10 ns, RAM: ~50-100 ns

- A **cache line** is 64 bytes (8 `int64` values) — the unit of data the CPU moves between RAM and cache

- **Spatial locality**: accessing `s[0]` also caches `s[1]`...`s[7]` in the same line — subsequent accesses are free

- **Temporal locality**: frequently accessed variables stay in cache

**Slice of structs vs. struct of slices:**

```go
// Slice of structs — iterating over field 'a' skips 'b' each time (poor spatial locality)
type Foo struct {
    a int64
    b int64  // Loaded into cache but never used when summing 'a'
}
func sumFoo(foos []Foo) int64 {
    var total int64
    for _, f := range foos {
        total += f.a  // Every other int64 in the cache line is wasted
    }
    return total
}

// Struct of slices — all 'a' values are contiguous in memory (good spatial locality)
type Bar struct {
    a []int64
    b []int64  // Separate slice — not loaded when iterating 'a'
}
func sumBar(bar Bar) int64 {
    var total int64
    for _, v := range bar.a {  // Iterates a densely packed slice — twice as few cache lines
        total += v
    }
    return total
}
// sumBar is ~20% faster because it fetches fewer cache lines
```

**Linked lists vs. slices — predictability matters:**

```go
// Linked list — non-unit stride: the CPU can't predict where next node is in memory
type node struct {
    value int64
    next  *node  // Pointer to anywhere in heap — unpredictable
}
// CPU can't prefetch; every access may be a cache miss

// Slice — unit stride: elements are contiguous; CPU prefetches ahead
func sumSlice(s []int64) int64 {
    var total int64
    for _, v := range s {  // Predictable: each element follows the last
        total += v
    }
    return total
}
// Slice iteration is ~70% faster than linked list iteration (even with same spatial locality)
```

**Critical stride — avoid power-of-2 sized rows in matrices:**

```go
// 512 columns: rows land on the same cache set → conflict misses
type matrix512 [N][512]int64  // 512 * 8 = 4096 bytes = power of 2 → cache conflicts

// 513 columns: rows land on different cache sets → no conflicts
type matrix513 [N][513]int64  // 513 * 8 = 4104 bytes → avoids critical stride

// When reusing the same matrix in a benchmark, matrix512 can be ~50% SLOWER
// because all rows compete for the same cache set
// Fix for benchmarks: create a new matrix each iteration (b.StopTimer/b.StartTimer)
```

**Guidelines:**

1. Prefer slices over linked lists for sequential access patterns

2. When only one field of a struct is used in a hot loop, consider struct-of-slices layout

3. Avoid matrix dimensions that are exact powers of 2 (critical stride)

4. Profile before optimizing — use `go test -bench`, `pprof`, and `perf`

Why this matters: Cache misses are 50-100x more expensive than cache hits. A function that accesses memory in a predictable, contiguous pattern will be dramatically faster than one that jumps around, even if both have identical algorithmic complexity. These are not micro-optimizations — they can represent 20-70% real-world differences in hot code paths.

### 3.2 Order Struct Fields by Size Descending to Reduce Padding and Memory Usage

**Impact: LOW**

Go aligns struct fields to multiples of their own size. Fields declared in a suboptimal order can cause the compiler to insert padding bytes, wasting memory and hurting cache performance. Sorting fields from largest to smallest eliminates unnecessary padding.

**Incorrect — field order causes padding:**

```go
// Foo uses 24 bytes due to alignment padding
type Foo struct {
    b1 byte    // 1 byte at 0x00
    // 7 bytes padding (compiler inserts to align i to multiple of 8)
    i  int64   // 8 bytes at 0x08
    b2 byte    // 1 byte at 0x10
    // 7 bytes padding (to make struct size a multiple of 8)
}
// Total: 1 + 7 (pad) + 8 + 1 + 7 (pad) = 24 bytes
// 14 of 24 bytes are padding!
```

**Correct — largest fields first:**

```go
// Foo uses only 16 bytes after reordering
type Foo struct {
    i  int64  // 8 bytes at 0x00
    b1 byte   // 1 byte at 0x08
    b2 byte   // 1 byte at 0x09
    // 6 bytes padding to reach next multiple of 8
}
// Total: 8 + 1 + 1 + 6 (pad) = 16 bytes
// 33% memory savings just from moving one field

// Rule of thumb: sort fields largest → smallest
// uint64/int64/float64/complex64/pointer: 8 bytes
// uint32/int32/float32: 4 bytes
// uint16/int16: 2 bytes
// uint8/int8/byte/bool: 1 byte
```

**Alignment guarantees in Go: 64-bit architecture**

| Type | Size | Alignment |

|------|------|-----------|

| byte, uint8, int8 | 1 byte | 1 byte |

| uint16, int16 | 2 bytes | 2 bytes |

| uint32, int32, float32 | 4 bytes | 4 bytes |

| uint64, int64, float64, complex64, pointer | 8 bytes | 8 bytes |

| complex128 | 16 bytes | 16 bytes |

**Checking alignment with unsafe:**

```go
import (
    "fmt"
    "unsafe"
)

type Foo struct {
    b1 byte
    i  int64
    b2 byte
}

fmt.Println(unsafe.Sizeof(Foo{}))   // 24 — includes padding
fmt.Println(unsafe.Alignof(Foo{}))  // 8 — alignment of the struct

type FooOpt struct {
    i  int64
    b1 byte
    b2 byte
}

fmt.Println(unsafe.Sizeof(FooOpt{}))  // 16 — no wasted padding
```

Why this matters: Padding bytes add up across millions of struct instances. A 24-byte struct vs. a 16-byte struct is a 33% memory increase; for 1 million instances, that's 8 MB of wasted memory. Beyond memory, larger structs mean fewer fit in a cache line, which increases cache misses when iterating over a slice. The compiler cannot reorder fields automatically (that would break binary compatibility). Use `betteralign` or `fieldalignment` from `golang.org/x/tools/go/analysis/passes/fieldalignment` to detect suboptimal struct layouts.

### 3.3 Preallocate Slices When Size is Known

**Impact: HIGH**

Appending to a slice without preallocation causes multiple memory allocations and copies as the slice grows. When you know the final size, preallocate with `make()` to avoid this overhead.

**Incorrect:**

```go
func ProcessItems(items []Item) []Result {
    var results []Result  // Starts with capacity 0
    for _, item := range items {
        result := process(item)
        results = append(results, result)  // May reallocate multiple times
    }
    return results
}
```

**Correct:**

```go
func ProcessItems(items []Item) []Result {
    results := make([]Result, 0, len(items))  // Preallocate capacity
    for _, item := range items {
        result := process(item)
        results = append(results, result)  // No reallocation needed
    }
    return results
}
```

Why this matters: Each time a slice grows beyond capacity, Go allocates a new larger array and copies all existing elements. For 1000 items, the incorrect version may allocate 10+ times. The correct version allocates once, saving both CPU time and reducing garbage collection pressure.

Benchmark impact: For large slices (10,000+ elements), preallocation can be 2-5x faster and reduce allocations by an order of magnitude. Always preallocate when converting one slice type to another or accumulating results.

Alternative: If you know the exact size, use `make([]Result, len(items))` and index directly instead of append.

### 3.4 Prevent False Sharing in Concurrent Code With Padding or Local Variables

**Impact: MEDIUM**

False sharing occurs when two goroutines on different CPU cores update independent variables that happen to share the same 64-byte cache line. Even though there is no data race, the CPU must invalidate and reload the shared cache line on every write, causing significant performance degradation (~40% in practice).

**Incorrect — false sharing:**

```go
type Result struct {
    sumA int64  // These two fields are likely in the same 64-byte cache line
    sumB int64
}

func count(inputs []Input) Result {
    wg := sync.WaitGroup{}
    wg.Add(2)
    result := Result{}

    go func() {
        for i := range inputs {
            result.sumA += inputs[i].a  // Goroutine 1 writes to cache line
        }
        wg.Done()
    }()

    go func() {
        for i := range inputs {
            result.sumB += inputs[i].b  // Goroutine 2 also writes to same cache line
        }
        wg.Done()
    }()

    wg.Wait()
    return result
    // Each write by goroutine 1 invalidates goroutine 2's copy of the cache line, and vice versa
}
```

**Correct option 1 — add padding to separate fields into different cache lines:**

```go
type Result struct {
    sumA int64
    _    [56]byte  // 56 bytes of padding (int64 = 8 bytes; cache line = 64 bytes; 64-8 = 56)
    sumB int64
    _    [56]byte  // Also pad after sumB if Result is used in a slice
}
// sumA and sumB are now in different cache lines → no false sharing
// ~40% faster in benchmarks
```

**Correct option 2 — use local variables and communicate results via channel:**

```go
func count(inputs []Input) Result {
    chA := make(chan int64)
    chB := make(chan int64)

    go func() {
        var sumA int64  // Goroutine 1 uses its own local variable (on its stack)
        for i := range inputs {
            sumA += inputs[i].a
        }
        chA <- sumA  // Communicate result when done
    }()

    go func() {
        var sumB int64  // Goroutine 2 uses its own local variable
        for i := range inputs {
            sumB += inputs[i].b
        }
        chB <- sumB
    }()

    return Result{sumA: <-chA, sumB: <-chB}
    // Each goroutine works on its own memory → no false sharing
}
```

Why this matters: L1 and L2 caches are per-physical-core. When goroutines on different cores share a cache line and at least one writes to it, the CPU must maintain cache coherency (using the MESI protocol). Every write invalidates the other core's cached copy, forcing a reload from L3 or RAM. This can make concurrent code slower than sequential code. The fix is to ensure that each goroutine's working data is in a separate cache line, either by padding (compile-time guarantee) or by using local variables with channel communication (the idiomatic Go approach).

### 3.5 Profile and Trace Go Applications Before Optimizing

**Impact: HIGH**

Go provides two essential diagnostics tools: `pprof` for profiling (CPU, heap, goroutines, blocking) and the execution tracer for visualizing runtime behavior. Never guess about performance bottlenecks — profile first to identify what actually needs optimization.

**Enabling pprof in a running service:**

```go
import (
    _ "net/http/pprof"  // Blank import registers /debug/pprof endpoints
    "net/http"
    "log"
)

func main() {
    // Your application code...
    log.Fatal(http.ListenAndServe(":8080", nil))
}
// Access profiles at: http://localhost:8080/debug/pprof/
```

**Profile types and when to use them:**

| Profile | Endpoint | Use when |

|---------|----------|----------|

| CPU | `/debug/pprof/profile?seconds=30` | CPU time is high; find hot functions |

| Heap | `/debug/pprof/heap?debug=0` | Memory is high; find allocation sources |

| Goroutine | `/debug/pprof/goroutine?debug=0` | Goroutine count is high; find leaks |

| Block | `/debug/pprof/block` | Goroutines block too long; find contention |

| Mutex | `/debug/pprof/mutex` | Mutex contention; find lock hot spots |

**Using pprof:**

```bash
# CPU profile during a benchmark
go test -bench=. -cpuprofile profile.out
go tool pprof -http=:8080 profile.out   # Opens browser with call graph

# Heap profile — force GC first for accurate data
curl http://localhost:8080/debug/pprof/heap?gc=1 -o heap1.out
# ... wait a few seconds ...
curl http://localhost:8080/debug/pprof/heap?gc=1 -o heap2.out
go tool pprof -http=:8080 -diff_base heap1.out heap2.out  # Compare for leaks

# Enable block profiling (must enable at runtime — disabled by default)
runtime.SetBlockProfileRate(1)  // 1 = record every blocking event
```

**Execution tracer — understand GC and goroutine scheduling:**

```bash
# Collect trace during a benchmark
go test -bench=. -v -trace=trace.out
go tool trace trace.out   # Opens browser with timeline visualization
```

**Reading pprof signals:**

- `runtime.mallocgc` dominates CPU → too many small heap allocations; use `sync.Pool`

- Channel/mutex operations dominate → contention; reduce lock scope or restructure

- `syscall.Read`/`syscall.Write` dominate → I/O bound; improve buffering

**Custom user-level traces:**

```go
import "runtime/trace"

ctx, task := trace.NewTask(context.Background(), "fibonacci")
trace.WithRegion(ctx, "main", func() {
    v = fibonacci(10)
})
task.End()
// Appears in go tool trace with duration distribution
```

**pprof best practices:**

- Enable only one profile at a time (CPU + heap simultaneously = erroneous results)

- CPU and heap profiling are safe in production (activated only when accessed, not continuous)

- Block and mutex profiling have overhead — enable selectively, use low rates in production

- `GODEBUG=gctrace=1` prints a line to stderr each time the GC runs

Why this matters: Optimizing code without profiling wastes effort and risks making things worse. CPU profiling (sample-based, per function) identifies hot code paths. Heap profiling shows allocation sources. The execution tracer (not sample-based, per goroutine) reveals GC behavior and goroutine scheduling problems invisible to the CPU profiler. Together they provide a complete picture of application performance. Use the execution tracer when concurrency or GC frequency is the bottleneck.

### 3.6 Reduce Data Hazards With Instruction-Level Parallelism

**Impact: LOW**

Modern CPUs can execute multiple independent instructions simultaneously (instruction-level parallelism, or ILP). When instructions depend on previous results — called *data hazards* — the CPU stalls waiting for the dependency to resolve. Restructuring code to eliminate data hazards can yield measurable speedups in tight loops.

**What a data hazard looks like:**

```go
// This loop has a data hazard on s[0]:
// Each iteration reads s[0], then writes s[0] — next iteration depends on that write
func addInt(s []int) {
    for i := 0; i < n; i++ {
        s[0] += s[i]  // Read s[0], add s[i], write s[0] — must wait for previous write
    }
}

// The CPU cannot start the next iteration until the current iteration's write to s[0]
// completes. Operations are serialized even though the CPU has parallel execution units.
```

**Eliminating the hazard with a local accumulator:**

```go
// BETTER: Accumulate in a local variable, write to s[0] once at the end
func addInt(s []int) {
    v := 0
    for i := 0; i < n; i++ {
        v += s[i]  // v is a CPU register — no memory dependency hazard
    }
    s[0] = v  // Single write at end
}

// The CPU can pipeline these additions much more efficiently.
// On a 1000-element slice, this can be ~20% faster.
```

**Mixed-operation example:**

```go
// BAD: data hazard — v1 used in three consecutive instructions
func doSomething(s []int32) {
    for i := 0; i < n; i++ {
        v1 := s[0]        // v1 depends on memory load
        s[0] = v1 + 1     // depends on v1
        if v1%2 == 0 {    // depends on v1 — but previous instruction also depended on v1
            s[1] += 2
        }
    }
}

// GOOD: split into independent operations the CPU can execute in parallel
func doSomething(s []int32) {
    for i := 0; i < n; i++ {
        v1 := s[0]
        s[0] = v1 + 1   // Write s[0]
        v2 := v1 % 2    // Compute in parallel with the write to s[0]
        if v2 == 0 {
            s[1] += 2
        }
    }
}
// CPU can execute the write to s[0] and the modulo computation simultaneously
// since they both depend on v1 (already in a register) but not on each other.
```

**How to detect ILP opportunities:**

- Look for hot loops where the same variable is read and written every iteration

- Look for consecutive operations that all depend on the same value

- Profile with `pprof` to confirm the function is a bottleneck before optimizing

- Benchmark before and after: `go test -bench=. -count=10 | benchstat`

**When this matters:**

- Tight inner loops processing large slices (numerical computation, string processing)

- The benefit is CPU-architecture specific and compiler/hardware dependent

- Modern compilers sometimes apply this optimization automatically; verify with benchmarks

Why this matters: CPUs have multiple execution units that can handle arithmetic, memory reads, and branching in parallel — but only when instructions are independent. Data hazards force the CPU to serialize operations, leaving execution units idle. Restructuring hot loops to use local accumulator variables or split dependent computations into independent chains lets the CPU utilize its full parallel execution capacity. The effect is most pronounced in tight numerical loops on large data sets.

### 3.7 Reduce Heap Allocations With sync.Pool and API Design

**Impact: MEDIUM**

Every heap allocation pressures the GC. Three practical techniques to reduce allocations: (1) design APIs to accept caller-provided buffers (sharing down); (2) rely on compiler optimizations like `string(bytes)` in map lookups; (3) use `sync.Pool` to reuse frequently allocated objects.

**Technique 1: API design — let the caller provide the buffer (sharing down):**

**Technique 2: Compiler optimization — string(bytes) map lookup avoids allocation:**

**Technique 3: sync.Pool — reuse frequently allocated objects:**

**sync.Pool rules:**

- Pool is safe for concurrent use by multiple goroutines

- Objects are cleared after each GC cycle (no fixed size or capacity)

- Always reset pooled objects before use (`buffer[:0]`, zeroing fields, etc.)

- Use for objects that are frequently allocated and discarded, not for long-lived state

- The GC will drain the pool; don't rely on objects persisting across GC cycles

Why this matters: Heap allocations are individually inexpensive but cumulative: millions of small allocations fill the heap, triggering frequent GCs that can use 25% of CPU. `sync.Pool` amortizes allocation costs across many callers by reusing objects. The compiler's `string(bytes)` map optimization is a free win — always use it directly instead of assigning to a variable first. For API design, passing a buffer in (sharing down) keeps allocations in the caller's control and often allows stack allocation.

### 3.8 Set GOMAXPROCS to Match Container CPU Quota

**Impact: MEDIUM**

Go sets `GOMAXPROCS` (the number of OS threads running Go code simultaneously) to the number of **host** logical CPUs, not the container's CPU limit. In Kubernetes, a pod with `cpu: 1000m` (1 core) running on an 8-core node will have `GOMAXPROCS=8`, causing CFS throttling and severe latency spikes.

**The problem — GOMAXPROCS defaults to host cores:**

```go
Host machine: 8 logical CPUs
Kubernetes pod: cpu: 1000m (1 core limit = 100ms quota per 100ms period)

Go runtime sets GOMAXPROCS = 8 (host cores)
→ 8 OS threads created, all competing for CPU
→ After consuming 100ms of CPU time, CFS throttles the container
→ All 8 threads freeze until next 100ms period begins
→ Result: latency spikes, throughput degradation
```

**Why CFS throttling is so damaging:**

```go
Timeline (100ms CFS period):
[0ms]     All 8 threads start running
[12ms]    Container has consumed its 100ms quota (8 threads × 12ms ≈ 96ms)
[12ms]    CFS THROTTLES the container — ALL goroutines freeze
[100ms]   CFS period resets, container unthrottled
[100-112ms] Normal operation resumes
...repeat

Effect: ~87% of each 100ms period the service is frozen.
Even with 1 core of CPU quota, the service behaves as if the host is overloaded.
```

**The fix — use automaxprocs:**

```bash
go get go.uber.org/automaxprocs
```

**Manual alternative — set GOMAXPROCS explicitly:**

```go
import (
    "os"
    "runtime"
    "strconv"
)

func init() {
    if val := os.Getenv("GOMAXPROCS"); val != "" {
        if n, err := strconv.Atoi(val); err == nil && n > 0 {
            runtime.GOMAXPROCS(n)
        }
    }
}

// Then set in Kubernetes manifest:
// env:
//   - name: GOMAXPROCS
//     valueFrom:
//       resourceFieldRef:
//         resource: limits.cpu
//         divisor: "1"
```

**Diagnosing the problem:**

```bash
# Check current GOMAXPROCS at runtime
import "runtime"
fmt.Println(runtime.GOMAXPROCS(0))  // 0 = query without changing

# Check if throttling is occurring (Linux, inside container)
cat /sys/fs/cgroup/cpu/cpu.stat | grep throttled
# throttled_time 12345678  ← microseconds of throttling

# Kubernetes: check container CPU throttling metrics
# kubectl top pod --containers
# Or use Prometheus: container_cpu_cfs_throttled_seconds_total
```

**Kubernetes resource configuration reference:**

```yaml
resources:
  requests:
    cpu: "1"       # Scheduling hint: 1 core requested
  limits:
    cpu: "1"       # Hard limit: 1 core = 100ms quota per 100ms period

# With automaxprocs: GOMAXPROCS=1 (matches the limit)
# Without automaxprocs: GOMAXPROCS=8 (or whatever the node has)
```

Why this matters: `GOMAXPROCS` controls how many goroutines run in parallel. When set higher than the container's CPU quota allows, the Linux CFS scheduler throttles the entire container after the quota is exhausted — pausing all goroutines until the next scheduling period. This creates predictable latency spikes (often 300%+ tail latency) that are difficult to diagnose without understanding the CFS/GOMAXPROCS interaction. The fix is a single blank import of `go.uber.org/automaxprocs` in `main.go`, which reads the container CPU quota from `/proc` at startup and sets `GOMAXPROCS` accordingly.

### 3.9 Tune the GC With GOGC to Reduce GC Pressure Under Load

**Impact: MEDIUM**

Go's garbage collector (concurrent mark-and-sweep) is triggered by the `GOGC` environment variable. The default `GOGC=100` means a GC runs when the heap doubles. Understanding this lets you tune GC frequency for your workload, reducing stop-the-world pauses during sudden load spikes.

**How the GC is triggered:**

```bash
# GOGC=100 (default): GC runs when heap size doubles
# If heap is 128 MB after last GC, next GC triggers at 256 MB

# Lower GOGC = more frequent GC, less memory usage
GOGC=50 ./myapp   # GC triggers when heap grows 50% (more GC cycles, less peak memory)

# Higher GOGC = less frequent GC, more memory usage
GOGC=200 ./myapp  # GC triggers when heap doubles + 100% more (fewer GCs, higher peak)

# Disable GC entirely (useful for batch jobs, never for long-running services)
GOGC=off ./myapp

# Print GC traces to stderr
GODEBUG=gctrace=1 go test -bench=. -v
```

**Pre-allocate a minimum heap to reduce GC frequency at startup:**

```go
// For services with a known peak heap, pre-allocate to raise the GC trigger baseline
// Uses virtual memory (lazy allocation via mmap) — won't consume physical RAM until accessed
var min = make([]byte, 1_000_000_000)  // 1 GB reservation

// With GOGC=100 and 1 GB baseline:
// GC won't trigger until heap reaches 2 GB, instead of triggering at 256 MB
// This is effective when:
//   1. You know the expected peak heap size
//   2. Traffic patterns cause rapid heap growth (sudden spike scenario)
//   3. You want to reduce stop-the-world pauses during load spikes
```

**When to tune GOGC:**

```go
Scenario 1: Steady gradual load increase
→ Keep GOGC=100 (default); GC frequency stays moderate

Scenario 2: Sudden traffic spike (0 → 1M users in minutes)
→ Bump GOGC to 200-400 to reduce GC cycles during the spike
→ Or pre-allocate a minimum heap matching expected peak

Scenario 3: Memory-constrained environment
→ Lower GOGC (50-80) to collect more aggressively

Scenario 4: Batch job (runs once, exits)
→ GOGC=off; GC overhead eliminated entirely
```

**GC concepts:**

- Mark stage: traverses all heap objects, marks live ones

- Sweep stage: deallocates unmarked objects

- Two stop-the-world phases per GC cycle (brief), then concurrent operation resumes

- Go GC can use up to 25% of available CPU capacity during concurrent phase

- `debug.FreeOSMemory()` forces a GC and returns free memory to OS (rarely needed)

**Detecting GC pressure:**

```bash
# Print GC events: shows heap size before/after, pause duration
GODEBUG=gctrace=1 ./myapp 2>&1 | grep "^gc"

# Use pprof heap profile to find what's allocating
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# In benchmarks, report allocs per operation
func BenchmarkFoo(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        // ...
    }
}
```

Why this matters: The Go GC is designed to be low-latency, but it still consumes CPU during collection and introduces brief stop-the-world pauses. For services experiencing sudden traffic spikes, frequent GC cycles (because the heap doubles many times rapidly) cause cascading latency increases visible to users. Tuning `GOGC` upward during known high-traffic periods, or pre-allocating a minimum heap baseline, can eliminate these cycles. Profile with `GODEBUG=gctrace=1` first to confirm GC frequency is actually the bottleneck.

### 3.10 Understand Stack vs. Heap: Returning Pointers Forces Heap Allocation

**Impact: MEDIUM**

Go allocates variables on the stack (fast, self-cleaning) or the heap (requires GC, 10x+ slower). The compiler decides via **escape analysis**: if a variable's address outlives the function, it *escapes* to the heap. Unnecessary heap allocations pressure the GC and can dominate application performance.

**Stack vs. heap characteristics:**

- **Stack**: per-goroutine (2 KB initial, grows as needed), LIFO, self-cleaning, no GC involved, ~1 ns alloc

- **Heap**: shared across goroutines, requires GC to reclaim, GC can use 25% of CPU and pause application

**Sharing up → heap; sharing down → stack:**

```go
// Sharing UP: returning a pointer to a local variable — z escapes to the heap
//go:noinline
func sumPtr(x, y int) *int {
    z := x + y
    return &z  // z's address outlives sumPtr → compiler allocates z on heap
}

// Sharing DOWN: passing pointers from parent to child — a and b stay on the stack
//go:noinline
func sum(x, y *int) int {
    return *x + *y  // x and y are owned by caller — no escape
}

func main() {
    a := 3
    b := 2
    c := sum(&a, &b)  // a and b stay on main's stack frame
    _ = c
}
```

**Benchmark comparison:**

```go
var globalValue int
var globalPtr *int

func BenchmarkSumValue(b *testing.B) {
    b.ReportAllocs()
    var local int
    for i := 0; i < b.N; i++ {
        local = sumValue(i, i)  // Stack allocation: 0 allocs/op
    }
    globalValue = local
}

func BenchmarkSumPtr(b *testing.B) {
    b.ReportAllocs()
    var local *int
    for i := 0; i < b.N; i++ {
        local = sumPtr(i, i)  // Heap allocation: 1 alloc/op
    }
    globalValue = *local
}
// BenchmarkSumValue: 1.26 ns/op   0 allocs/op
// BenchmarkSumPtr:  14.84 ns/op   1 allocs/op  ← ~10x slower
```

**Variables that escape to the heap:**

```go
// 1. Returned pointer (sharing up)
func newFoo() *Foo { return &Foo{} }  // Escapes

// 2. Global variables (accessible by all goroutines)
var g *Foo
func set() { g = &Foo{} }  // Escapes

// 3. Pointer sent to a channel
ch <- &Foo{}  // Escapes

// 4. Variable too large for the stack
s := make([]int, n)  // Escapes if n is a variable (size unknown at compile time)
s := make([]int, 10) // May stay on stack (size known)

// 5. Backing array reallocated by append (may escape)
```

**Inspect escape analysis decisions:**

```bash
# See what escapes and why
go build -gcflags "-m=2" ./...
# Example output:
# ./main.go:12:2: z escapes to heap

# Use b.ReportAllocs() in benchmarks to count heap allocs per operation
```

Why this matters: Returning a pointer (e.g., `return &localVar`) is often written to "avoid a copy," but it actually forces a heap allocation that is 10x more expensive than a stack allocation. The GC must eventually collect all heap allocations, using up to 25% of available CPU. In data-intensive hot paths, GC pressure from unnecessary allocations can dominate performance. Favor value semantics unless sharing is semantically required. Use `go build -gcflags "-m=2"` and `b.ReportAllocs()` to audit heap allocations before optimizing.

### 3.11 Use Fast-Path Inlining to Optimize Hot Code Paths

**Impact: LOW**

Go's compiler automatically inlines small functions (within the *inlining budget*). Understanding inlining lets you structure code so the common fast path gets inlined while complex slow paths do not. This technique was used in Go's own `sync.Mutex` implementation to improve mutex acquisition by ~5%.

**How inlining works:**

```go
// Simple function — compiler will inline this (cost < budget)
func sum(a, b int) int {
    return a + b
}

func main() {
    s := sum(3, 2)  // Compiler replaces with: s := 3 + 2
    println(s)
}

// Check inlining decisions:
// $ go build -gcflags "-m=2" ./...
// ./main.go: can inline sum with cost 4 as: func(int, int) int { return a + b }
// ./main.go: inlining call to sum func(int, int) int { return a + b }
//
// If too complex:
// ./main.go: cannot inline foo: function too complex: cost 84 exceeds budget 80
```

**Fast-path inlining — extract slow path into a separate function:**

```go
// BEFORE: The whole Lock function is too complex to inline
func (m *Mutex) Lock() {
    if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
        // Fast path: mutex is unlocked — common case
        return
    }
    // Slow path: mutex is already locked — complex logic
    var waitStartTime int64
    starving := false
    // ... many more lines ...
}
// Cannot inline → every Lock() call has function call overhead

// AFTER: Extract slow path → fast path becomes inlinable
func (m *Mutex) Lock() {
    if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
        return  // Fast path — only these few lines, within inlining budget
    }
    m.lockSlow()  // Slow path — separate function, not inlined
}

func (m *Mutex) lockSlow() {
    var waitStartTime int64
    starving := false
    // ... complex logic ...
}
// Lock() is now inlinable: if mutex is unlocked (common case), zero function call overhead
// ~5% speed improvement for uncontended mutex acquisition
```

**Benefits of inlining:**

1. Removes function call overhead (save/restore registers, stack frame setup)

2. Enables further compiler optimizations — a variable that would escape to heap via a function call may stay on the stack after inlining

**When to apply fast-path inlining:**

- Profile first: identify hot functions with simple fast paths and complex slow paths

- The fast path is the common case (e.g., cache hit, uncontended lock, successful validation)

- The slow path is rare (e.g., cache miss, contended lock, error handling)

- Verify with `go build -gcflags "-m=2"` that the inlining actually happens

Why this matters: Function call overhead is small in absolute terms (~1 ns), but in hot loops called millions of times per second it accumulates. More importantly, inlining enables subsequent compiler optimizations like escape analysis improvements, constant folding, and dead code elimination. The fast-path inlining pattern — where the common path is a few lines calling into a slow path function — is the same technique Go's standard library uses.

### 3.12 Use strings.Builder for Efficient String Concatenation

**Impact: MEDIUM**

Concatenating strings with `+` creates a new string each time because strings are immutable. For loops building strings, this causes O(n²) complexity and excessive allocations. Use `strings.Builder` instead.

**Incorrect:**

```go
func BuildQuery(fields []string) string {
    query := "SELECT "
    for i, field := range fields {
        if i > 0 {
            query += ", "  // New allocation every iteration!
        }
        query += field
    }
    query += " FROM table"
    return query
}
```

**Correct:**

```go
func BuildQuery(fields []string) string {
    var b strings.Builder
    b.WriteString("SELECT ")
    for i, field := range fields {
        if i > 0 {
            b.WriteString(", ")
        }
        b.WriteString(field)
    }
    b.WriteString(" FROM table")
    return b.String()
}
```

Why this matters: String concatenation with `+` copies both strings into a new allocation. With n fields, the incorrect version makes O(n²) allocations and copies. `strings.Builder` grows a single buffer, resulting in O(n) performance. For large strings or loops, this is significantly faster and reduces GC pressure.

When to use: Any loop concatenating strings, building responses, or constructing large strings. For 2-3 simple concatenations, `+` is fine and more readable.

---

## References

1. [https://go.dev/doc/effective_go](https://go.dev/doc/effective_go)
2. [https://go.dev/testing](https://go.dev/testing)
3. [https://github.com/golang/go/wiki/CommonMistakes](https://github.com/golang/go/wiki/CommonMistakes)
