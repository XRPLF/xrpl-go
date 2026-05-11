# Go Concurrency

**Version 0.1.0**  
Agent Skills  
March 2026

> **Note:**  
> This document is mainly for agents and LLMs to follow when maintaining,  
> generating, or refactoring Go goroutines, channels, mutexes, sync primitives, and data races. Humans  
> may also find it useful, but guidance here is optimized for automation  
> and consistency by AI-assisted workflows.

---

## Abstract

Guidelines for safe and correct Go concurrency. Covers goroutine lifecycle, data races, channel vs mutex patterns, context propagation, goroutine leaks, sync primitives, and errgroup usage.

---

## Table of Contents

1. [Concurrency Foundations](#1-concurrency-foundations) — **CRITICAL**
   - 1.1 [Always Handle Context Cancellation](#11-always-handle-context-cancellation)
   - 1.2 [Concurrency Is Not Always Faster Than Sequential Code](#12-concurrency-is-not-always-faster-than-sequential-code)
   - 1.3 [Don't Propagate a Context That May Already Be Canceled](#13-dont-propagate-a-context-that-may-already-be-canceled)
   - 1.4 [Know When to Use Channels Versus Mutexes](#14-know-when-to-use-channels-versus-mutexes)
   - 1.5 [Set Worker Pool Size Based on Workload Type (CPU vs I/O)](#15-set-worker-pool-size-based-on-workload-type-cpu-vs-io)
   - 1.6 [Understand the Difference Between Concurrency and Parallelism](#16-understand-the-difference-between-concurrency-and-parallelism)
2. [Concurrency Practice](#2-concurrency-practice) — **CRITICAL**
   - 2.1 [Assigning a Map or Slice Doesn't Copy the Data — Protect the Whole Operation](#21-assigning-a-map-or-slice-doesnt-copy-the-data--protect-the-whole-operation)
   - 2.2 [Call wg.Add() Before Spawning a Goroutine, Not Inside It](#22-call-wgadd-before-spawning-a-goroutine-not-inside-it)
   - 2.3 [Concurrent append on a Shared Slice Can Cause a Data Race](#23-concurrent-append-on-a-shared-slice-can-cause-a-data-race)
   - 2.4 [Goroutine Closures Capture Loop Variables by Reference](#24-goroutine-closures-capture-loop-variables-by-reference)
   - 2.5 [Never Copy sync Types (Mutex, WaitGroup, Cond)](#25-never-copy-sync-types-mutex-waitgroup-cond)
   - 2.6 [Prevent Goroutine Leaks with Context or Channels](#26-prevent-goroutine-leaks-with-context-or-channels)
   - 2.7 [Protect Shared State with Mutexes or Channels](#27-protect-shared-state-with-mutexes-or-channels)
   - 2.8 [select With Multiple Ready Channels Chooses Randomly, Not by Order](#28-select-with-multiple-ready-channels-chooses-randomly-not-by-order)
   - 2.9 [String Formatting Can Cause Data Races or Deadlocks in Concurrent Code](#29-string-formatting-can-cause-data-races-or-deadlocks-in-concurrent-code)
   - 2.10 [Understand Nil Channel Behavior](#210-understand-nil-channel-behavior)
   - 2.11 [Use chan struct{} for Notification Channels, Not chan bool](#211-use-chan-struct-for-notification-channels-not-chan-bool)
   - 2.12 [Use errgroup for Parallel Goroutines That Need Error Propagation](#212-use-errgroup-for-parallel-goroutines-that-need-error-propagation)
   - 2.13 [Use Purposeful Channel Sizes — Default to 1 for Buffered Channels](#213-use-purposeful-channel-sizes--default-to-1-for-buffered-channels)
   - 2.14 [Use sync.Cond to Broadcast Notifications to Multiple Goroutines](#214-use-synccond-to-broadcast-notifications-to-multiple-goroutines)

---

## 1. Concurrency Foundations

**Impact: CRITICAL**

Go's concurrency primitives are powerful but dangerous when misused. Understanding goroutines vs parallelism, race conditions, channel vs mutex patterns, and context usage is essential. These patterns cover the foundations of safe concurrent code.

### 1.1 Always Handle Context Cancellation

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

### 1.2 Concurrency Is Not Always Faster Than Sequential Code

**Impact: MEDIUM**

Goroutines have overhead: scheduling, context switching, and synchronization. For small workloads, a concurrent solution can be significantly slower than a simple sequential loop. Always benchmark before choosing a concurrent approach; only parallelize when the workload is large enough to offset the overhead.

**Incorrect:**

```go
// Assumes goroutines are always faster — this can be SLOWER for small n
func merge(sequences ...[]int) []int {
    // Spawning goroutines per element is wasteful for small inputs
    resultCh := make(chan []int, len(sequences))
    var wg sync.WaitGroup

    for _, seq := range sequences {
        wg.Add(1)
        go func(s []int) {       // Goroutine overhead: scheduling, stack alloc
            defer wg.Done()
            resultCh <- sortedCopy(s)
        }(seq)
    }

    wg.Wait()
    close(resultCh)
    // Merge results...
    // For 3-4 small slices, this is far slower than just sorting sequentially
}
```

**Correct:**

```go
const minParallelSize = 2048  // Empirical threshold from benchmarking

func merge(sequences ...[]int) []int {
    if totalSize(sequences) < minParallelSize {
        // Sequential: no goroutine overhead for small inputs
        return mergeSequential(sequences)
    }

    // Parallel: worthwhile overhead for large inputs
    resultCh := make(chan []int, len(sequences))
    var wg sync.WaitGroup
    for _, seq := range sequences {
        wg.Add(1)
        go func(s []int) {
            defer wg.Done()
            resultCh <- sortedCopy(s)
        }(seq)
    }
    wg.Wait()
    close(resultCh)
    return mergeFromChannel(resultCh)
}

// Always benchmark to find your threshold:
// BenchmarkSequential-8    50000    23451 ns/op
// BenchmarkConcurrent-8    10000    98234 ns/op  <- slower for small n
// BenchmarkConcurrent-8     5000    18234 ns/op  <- faster for large n
```

Why this matters: Goroutine creation costs time (stack allocation, scheduler registration). For CPU-bound tasks, splitting work across goroutines requires additional synchronization and result merging. The crossover point — where concurrent execution beats sequential — depends on the workload size and computation cost. Always benchmark with realistic data sizes. A common pattern is to run sequentially below a threshold and concurrently above it.

### 1.3 Don't Propagate a Context That May Already Be Canceled

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

### 1.4 Know When to Use Channels Versus Mutexes

**Impact: MEDIUM**

Channels and mutexes are complementary tools. Choosing the wrong one leads to either complex coordination code or unnecessary serialization. Use mutexes when goroutines need to share/access a resource; use channels when goroutines need to communicate and coordinate.

**Decision Guide:**

```go
// Parallel goroutines sharing a resource: use mutex
type Cache struct {
    mu    sync.RWMutex
    items map[string]int
}

func (c *Cache) Get(key string) int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.items[key]
}

func (c *Cache) Set(key string, val int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = val
}

// Concurrent goroutines coordinating work: use channels
func pipeline() {
    results := make(chan int)

    go func() {
        results <- compute()  // G1 signals completion to G2
    }()

    total := 0
    for r := range results {
        total += r
    }
}
```

**Incorrect: using a channel where a mutex is clearer**

**Correct: mutex for shared state, channel for coordination**

Why this matters: Mutexes protect critical sections where multiple goroutines access shared state. Channels communicate between goroutines that need to pass data or signal events. Parallel goroutines (doing the same task independently) need mutexes to synchronize. Concurrent goroutines (doing different tasks in a pipeline) need channels to coordinate. The rule of thumb: "share memory by communicating" means using channels to transfer ownership, not to wrap every shared access.

### 1.5 Set Worker Pool Size Based on Workload Type (CPU vs I/O)

**Impact: MEDIUM**

The optimal goroutine pool size depends on whether the workload is CPU-bound or I/O-bound. Using the wrong pool size leads to underutilization or excessive context switching. CPU-bound workloads should match GOMAXPROCS; I/O-bound workloads can use larger pools.

**Incorrect:**

```go
// Using an arbitrary fixed pool size regardless of workload type
const poolSize = 100  // Why 100? No reasoning — likely wrong for CPU-bound tasks

func processItems(items []Item) {
    ch := make(chan Item, poolSize)
    for i := 0; i < poolSize; i++ {
        go func() {
            for item := range ch {
                cpuIntensiveTask(item)  // CPU-bound: 100 goroutines on 4 cores = thrashing
            }
        }()
    }
    for _, item := range items {
        ch <- item
    }
    close(ch)
}
```

**Correct:**

```go
// CPU-bound: pool size = GOMAXPROCS (number of OS threads)
func processCPUBound(items []Item) {
    n := runtime.GOMAXPROCS(0)  // Returns current value without changing it
    ch := make(chan Item, n)

    var wg sync.WaitGroup
    wg.Add(n)
    for i := 0; i < n; i++ {
        go func() {
            defer wg.Done()
            for item := range ch {
                cpuIntensiveTask(item)  // One goroutine per OS thread = no thrashing
            }
        }()
    }
    for _, item := range items {
        ch <- item
    }
    close(ch)
    wg.Wait()
}

// I/O-bound: pool size depends on external system capacity
func processIOBound(items []Item) {
    // I/O tasks block; while blocked, other goroutines can run
    // Larger pools increase throughput up to the external system's limit
    const poolSize = 50  // Determined by external API rate limits, DB connections, etc.
    ch := make(chan Item, poolSize)

    var wg sync.WaitGroup
    wg.Add(poolSize)
    for i := 0; i < poolSize; i++ {
        go func() {
            defer wg.Done()
            for item := range ch {
                callExternalAPI(item)  // Blocks on I/O, allows other goroutines to run
            }
        }()
    }
    for _, item := range items {
        ch <- item
    }
    close(ch)
    wg.Wait()
}
```

Why this matters: For CPU-bound tasks, more goroutines than `GOMAXPROCS` increases context switching without adding parallelism. For I/O-bound tasks, goroutines block on I/O and the Go scheduler can run others, so a larger pool increases throughput. Use `runtime.GOMAXPROCS(0)` for CPU-bound pools. For I/O-bound, the limit is the external system's capacity (DB connections, API rate limits). Always validate with benchmarks.

### 1.6 Understand the Difference Between Concurrency and Parallelism

**Impact: LOW**

Concurrency and parallelism are distinct concepts often confused. Concurrency is a design approach — structuring a program to handle multiple tasks by interleaving their execution. Parallelism is physical simultaneous execution on multiple CPU cores. A concurrent program may or may not run in parallel.

**Key Distinctions:**

```go
Concurrency                        Parallelism
──────────────────────────────     ──────────────────────────────
About program STRUCTURE            About program EXECUTION
Multiple tasks in progress         Multiple tasks executing
Can run on 1 CPU (time-slicing)    Requires multiple CPUs
Goroutines: concurrent by design   GOMAXPROCS controls parallelism
"Dealing with many things at once" "Doing many things at once"
```

**Example:**

```go
// This is CONCURRENT (goroutines interleave on available cores)
func processConcurrently(tasks []Task) {
    var wg sync.WaitGroup
    for _, task := range tasks {
        wg.Add(1)
        go func(t Task) {  // Concurrent goroutine
            defer wg.Done()
            process(t)
        }(task)
    }
    wg.Wait()
}

// Whether it runs in PARALLEL depends on runtime.GOMAXPROCS
// Default: GOMAXPROCS = number of CPUs, so it IS parallel on multi-core machines
// Single-core machine: concurrent but NOT parallel (interleaved, not simultaneous)

// Explicitly set parallelism (rarely needed — defaults are usually correct):
runtime.GOMAXPROCS(4)  // Use 4 OS threads for parallel execution
```

**Why Concurrency Doesn't Always Mean Faster:**

```go
// Concurrent but NOT faster for CPU-bound tasks with heavy synchronization:
var mu sync.Mutex
var total int

// If every goroutine serializes on mu, concurrency overhead makes it slower
// than sequential execution
for _, n := range numbers {
    go func(v int) {
        mu.Lock()
        total += v       // Heavily contended lock = serialized anyway
        mu.Unlock()
    }(n)
}
```

Why this matters: Confusing concurrency with parallelism leads to incorrect assumptions about performance. A concurrent Go program on a single-core machine interleaves goroutines without true parallelism. Adding goroutines doesn't automatically improve throughput — it depends on workload type (I/O-bound vs CPU-bound), synchronization overhead, and available CPU cores. Use `GOMAXPROCS` (default = nCPUs) and profile before assuming goroutines speed things up.

---

## 2. Concurrency Practice

**Impact: CRITICAL**

Practical concurrency involves goroutine lifecycle management, channel operations, sync package usage, and avoiding common pitfalls. Goroutine leaks, deadlocks, data races, and improper synchronization cause production incidents. These patterns ensure correct concurrent programs.

### 2.1 Assigning a Map or Slice Doesn't Copy the Data — Protect the Whole Operation

**Impact: HIGH**

Assigning a map or slice to a new variable copies the header (pointer + metadata), not the underlying data. Two goroutines accessing a "copy" of a map/slice still access the same backing store. A mutex protecting only the assignment is insufficient — protect the entire read/write operation.

**Incorrect:**

```go
type Cache struct {
    mu       sync.RWMutex
    balances map[string]float64
}

func (c *Cache) AverageBalance() float64 {
    c.mu.RLock()
    balances := c.balances  // BUG: copies the map header, not the data
    c.mu.RUnlock()          // Unlock before iteration

    // DATA RACE: concurrent AddBalance can mutate the same underlying map data
    sum := 0.0
    for _, b := range balances {  // Iterating unprotected shared data
        sum += b
    }
    return sum / float64(len(balances))
}
```

**Correct:**

```go
// Option 1: Keep the lock for the entire operation (best for lightweight iteration)
func (c *Cache) AverageBalance() float64 {
    c.mu.RLock()
    defer c.mu.RUnlock()

    sum := 0.0
    for _, b := range c.balances {  // Protected by the lock the whole time
        sum += b
    }
    return sum / float64(len(c.balances))
}

// Option 2: Make a deep copy, then release the lock (best for heavy operations)
func (c *Cache) AverageBalance() float64 {
    c.mu.RLock()
    m := make(map[string]float64, len(c.balances))
    for k, v := range c.balances {
        m[k] = v  // Deep copy: independent backing store
    }
    c.mu.RUnlock()  // Safe to release now — m is independent

    sum := 0.0
    for _, b := range m {
        sum += b
    }
    return sum / float64(len(m))
}
```

Why this matters: In Go, maps and slices are reference types. `m2 := m1` creates a new header that points to the same underlying data. A mutex around only the assignment protects the header copy, not the data access. Any goroutine reading or iterating the "copy" while another writes to the original causes a data race. Either hold the lock for the entire read operation, or create a true deep copy while holding the lock, then operate on the copy outside the lock.

### 2.2 Call wg.Add() Before Spawning a Goroutine, Not Inside It

**Impact: HIGH**

`sync.WaitGroup.Add(n)` must be called in the parent goroutine before `go func()`. Calling `wg.Add` inside the spawned goroutine creates a race: the parent may call `wg.Wait()` before the goroutine runs and calls `wg.Add`, causing the wait group to reach zero prematurely and unblocking `Wait` too early.

**Incorrect:**

```go
wg := sync.WaitGroup{}
var v uint64

for i := 0; i < 3; i++ {
    go func() {
        wg.Add(1)              // BUG: called inside goroutine — race with wg.Wait()
        defer wg.Done()
        atomic.AddUint64(&v, 1)
    }()
}

wg.Wait()              // May return before all goroutines run
fmt.Println(v)         // Non-deterministic: could print 0, 1, 2, or 3
```

**Correct:**

```go
wg := sync.WaitGroup{}
var v uint64

// Option 1: Add total count before the loop (when count is known)
wg.Add(3)
for i := 0; i < 3; i++ {
    go func() {
        defer wg.Done()
        atomic.AddUint64(&v, 1)
    }()
}

// Option 2: Add(1) per iteration, BEFORE spawning the goroutine
for i := 0; i < 3; i++ {
    wg.Add(1)           // Called in parent goroutine — guaranteed before go func()
    go func() {
        defer wg.Done()
        atomic.AddUint64(&v, 1)
    }()
}

wg.Wait()
fmt.Println(v)  // Guaranteed to print 3
```

Why this matters: `wg.Wait()` blocks until the internal counter reaches zero. If `wg.Add` is called inside goroutines, the counter may still be zero when `Wait` is called — before any goroutine has a chance to increment it. The parent goroutine unblocks immediately and reads a partially-updated result. Always call `wg.Add` in the parent before launching the goroutine. Use `defer wg.Done()` inside the goroutine to ensure it always decrements, even on early return.

### 2.3 Concurrent append on a Shared Slice Can Cause a Data Race

**Impact: HIGH**

`append` is not thread-safe. If a slice has spare capacity (`len < cap`), two goroutines appending concurrently write to the same backing array, causing a data race. If the slice is full (`len == cap`), each `append` allocates a new backing array, so there's no race — but this is fragile and depends on initialization.

**Incorrect:**

```go
// Slice with spare capacity — data race!
s := make([]int, 0, 1)  // len=0, cap=1: has spare capacity

go func() {
    s1 := append(s, 1)  // Both goroutines write to index 0 of the same backing array
    fmt.Println(s1)
}()

go func() {
    s2 := append(s, 1)  // DATA RACE: concurrent write to same memory
    fmt.Println(s2)
}()

// Coincidentally safe (but fragile): slice at full capacity
s := make([]int, 1, 1)  // len=1, cap=1: full capacity
// Each append allocates a new backing array — no race, but only by accident
```

**Correct:**

```go
// Each goroutine works on its own copy
s := make([]int, 0, 1)

go func() {
    sCopy := make([]int, len(s), cap(s))
    copy(sCopy, s)
    s1 := append(sCopy, 1)  // Appends to its own copy — no race
    fmt.Println(s1)
}()

go func() {
    sCopy := make([]int, len(s), cap(s))
    copy(sCopy, s)
    s2 := append(sCopy, 1)  // Independent copy
    fmt.Println(s2)
}()

// Or: protect with a mutex if goroutines must share the slice
var mu sync.Mutex
var shared []int

go func() {
    mu.Lock()
    shared = append(shared, 1)
    mu.Unlock()
}()
```

Why this matters: The behavior of `append` depends on whether the slice is full. If `len < cap`, `append` writes into the existing backing array without allocating — making concurrent appends a data race. If `len == cap`, `append` allocates a new array, so goroutines don't share memory. Never rely on this coincidence. For concurrent slice modification, either give each goroutine its own copy, use a mutex, or design the algorithm so goroutines write to different indices (guaranteed non-overlapping).

### 2.4 Goroutine Closures Capture Loop Variables by Reference

**Impact: HIGH**

A goroutine closure that references a loop variable captures the variable itself, not its value at the time the goroutine is created. By the time the goroutine runs, the loop variable may have advanced to a later value. This is one of the most common Go concurrency bugs.

**Incorrect:**

```go
s := []int{1, 2, 3}

for _, i := range s {
    go func() {
        fmt.Print(i)  // BUG: captures the variable i, not its value
    }()
}
// Expected: prints 1 2 3 (in some order)
// Actual: may print 2 3 3, or 3 3 3, or other combinations
// All goroutines share the same i variable; most run after the loop ends
```

**Correct:**

```go
// Option 1: Create a local copy inside the loop body
for _, i := range s {
    val := i           // New variable created each iteration
    go func() {
        fmt.Print(val) // Captures val, which is fixed per iteration
    }()
}

// Option 2: Pass the value as a function argument (not a closure)
for _, i := range s {
    go func(val int) {
        fmt.Print(val) // val is a function parameter — not a captured variable
    }(i)               // i is evaluated and passed NOW
}

// Note: Go 1.22+ changed loop variable semantics — each iteration
// gets a new variable, so the bug no longer occurs in Go 1.22+
// But for compatibility and clarity, either option above is still recommended
```

Why this matters: Closures in Go capture variables by reference, not by value. In a goroutine closure, the captured variable `i` is the same memory location used by the loop — it changes each iteration. When the goroutines execute (usually after the loop), they all read the current value of `i`, which is typically the last iteration's value or beyond. This produces non-deterministic output. The fix is to either shadow the variable locally (`val := i`) or pass it as a function argument.

### 2.5 Never Copy sync Types (Mutex, WaitGroup, Cond)

**Impact: CRITICAL**

`sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, and `sync.Cond` must never be copied after first use. Copying a mutex copies its internal state — including the locked/unlocked status — causing two goroutines to operate on separate, desynchronized mutexes. This breaks mutual exclusion entirely.

**Incorrect:**

```go
type Counter struct {
    mu       sync.Mutex  // sync type embedded by value
    counters map[string]int
}

// Value receiver: COPIES the Counter struct, including mu
func (c Counter) Increment(name string) {
    c.mu.Lock()           // BUG: locking a COPY of mu, not the original
    defer c.mu.Unlock()
    c.counters[name]++    // Two goroutines may both hold their own "locked" copy
}

// Passing by value also copies:
func process(c Counter) { ... }  // BUG: c is a copy with a copied mutex

// Assigning also copies:
c2 := c  // BUG: c2.mu is a copy of c.mu's state
```

**Correct:**

```go
type Counter struct {
    mu       sync.Mutex
    counters map[string]int
}

// Pointer receiver: operates on the original struct, no copy
func (c *Counter) Increment(name string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.counters[name]++
}

// Pass pointer:
func process(c *Counter) { ... }

// The go vet tool detects sync copies:
// go vet ./...
// suspicious assignment copies lock value to c2: type Counter contains sync.Mutex

// When embedding sync types, always embed the pointer:
type SafeMap struct {
    mu *sync.Mutex  // Pointer: embedding by pointer avoids copy issues
    m  map[string]int
}
```

Why this matters: A mutex is a lock by containing internal state (often an integer). When you copy a `sync.Mutex`, both the original and the copy have independent states. If the original is locked, the copy appears unlocked. Two goroutines can each call `Lock()` on what they think is the same mutex but are actually operating on separate copies — mutual exclusion is completely broken. The `go vet` tool reports sync-copy violations. Always use pointer receivers for types containing sync primitives.

### 2.6 Prevent Goroutine Leaks with Context or Channels

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

### 2.7 Protect Shared State with Mutexes or Channels

**Impact: CRITICAL**

Accessing shared variables from multiple goroutines without synchronization causes data races—undefined behavior that can corrupt memory and cause crashes. Use `sync.Mutex` for shared state or communicate via channels.

**Incorrect:**

```go
type Counter struct {
    count int
}

func (c *Counter) Increment() {
    c.count++  // Race condition!
}

func main() {
    counter := &Counter{}
    for i := 0; i < 1000; i++ {
        go counter.Increment()
    }
}
```

**Correct: using Mutex**

```go
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

**Correct: using Channels**

```go
type Counter struct {
    ops chan func(*int)
}

func NewCounter() *Counter {
    c := &Counter{ops: make(chan func(*int))}
    go func() {
        count := 0
        for op := range c.ops {
            op(&count)
        }
    }()
    return c
}

func (c *Counter) Increment() {
    c.ops <- func(count *int) { *count++ }
}
```

Why this matters: Data races are unpredictable—they may work fine in testing but fail randomly in production. The race detector (`go test -race`) catches these, but prevention is better. Choose mutexes for simple shared state, channels for communication and coordination.

Rule of thumb: "Don't communicate by sharing memory; share memory by communicating" (via channels). When you must share, protect with mutexes.

### 2.8 select With Multiple Ready Channels Chooses Randomly, Not by Order

**Impact: MEDIUM**

When multiple `case` clauses in a `select` statement are ready simultaneously, Go chooses one **at random** — not in source order. Code that assumes the first `case` has priority will have non-deterministic behavior. Implement explicit prioritization when channel ordering matters.

**Incorrect:**

```go
// WRONG assumption: messageCh case runs first because it appears first
for {
    select {
    case v := <-messageCh:    // Assumed to have priority — it doesn't
        fmt.Println(v)
    case <-disconnectCh:
        fmt.Println("disconnected")
        return
    }
}
// If both channels have data, Go may pick disconnectCh first,
// causing some messages to be missed
```

**Correct:**

```go
// For a single producer: use inner select + default to drain messageCh first
for {
    select {
    case v := <-messageCh:
        fmt.Println(v)
    case <-disconnectCh:
        // Drain remaining messages before returning
        for {
            select {
            case v := <-messageCh:
                fmt.Println(v)
            default:
                fmt.Println("disconnected")
                return  // default fires when messageCh is empty
            }
        }
    }
}

// Alternative: use nil channels to remove a case once its source is done
func merge(ch1, ch2 <-chan int) <-chan int {
    ch := make(chan int, 1)
    go func() {
        for ch1 != nil || ch2 != nil {
            select {
            case v, open := <-ch1:
                if !open { ch1 = nil; break }  // Remove ch1 from select
                ch <- v
            case v, open := <-ch2:
                if !open { ch2 = nil; break }  // Remove ch2 from select
                ch <- v
            }
        }
        close(ch)
    }()
    return ch
}
```

Why this matters: The Go specification states that when multiple `select` cases are ready, one is chosen via uniform pseudo-random selection. This prevents starvation of a slow channel by a fast one, but it breaks any assumption about ordering. The `default` case inside an inner `for/select` provides a way to drain one channel before returning. Setting a channel to `nil` elegantly removes it from a `select` when it's closed, since receiving from a nil channel blocks forever.

### 2.9 String Formatting Can Cause Data Races or Deadlocks in Concurrent Code

**Impact: HIGH**

`fmt.Sprintf` and related functions call the `String()` method on values. If a `String()` method acquires a mutex that is already held in the calling goroutine, it causes a deadlock. If formatting traverses mutable shared data (like a `context.WithValue` chain), it can cause a data race.

**Incorrect: deadlock**

```go
type Customer struct {
    mu  sync.RWMutex
    id  string
    age int
}

func (c *Customer) UpdateAge(age int) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if age < 0 {
        // BUG: %v calls c.String(), which tries to acquire c.mu.RLock()
        // But c.mu is already write-locked — DEADLOCK
        return fmt.Errorf("age should be positive for customer %v", c)
    }
    c.age = age
    return nil
}

func (c *Customer) String() string {
    c.mu.RLock()           // Tries to lock an already write-locked mutex
    defer c.mu.RUnlock()
    return fmt.Sprintf("id %s, age %d", c.id, c.age)
}
```

**Correct:**

```go
func (c *Customer) UpdateAge(age int) error {
    // Option 1: Validate BEFORE acquiring the lock
    if age < 0 {
        // c.mu is not locked here — String() can acquire RLock safely
        return fmt.Errorf("age should be positive for customer %v", c)
    }

    c.mu.Lock()
    defer c.mu.Unlock()
    c.age = age
    return nil
}

// Option 2: Format specific fields directly, don't call String()
func (c *Customer) UpdateAge(age int) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if age < 0 {
        return fmt.Errorf("age should be positive for customer id %s", c.id)
        // Accessing c.id directly — no String() call, no deadlock
    }
    c.age = age
    return nil
}
```

Why this matters: The `%v` and `%s` format directives call the `String()` method if implemented. If `String()` tries to acquire a lock that the formatting call's goroutine already holds, the goroutine deadlocks. Additionally, `fmt.Sprintf("%v", ctx)` traverses the context chain, which may access mutable values across goroutines — a potential data race. Be especially careful when formatting structs inside locked critical sections.

### 2.10 Understand Nil Channel Behavior

**Impact: MEDIUM**

Nil channels block forever on send and receive operations. This is useful for disabling cases in select statements but causes deadlocks if misunderstood.

**Incorrect:**

```go
func merge(ch1, ch2 <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for {
            select {
            case v := <-ch1:  // If ch1 is nil, blocks forever!
                out <- v
            case v := <-ch2:
                out <- v
            }
        }
    }()
    return out
}
```

**Correct:**

```go
func merge(ch1, ch2 <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for ch1 != nil || ch2 != nil {
            select {
            case v, ok := <-ch1:
                if !ok {
                    ch1 = nil  // Disable this case
                    continue
                }
                out <- v
            case v, ok := <-ch2:
                if !ok {
                    ch2 = nil  // Disable this case
                    continue
                }
                out <- v
            }
        }
    }()
    return out
}
```

Why this matters: Setting a channel to `nil` makes select ignore that case, allowing graceful shutdown when one channel closes. Without this, select would continuously receive zero values from the closed channel.

### 2.11 Use chan struct{} for Notification Channels, Not chan bool

**Impact: LOW**

When a channel carries no meaningful data — only the fact that a signal was sent — use `chan struct{}` (empty struct), not `chan bool` or `chan int`. This makes the intent clear and uses zero memory.

**Incorrect:**

```go
// Using bool is ambiguous: what does false mean?
disconnectCh := make(chan bool)

// Receiver must wonder: does true mean "disconnected" and false mean "reconnected"?
// What if false is never sent? Should I expect it?
case connected := <-disconnectCh:
    if !connected { ... }  // Confusing

// Using int is even worse — what do the values mean?
doneCh := make(chan int)
```

**Correct:**

```go
// chan struct{} clearly signals "an event occurred" with no ambiguity about value
disconnectCh := make(chan struct{})

// Sending a notification:
disconnectCh <- struct{}{}

// Closing to broadcast to all receivers (most common for done signals):
close(disconnectCh)

// Receiver only cares that the channel fired, not what value it holds
case <-disconnectCh:
    fmt.Println("disconnected")  // No confusion about what the value means

// Examples from stdlib using this pattern:
// context.Context.Done() returns <-chan struct{}
// sync.WaitGroup uses this concept internally
// time.After returns <-chan time.Time, but done/quit channels use chan struct{}

// Empty struct costs zero bytes:
var s struct{}
fmt.Println(unsafe.Sizeof(s))  // 0

// Contrast with bool: 1 byte; interface{}: 16 bytes
```

Why this matters: `chan bool` implies the value of the boolean matters to the receiver, creating confusion about what `true` vs `false` conveys. `chan struct{}` is the idiomatic Go way to express "signal only, no data." It's zero-sized, so sending a `struct{}{}` is cheap. The pattern appears throughout the standard library (e.g., `context.Done()` channels). Use it for done signals, quit signals, and any channel where the fact of receipt is the message.

### 2.12 Use errgroup for Parallel Goroutines That Need Error Propagation

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

### 2.13 Use Purposeful Channel Sizes — Default to 1 for Buffered Channels

**Impact: LOW**

Choosing a buffered channel size arbitrarily is a common mistake. The size affects backpressure, memory usage, and synchronization semantics. When in doubt, start with a size of 1 or use an unbuffered channel, and use other sizes only when there's a specific, documented reason.

**Incorrect:**

```go
// Magic number with no justification — why 40? Why not 50 or 1000?
ch := make(chan int, 40)

// Unbuffered channel when decoupling sender/receiver is needed:
ch := make(chan int)  // Sender blocks until receiver is ready — may not be intended
```

**Correct:**

```go
// Unbuffered: provides synchronization — sender blocks until receiver is ready
// Use when: you need guaranteed delivery or want to know when work was received
ch := make(chan int)

// Buffered size 1: minimal decoupling — allows sender to proceed without waiting
// Use as the default buffered channel size when unsure
ch := make(chan int, 1)

// Buffered size = pool size: for worker pool pattern
poolSize := runtime.GOMAXPROCS(0)
taskCh := make(chan Task, poolSize)  // Tied to the number of workers

// Buffered size = rate limit: for rate-limiting scenarios
const maxConcurrentRequests = 10
semaphore := make(chan struct{}, maxConcurrentRequests)

// Document any other size with a comment explaining the rationale
// Size 256: empirically determined from benchmark on production workload
ch := make(chan Event, 256)
```

Why this matters: An unbuffered channel provides synchronization guarantees (sender blocks until receiver receives). A buffered channel decouples sender and receiver but can lead to obscure deadlocks if the buffer fills and no one is reading. The minimum useful buffer size is 1. Larger sizes should be tied to concrete values (worker pool size, rate limits) or determined via benchmarks. Magic numbers like `make(chan int, 40)` are a code smell — always comment the rationale.

### 2.14 Use sync.Cond to Broadcast Notifications to Multiple Goroutines

**Impact: LOW**

When multiple goroutines need to wait for the same repeating condition, `sync.Cond` is the right tool. A channel can only deliver a message to one goroutine at a time; only a channel closure broadcasts to all, but closing is a one-shot action. `sync.Cond.Broadcast()` wakes all waiting goroutines each time a condition changes.

**Incorrect:**

```go
// Busy loop wastes CPU — checking condition repeatedly
for donation.balance < goal {
    // spinning without sleeping burns CPU at 100%
}

// Channel approach: each message goes to ONE goroutine (round-robin)
// Multiple listeners miss notifications
ch <- balance  // Only one listener receives each update
```

**Correct:**

```go
type Donation struct {
    cond    *sync.Cond
    balance int
}

donation := &Donation{
    cond: sync.NewCond(&sync.Mutex{}),
}

// Listener goroutines — wait for condition to be met
go func(goal int) {
    donation.cond.L.Lock()
    for donation.balance < goal {
        donation.cond.Wait()  // Atomically: unlock, suspend, re-lock when woken
    }
    fmt.Printf("$%d goal reached\n", donation.balance)
    donation.cond.L.Unlock()
}(10)

go func(goal int) {
    donation.cond.L.Lock()
    for donation.balance < goal {
        donation.cond.Wait()
    }
    fmt.Printf("$%d goal reached\n", donation.balance)
    donation.cond.L.Unlock()
}(15)

// Updater goroutine — broadcasts every time the condition changes
for {
    time.Sleep(time.Second)
    donation.cond.L.Lock()
    donation.balance++
    donation.cond.L.Unlock()
    donation.cond.Broadcast()  // Wakes ALL waiting goroutines to re-check condition
}
```

Why this matters: `sync.Cond.Wait()` atomically releases the lock and suspends the goroutine, then re-acquires the lock when woken — no busy loop, no wasted CPU. `Broadcast()` wakes all goroutines waiting on the condition; `Signal()` wakes one. Always check the condition in a `for` loop (not `if`) after `Wait` returns, because spurious wakeups can occur. Use `sync.Cond` when multiple goroutines need repeated broadcast notifications about a shared state change.

---

## References

1. [https://go.dev/doc/effective_go](https://go.dev/doc/effective_go)
2. [https://go.dev/ref/mem](https://go.dev/ref/mem)
3. [https://github.com/golang/go/wiki/CommonMistakes](https://github.com/golang/go/wiki/CommonMistakes)
