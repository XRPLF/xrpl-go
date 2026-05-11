## Understand Slice Length vs Capacity

**Impact: HIGH**

Slices have both length (`len`) and capacity (`cap`). Length is the number of elements; capacity is the size of the underlying array. Confusing these leads to unexpected slice behavior, data loss, and shared backing arrays causing mysterious mutations.

**Incorrect:**

```go
func ProcessBatch(items []int) []int {
    results := make([]int, 10)  // Length 10, capacity 10
    for i, item := range items {
        results[i] = item * 2  // Panics if len(items) > 10!
    }
    return results
}
```

**Correct:**

```go
func ProcessBatch(items []int) []int {
    results := make([]int, 0, len(items))  // Length 0, capacity len(items)
    for _, item := range items {
        results = append(results, item*2)  // Safe, preallocated
    }
    return results
}
```

Why this matters: `make([]T, n)` creates a slice with `n` zero elements. `make([]T, 0, n)` creates an empty slice with space for `n` elements. The first form causes index-out-of-bounds panics or overwrites zeros. Slicing operations also preserve the underlying array—slicing `s[0:5]` from a 100-element slice still references all 100 elements, preventing garbage collection.

Pattern: Use `make([]T, 0, n)` when building slices with append. Use `make([]T, n)` when setting elements by index. Be aware that `slice[i:j]` shares the backing array with the original slice.