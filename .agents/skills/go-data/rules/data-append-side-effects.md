## Beware of Append Side Effects with Shared Backing Arrays

**Impact: HIGH**

Slices created by slicing a larger slice share the same backing array. Appending to one may modify the other if the slice hasn't grown beyond the original capacity. This leads to hard-to-debug data corruption.

**Incorrect:**

```go
func main() {
    original := []int{1, 2, 3, 4, 5}

    // Both share the same backing array
    first := original[:3]   // [1, 2, 3], cap=5
    second := original[3:]  // [4, 5], cap=2

    // Appending to first when len < cap modifies original's backing array!
    first = append(first, 99)
    fmt.Println(original)  // [1, 2, 3, 99, 5] -- original is mutated!
    fmt.Println(first)     // [1, 2, 3, 99]

    // Functions receiving sub-slices can mutate caller's data
    func modifySlice(s []int) {
        s[0] = 999  // Modifies backing array — visible to caller!
    }
}
```

**Correct:**

```go
func main() {
    original := []int{1, 2, 3, 4, 5}

    // Option 1: Use full slice expression to limit capacity
    // s[low:high:max] — cap of result = max - low
    first := original[:3:3]  // cap=3, append will allocate new array
    first = append(first, 99)
    fmt.Println(original)    // [1, 2, 3, 4, 5] -- original unchanged ✓
    fmt.Println(first)       // [1, 2, 3, 99]

    // Option 2: Copy the slice to get independent backing array
    firstCopy := make([]int, 3)
    copy(firstCopy, original[:3])
    firstCopy = append(firstCopy, 99)
    fmt.Println(original)    // [1, 2, 3, 4, 5] -- unchanged ✓

    // Option 3: append idiom for independent copy
    firstCopy2 := append([]int(nil), original[:3]...)
}

// When returning sub-slices from functions, always copy
func GetFirst3(data []int) []int {
    if len(data) < 3 {
        return append([]int(nil), data...)
    }
    return append([]int(nil), data[:3]...)  // Independent copy
}
```

Why this matters: Sharing backing arrays is an optimization in Go's slice design, but it creates a hidden coupling between slices. When you pass a sub-slice to a function or store it in a struct, mutations can propagate unexpectedly. The full slice expression `s[low:high:max]` caps the capacity so that the next append triggers a new allocation and breaks the sharing.
