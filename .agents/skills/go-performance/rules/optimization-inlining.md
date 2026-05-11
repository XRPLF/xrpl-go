## Use Fast-Path Inlining to Optimize Hot Code Paths

**Impact: LOW**

Go's compiler automatically inlines small functions (within the *inlining budget*). Understanding inlining lets you structure code so the common fast path gets inlined while complex slow paths do not. This technique was used in Go's own `sync.Mutex` implementation to improve mutex acquisition by ~5%.

**How inlining works:**

```go
// Simple function — compiler will inline this (cost < budget)
func sum(a, b int) int {
    return a + b
}

func main() {
    s := sum(3, 2)  // Compiler replaces with: s := 3 + 2
    println(s)
}

// Check inlining decisions:
// $ go build -gcflags "-m=2" ./...
// ./main.go: can inline sum with cost 4 as: func(int, int) int { return a + b }
// ./main.go: inlining call to sum func(int, int) int { return a + b }
//
// If too complex:
// ./main.go: cannot inline foo: function too complex: cost 84 exceeds budget 80
```

**Fast-path inlining — extract slow path into a separate function:**

```go
// BEFORE: The whole Lock function is too complex to inline
func (m *Mutex) Lock() {
    if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
        // Fast path: mutex is unlocked — common case
        return
    }
    // Slow path: mutex is already locked — complex logic
    var waitStartTime int64
    starving := false
    // ... many more lines ...
}
// Cannot inline → every Lock() call has function call overhead

// AFTER: Extract slow path → fast path becomes inlinable
func (m *Mutex) Lock() {
    if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
        return  // Fast path — only these few lines, within inlining budget
    }
    m.lockSlow()  // Slow path — separate function, not inlined
}

func (m *Mutex) lockSlow() {
    var waitStartTime int64
    starving := false
    // ... complex logic ...
}
// Lock() is now inlinable: if mutex is unlocked (common case), zero function call overhead
// ~5% speed improvement for uncontended mutex acquisition
```

**Benefits of inlining:**
1. Removes function call overhead (save/restore registers, stack frame setup)
2. Enables further compiler optimizations — a variable that would escape to heap via a function call may stay on the stack after inlining

**When to apply fast-path inlining:**
- Profile first: identify hot functions with simple fast paths and complex slow paths
- The fast path is the common case (e.g., cache hit, uncontended lock, successful validation)
- The slow path is rare (e.g., cache miss, contended lock, error handling)
- Verify with `go build -gcflags "-m=2"` that the inlining actually happens

Why this matters: Function call overhead is small in absolute terms (~1 ns), but in hot loops called millions of times per second it accumulates. More importantly, inlining enables subsequent compiler optimizations like escape analysis improvements, constant folding, and dead code elimination. The fast-path inlining pattern — where the common path is a few lines calling into a slow path function — is the same technique Go's standard library uses.
