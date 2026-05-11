## Call wg.Add() Before Spawning a Goroutine, Not Inside It

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
