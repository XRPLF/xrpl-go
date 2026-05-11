## Set Worker Pool Size Based on Workload Type (CPU vs I/O)

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
