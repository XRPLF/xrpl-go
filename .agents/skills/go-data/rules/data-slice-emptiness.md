## Check Slice Emptiness with len(), Not nil Comparison

**Impact: LOW**

Testing `s == nil` to check if a slice is empty misses non-nil empty slices (`make([]T, 0)`). Use `len(s) == 0` to reliably check if a slice has no elements, regardless of whether it's nil or just empty.

**Incorrect:**

```go
func ProcessItems(items []string) {
    if items == nil {
        fmt.Println("no items")
        return
    }
    // items might still be empty! (non-nil, length 0)
    for _, item := range items {
        process(item)
    }
}

// This works fine but...
ProcessItems(nil)          // "no items" ✓
ProcessItems([]string{})   // Enters the loop — no output ✓ (loop runs 0 times)
// Actually range on empty slice is fine, but the intent check is wrong:
func HasItems(items []string) bool {
    return items != nil   // Returns false for nil, but also false for []string{}!
}
```

**Correct:**

```go
func ProcessItems(items []string) {
    if len(items) == 0 {
        fmt.Println("no items")
        return
    }
    for _, item := range items {
        process(item)
    }
}

func HasItems(items []string) bool {
    return len(items) > 0  // Works for both nil and non-nil empty slices
}

// Both cases handled correctly:
HasItems(nil)          // false ✓
HasItems([]string{})   // false ✓
HasItems([]string{"a"}) // true ✓
```

Why this matters: `len(nil)` returns 0 in Go, so `len(s) == 0` works correctly for both nil and empty slices. Using `s == nil` introduces a distinction that usually doesn't matter for the caller's intent — they typically want to know "are there elements?", not "was this explicitly initialized?". Using `len` is both correct and communicates intent clearly.
