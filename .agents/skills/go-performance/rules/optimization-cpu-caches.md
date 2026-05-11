## Design Data Structures for CPU Cache Efficiency

**Impact: MEDIUM**

Modern CPUs use multi-level caches (L1/L2/L3) that are orders of magnitude faster than RAM. Code that accesses memory in predictable, contiguous patterns gets most data from cache (fast); code that jumps around memory gets frequent cache misses (slow). Understanding this allows data structure choices that can yield 20-70% performance improvements.

**Cache basics:**
- L1: ~1 ns, L2: ~4 ns, L3: ~10 ns, RAM: ~50-100 ns
- A **cache line** is 64 bytes (8 `int64` values) — the unit of data the CPU moves between RAM and cache
- **Spatial locality**: accessing `s[0]` also caches `s[1]`...`s[7]` in the same line — subsequent accesses are free
- **Temporal locality**: frequently accessed variables stay in cache

**Slice of structs vs. struct of slices:**

```go
// Slice of structs — iterating over field 'a' skips 'b' each time (poor spatial locality)
type Foo struct {
    a int64
    b int64  // Loaded into cache but never used when summing 'a'
}
func sumFoo(foos []Foo) int64 {
    var total int64
    for _, f := range foos {
        total += f.a  // Every other int64 in the cache line is wasted
    }
    return total
}

// Struct of slices — all 'a' values are contiguous in memory (good spatial locality)
type Bar struct {
    a []int64
    b []int64  // Separate slice — not loaded when iterating 'a'
}
func sumBar(bar Bar) int64 {
    var total int64
    for _, v := range bar.a {  // Iterates a densely packed slice — twice as few cache lines
        total += v
    }
    return total
}
// sumBar is ~20% faster because it fetches fewer cache lines
```

**Linked lists vs. slices — predictability matters:**

```go
// Linked list — non-unit stride: the CPU can't predict where next node is in memory
type node struct {
    value int64
    next  *node  // Pointer to anywhere in heap — unpredictable
}
// CPU can't prefetch; every access may be a cache miss

// Slice — unit stride: elements are contiguous; CPU prefetches ahead
func sumSlice(s []int64) int64 {
    var total int64
    for _, v := range s {  // Predictable: each element follows the last
        total += v
    }
    return total
}
// Slice iteration is ~70% faster than linked list iteration (even with same spatial locality)
```

**Critical stride — avoid power-of-2 sized rows in matrices:**

```go
// 512 columns: rows land on the same cache set → conflict misses
type matrix512 [N][512]int64  // 512 * 8 = 4096 bytes = power of 2 → cache conflicts

// 513 columns: rows land on different cache sets → no conflicts
type matrix513 [N][513]int64  // 513 * 8 = 4104 bytes → avoids critical stride

// When reusing the same matrix in a benchmark, matrix512 can be ~50% SLOWER
// because all rows compete for the same cache set
// Fix for benchmarks: create a new matrix each iteration (b.StopTimer/b.StartTimer)
```

**Guidelines:**
1. Prefer slices over linked lists for sequential access patterns
2. When only one field of a struct is used in a hot loop, consider struct-of-slices layout
3. Avoid matrix dimensions that are exact powers of 2 (critical stride)
4. Profile before optimizing — use `go test -bench`, `pprof`, and `perf`

Why this matters: Cache misses are 50-100x more expensive than cache hits. A function that accesses memory in a predictable, contiguous pattern will be dramatically faster than one that jumps around, even if both have identical algorithmic complexity. These are not micro-optimizations — they can represent 20-70% real-world differences in hot code paths.
