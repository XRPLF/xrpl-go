## Use Table-Driven Tests for Multiple Cases

**Impact: MEDIUM**

When testing a function with multiple input/output combinations, use table-driven tests instead of duplicating test code. This makes tests more maintainable and easier to extend.

**Incorrect:**

```go
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }

    result = Add(0, 0)
    if result != 0 {
        t.Errorf("Add(0, 0) = %d; want 0", result)
    }

    result = Add(-1, 1)
    if result != 0 {
        t.Errorf("Add(-1, 1) = %d; want 0", result)
    }
    // Lots of duplication...
}
```

**Correct:**

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {name: "positive numbers", a: 2, b: 3, want: 5},
        {name: "zeros", a: 0, b: 0, want: 0},
        {name: "negative and positive", a: -1, b: 1, want: 0},
        {name: "large numbers", a: 1000, b: 2000, want: 3000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

Why this matters: Table-driven tests reduce duplication, make it easy to add new test cases (just add a struct), and provide clear test names. When a test fails, you immediately know which case failed. They also work well with `t.Run()` for subtests.

Benefits:
- Adding new test cases is trivial (one line)
- Test output shows which specific case failed
- Easy to run individual cases: `go test -run TestAdd/zeros`
- Less code to maintain

Pattern: Define a slice of test case structs with inputs, expected outputs, and descriptive names. Loop through and run each as a subtest with `t.Run()`.
