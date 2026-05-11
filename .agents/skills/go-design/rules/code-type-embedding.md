## Be Careful with Type Embedding

**Impact: MEDIUM**

Type embedding promotes all methods of the embedded type to the outer struct. This is powerful but dangerous when you embed types that expose behavior you don't want to be part of your API. Accidentally exposing internal concurrency primitives is the most common mistake.

**Incorrect:**

```go
type InMemoryCache struct {
    sync.Mutex  // Exposes Lock() and Unlock() as public methods!
    items map[string]string
}

func NewInMemoryCache() *InMemoryCache {
    return &InMemoryCache{items: make(map[string]string)}
}

// Callers can now call cache.Lock() directly — this breaks encapsulation
// and allows callers to deadlock your cache
cache := NewInMemoryCache()
cache.Lock()   // Exposed through embedding — this is wrong
cache.Set("key", "value")
```

**Correct:**

```go
// Option 1: Use a named field (composition, not embedding)
type InMemoryCache struct {
    mu    sync.Mutex  // Unexported field — not promoted
    items map[string]string
}

func (c *InMemoryCache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = value
}

// Option 2: Embed only when you want to expose the embedded type's API
type ReadOnlyDB struct {
    *sql.DB  // Intentionally exposing DB's query methods
}

// Override to block writes
func (r *ReadOnlyDB) Exec(query string, args ...any) (sql.Result, error) {
    return nil, errors.New("read-only database")
}
```

Why this matters: Embedding `sync.Mutex` exposes `Lock()` and `Unlock()` as public methods on your struct. This lets callers hold the lock externally and cause deadlocks. The rule: embed when you want to expose, use a named (unexported) field when you want to encapsulate. Ask: "Do I want callers to use the embedded type's methods directly?"
