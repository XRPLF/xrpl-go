## Understand That Maps Never Shrink

**Impact: HIGH**

Go maps can only grow — they never shrink their bucket count. When you add 1 million elements and then delete them all, the map retains all its allocated buckets. This causes permanent high memory consumption in applications with spiky load patterns.

**Incorrect:**

```go
// Cache that grows during peak traffic and never releases memory
type SessionCache struct {
    mu      sync.Mutex
    sessions map[int][128]byte
}

func (c *SessionCache) Add(id int, data [128]byte) {
    c.mu.Lock()
    c.sessions[id] = data
    c.mu.Unlock()
}

func (c *SessionCache) Remove(id int) {
    c.mu.Lock()
    delete(c.sessions, id)  // Frees the value but NOT the bucket
    c.mu.Unlock()
}
// After Black Friday peak: map held 2M entries (461 MB)
// After removing all entries and GC: still 293 MB allocated (buckets remain!)
```

**Correct:**

```go
// Solution 1: Periodically recreate the map to release buckets
func (c *SessionCache) Compact() {
    c.mu.Lock()
    defer c.mu.Unlock()
    newMap := make(map[int][128]byte, len(c.sessions))
    for k, v := range c.sessions {
        newMap[k] = v
    }
    c.sessions = newMap
    // Old map is now GC-eligible; bucket count matches current size
}

// Solution 2: Store pointers instead of values
// map[int]*[128]byte uses much less bucket space when values are large
// After removing all entries: only pointer-sized slots remain in buckets
type SessionCache struct {
    mu       sync.Mutex
    sessions map[int]*[128]byte
}

// Comparison for 1M elements then GC:
// map[int][128]byte:  add=461MB, remove+GC=293MB (still large)
// map[int]*[128]byte: add=182MB, remove+GC=38MB  (buckets freed)
```

Why this matters: Go's map implementation uses a `B` field tracking the number of buckets as a power of 2. After adding 1M elements, `B=18` (262,144 buckets). Deleting all elements zeroes the bucket slots but `B` stays at 18. The bucket array itself is never freed. For maps with large values, storing pointers reduces both peak memory and post-deletion memory significantly.
