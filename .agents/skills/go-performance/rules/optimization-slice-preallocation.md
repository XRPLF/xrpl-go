## Preallocate Slices When Size is Known

**Impact: HIGH**

Appending to a slice without preallocation causes multiple memory allocations and copies as the slice grows. When you know the final size, preallocate with `make()` to avoid this overhead.

**Incorrect:**

```go
func ProcessItems(items []Item) []Result {
    var results []Result  // Starts with capacity 0
    for _, item := range items {
        result := process(item)
        results = append(results, result)  // May reallocate multiple times
    }
    return results
}
```

**Correct:**

```go
func ProcessItems(items []Item) []Result {
    results := make([]Result, 0, len(items))  // Preallocate capacity
    for _, item := range items {
        result := process(item)
        results = append(results, result)  // No reallocation needed
    }
    return results
}
```

Why this matters: Each time a slice grows beyond capacity, Go allocates a new larger array and copies all existing elements. For 1000 items, the incorrect version may allocate 10+ times. The correct version allocates once, saving both CPU time and reducing garbage collection pressure.

Benchmark impact: For large slices (10,000+ elements), preallocation can be 2-5x faster and reduce allocations by an order of magnitude. Always preallocate when converting one slice type to another or accumulating results.

Alternative: If you know the exact size, use `make([]Result, len(items))` and index directly instead of append.
