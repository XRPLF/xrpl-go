## Understand the Difference Between Concurrency and Parallelism

**Impact: LOW**

Concurrency and parallelism are distinct concepts often confused. Concurrency is a design approach — structuring a program to handle multiple tasks by interleaving their execution. Parallelism is physical simultaneous execution on multiple CPU cores. A concurrent program may or may not run in parallel.

**Key Distinctions:**

```
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
