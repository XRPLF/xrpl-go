## Range Expression Is Evaluated Only Once

**Impact: MEDIUM**

The expression in a `range` statement is evaluated exactly once, before the loop begins — it's copied to a temporary variable. Changes to the original slice/channel/array during the loop are not seen by the range iterator, because it's working on the copy.

**Incorrect:**

```go
// Trying to process all elements including dynamically added ones
s := []int{0, 1, 2}
for range s {
    s = append(s, 10)  // BUG: adding elements doesn't extend the loop
}
fmt.Println(s)  // [0 1 2 10 10 10] — loop ran exactly 3 times, not forever

// Trying to switch channels mid-loop
ch := ch1
for v := range ch {     // range evaluates ch once — copies ch1's value
    fmt.Println(v)
    ch = ch2            // Changing ch has no effect — loop still reads ch1
}
```

**Correct:**

```go
// If you need to process dynamically-added elements, track length manually
s := []int{0, 1, 2}
for i := 0; i < len(s); i++ {   // len(s) re-evaluated each iteration
    s = append(s, 10)            // Infinite loop! len grows each iteration
    // Don't actually do this — but this is the difference
}

// For arrays: range over pointer to avoid copying large array
a := [3]int{0, 1, 2}
a[2] = 10
for i, v := range &a {  // Range over pointer — sees live array values
    if i == 2 {
        fmt.Println(v)   // Prints 10, not 2
    }
}

// For channels: don't reassign the channel variable to switch targets
// Use select with multiple channels instead
```

Why this matters: `range` copies its expression once. For slices this means a snapshot of (ptr, len, cap). Appending during the loop may reallocate the backing array, but the range iterator still uses the original length. For arrays, range copies the entire array. Use `range &arr` to avoid the copy and see live updates. For channels, reassigning the variable doesn't switch what's being ranged over.
