## Never Copy sync Types (Mutex, WaitGroup, Cond)

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
