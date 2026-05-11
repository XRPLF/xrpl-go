## Concurrency Is Not Always Faster Than Sequential Code

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
