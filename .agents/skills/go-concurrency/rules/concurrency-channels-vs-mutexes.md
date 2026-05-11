## Know When to Use Channels Versus Mutexes

**Impact: MEDIUM**

Channels and mutexes are complementary tools. Choosing the wrong one leads to either complex coordination code or unnecessary serialization. Use mutexes when goroutines need to share/access a resource; use channels when goroutines need to communicate and coordinate.

**Decision Guide:**

```
Goroutine Relationship    Problem Type           Use
──────────────────────    ───────────────────    ──────────
Parallel goroutines       Shared resource        Mutex
                          (access/mutation)
Concurrent goroutines     Communication          Channel
                          Ownership transfer
                          Coordination/signal
```

**Incorrect: using a channel where a mutex is clearer**

```go
// Shared cache accessed by parallel goroutines — channel is awkward here
type Cache struct {
    ch chan map[string]int  // Using channel to "protect" a map
}

func (c *Cache) Get(key string) int {
    m := <-c.ch           // Take ownership
    val := m[key]
    c.ch <- m             // Return ownership — verbose and error-prone
    return val
}
```

**Correct: mutex for shared state, channel for coordination**

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

Why this matters: Mutexes protect critical sections where multiple goroutines access shared state. Channels communicate between goroutines that need to pass data or signal events. Parallel goroutines (doing the same task independently) need mutexes to synchronize. Concurrent goroutines (doing different tasks in a pipeline) need channels to coordinate. The rule of thumb: "share memory by communicating" means using channels to transfer ownership, not to wrap every shared access.
