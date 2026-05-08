## Prevent False Sharing in Concurrent Code With Padding or Local Variables

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
