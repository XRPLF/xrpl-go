## Use copy Correctly: Destination Must Have Length Set

**Impact: MEDIUM**

The built-in `copy(dst, src)` copies `min(len(dst), len(src))` elements. If the destination slice has zero length (even if it has non-zero capacity), nothing is copied. This is a frequent source of bugs when developers allocate capacity but forget to set length.

**Incorrect:**

```go
src := []int{1, 2, 3, 4, 5}

// Bug: make([]int, 0, len(src)) creates len=0, cap=5
dst := make([]int, 0, len(src))
n := copy(dst, src)
fmt.Println(n, dst)  // Prints: 0 []  -- nothing was copied!

// Bug: nil dst
var dst2 []int
copy(dst2, src)  // Copies nothing, no error
```

**Correct:**

```go
src := []int{1, 2, 3, 4, 5}

// Correct: make([]int, len(src)) sets both length and capacity
dst := make([]int, len(src))
n := copy(dst, src)
fmt.Println(n, dst)  // Prints: 5 [1 2 3 4 5] ✓

// Alternative: append to copy (idiomatic)
dst2 := append([]int(nil), src...)
fmt.Println(dst2)   // [1 2 3 4 5] ✓

// Partial copy: copy first 3 elements
dst3 := make([]int, 3)
copy(dst3, src)
fmt.Println(dst3)   // [1 2 3]

// Copy into a sub-slice
dst4 := make([]int, 10)
copy(dst4[2:], src)  // Copies src into dst4 starting at index 2
fmt.Println(dst4)    // [0 0 1 2 3 4 5 0 0 0]
```

Why this matters: The `copy` function doesn't resize the destination — that's `append`'s job. `make([]int, 0, cap)` creates a slice with length 0, not `cap`. Always use `make([]T, length)` when you want to `copy` into it. If you want a defensive copy of a slice, `append([]T(nil), src...)` is idiomatic and harder to get wrong.
