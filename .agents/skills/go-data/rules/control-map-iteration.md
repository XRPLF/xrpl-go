## Never Assume Map Iteration Order or Stability

**Impact: MEDIUM**

Go maps have intentionally non-deterministic iteration order. The order changes between runs and even between iterations of the same map in the same program. Additionally, inserting elements into a map during iteration may or may not be visited — the behavior is unspecified.

**Incorrect:**

```go
// Assuming alphabetical or insertion order
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k := range m {
    fmt.Print(k)  // Output varies: "abc", "bac", "cab", etc.
}

// Assuming stable order across iterations
for i := 0; i < 2; i++ {
    for k := range m {
        fmt.Print(k)
    }
    fmt.Println()
}
// Might print:
// zdyaec
// czyade  (different order each time)

// Assuming inserted elements are visited
m2 := map[int]bool{0: true, 1: false, 2: true}
for k, v := range m2 {
    if v {
        m2[10+k] = true  // May or may not be visited — non-deterministic
    }
}
// Result varies between runs
```

**Correct:**

```go
// If order matters, collect keys and sort them
keys := make([]string, 0, len(m))
for k := range m {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
    fmt.Printf("%s: %d\n", k, m[k])
}

// If you need to update a map based on iteration, use a copy
m2Copy := copyMap(m2)
for k, v := range m2 {
    if v {
        m2Copy[10+k] = true  // Update the copy, iterate the original
    }
}
```

Why this matters: Go intentionally randomizes map iteration to prevent developers from relying on ordering (a design choice made explicit in the spec). An element added during iteration "may be produced during the iteration or skipped" — both are valid. Never rely on map ordering for correctness; use sorted slices or ordered data structures when order matters.
