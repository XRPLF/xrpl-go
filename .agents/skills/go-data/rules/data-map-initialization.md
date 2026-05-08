## Initialize Maps with Expected Size

**Impact: HIGH**

Like slices, maps can be pre-sized at creation time. When a map grows beyond its load factor (6.5 elements per bucket), Go allocates new buckets and rehashes all keys — an O(n) operation. Providing an initial size hint avoids repeated growth operations and is significantly faster.

**Incorrect:**

```go
// Map starts with 1 bucket and grows repeatedly as elements are added
func buildIndex(words []string) map[string]int {
    index := make(map[string]int)  // No size hint
    for i, word := range words {
        index[word] = i
    }
    return index
}

// Benchmark: inserting 1 million elements
// Without size: ~227 ms/op (repeated bucket allocations and rehashing)
```

**Correct:**

```go
// Provide the expected number of elements as a size hint
func buildIndex(words []string) map[string]int {
    index := make(map[string]int, len(words))  // Pre-size with expected count
    for i, word := range words {
        index[word] = i
    }
    return index
}

// Benchmark: inserting 1 million elements
// With size: ~91 ms/op — about 60% faster

// Note: the size hint is not a maximum — you can always add more elements
// The hint just tells the runtime how many buckets to pre-allocate
```

Why this matters: A map grows by doubling its bucket count when the average bucket load exceeds ~6.5. Each growth requires rehashing all existing keys. For large maps, this happens logarithmically many times during construction, causing significant allocation overhead. Providing a size hint with `make(map[K]V, n)` pre-allocates enough buckets, reducing or eliminating growth operations.

Unlike slices, maps only accept a single size argument (no separate capacity). The size is a hint — if you provide `n`, Go allocates enough buckets to hold roughly `n` elements without growing.
