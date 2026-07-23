## Reduce Heap Allocations With sync.Pool and API Design

**Impact: MEDIUM**

Every heap allocation pressures the GC. Three practical techniques to reduce allocations: (1) design APIs to accept caller-provided buffers (sharing down); (2) rely on compiler optimizations like `string(bytes)` in map lookups; (3) use `sync.Pool` to reuse frequently allocated objects.

**Technique 1: API design — let the caller provide the buffer (sharing down):**

```go
// BAD: Returns a new []byte on every call — caller cannot reuse it
func getResponse() []byte {
    return make([]byte, 1024)
}

// GOOD: Accept a caller-provided slice — no new allocation
func getResponse(buf []byte) []byte {
    // Write into buf instead of allocating a new one
    return buf[:len(result)]
}
```

**Technique 2: Compiler optimization — string(bytes) map lookup avoids allocation:**

```go
// Allocates a new string on every call (key variable assignment)
func (c *cache) get(bytes []byte) (int, bool) {
    key := string(bytes)     // Creates a new string allocation
    v, ok := c.m[key]
    return v, ok
}

// FASTER: compiler elides the allocation when string(bytes) is used directly as a map key
func (c *cache) get(bytes []byte) (int, bool) {
    v, ok := c.m[string(bytes)]  // No allocation — compiler special-cases this pattern
    return v, ok
}
```

**Technique 3: sync.Pool — reuse frequently allocated objects:**

```go
// Without pool: allocates a new []byte on every write call
func write(w io.Writer) {
    b := getResponse()       // New allocation every time
    _, _ = w.Write(b)
}

// With sync.Pool: reuses existing allocations
var pool = sync.Pool{
    New: func() any {
        return make([]byte, 1024)  // Factory: called only when pool is empty
    },
}

func write(w io.Writer) {
    buffer := pool.Get().([]byte)  // Get from pool (or create new if empty)
    buffer = buffer[:0]            // Reset the slice (length to 0, capacity unchanged)
    defer pool.Put(buffer)         // Return to pool when done

    getResponse(buffer)
    _, _ = w.Write(buffer)
}
// Pool objects are destroyed after each GC — amortized allocation cost is reduced
```

**sync.Pool rules:**
- Pool is safe for concurrent use by multiple goroutines
- Objects are cleared after each GC cycle (no fixed size or capacity)
- Always reset pooled objects before use (`buffer[:0]`, zeroing fields, etc.)
- Use for objects that are frequently allocated and discarded, not for long-lived state
- The GC will drain the pool; don't rely on objects persisting across GC cycles

Why this matters: Heap allocations are individually inexpensive but cumulative: millions of small allocations fill the heap, triggering frequent GCs that can use 25% of CPU. `sync.Pool` amortizes allocation costs across many callers by reusing objects. The compiler's `string(bytes)` map optimization is a free win — always use it directly instead of assigning to a variable first. For API design, passing a buffer in (sharing down) keeps allocations in the caller's control and often allows stack allocation.
