## Protect Shared State with Mutexes or Channels

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

**Correct (using Mutex):**

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

**Correct (using Channels):**

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
