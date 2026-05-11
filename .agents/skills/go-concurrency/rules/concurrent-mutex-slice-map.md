## Assigning a Map or Slice Doesn't Copy the Data — Protect the Whole Operation

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
