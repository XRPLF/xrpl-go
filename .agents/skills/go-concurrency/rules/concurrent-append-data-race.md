## Concurrent append on a Shared Slice Can Cause a Data Race

**Impact: HIGH**

`append` is not thread-safe. If a slice has spare capacity (`len < cap`), two goroutines appending concurrently write to the same backing array, causing a data race. If the slice is full (`len == cap`), each `append` allocates a new backing array, so there's no race — but this is fragile and depends on initialization.

**Incorrect:**

```go
// Slice with spare capacity — data race!
s := make([]int, 0, 1)  // len=0, cap=1: has spare capacity

go func() {
    s1 := append(s, 1)  // Both goroutines write to index 0 of the same backing array
    fmt.Println(s1)
}()

go func() {
    s2 := append(s, 1)  // DATA RACE: concurrent write to same memory
    fmt.Println(s2)
}()

// Coincidentally safe (but fragile): slice at full capacity
s := make([]int, 1, 1)  // len=1, cap=1: full capacity
// Each append allocates a new backing array — no race, but only by accident
```

**Correct:**

```go
// Each goroutine works on its own copy
s := make([]int, 0, 1)

go func() {
    sCopy := make([]int, len(s), cap(s))
    copy(sCopy, s)
    s1 := append(sCopy, 1)  // Appends to its own copy — no race
    fmt.Println(s1)
}()

go func() {
    sCopy := make([]int, len(s), cap(s))
    copy(sCopy, s)
    s2 := append(sCopy, 1)  // Independent copy
    fmt.Println(s2)
}()

// Or: protect with a mutex if goroutines must share the slice
var mu sync.Mutex
var shared []int

go func() {
    mu.Lock()
    shared = append(shared, 1)
    mu.Unlock()
}()
```

Why this matters: The behavior of `append` depends on whether the slice is full. If `len < cap`, `append` writes into the existing backing array without allocating — making concurrent appends a data race. If `len == cap`, `append` allocates a new array, so goroutines don't share memory. Never rely on this coincidence. For concurrent slice modification, either give each goroutine its own copy, use a mutex, or design the algorithm so goroutines write to different indices (guaranteed non-overlapping).
