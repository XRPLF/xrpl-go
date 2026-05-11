## Goroutine Closures Capture Loop Variables by Reference

**Impact: HIGH**

A goroutine closure that references a loop variable captures the variable itself, not its value at the time the goroutine is created. By the time the goroutine runs, the loop variable may have advanced to a later value. This is one of the most common Go concurrency bugs.

**Incorrect:**

```go
s := []int{1, 2, 3}

for _, i := range s {
    go func() {
        fmt.Print(i)  // BUG: captures the variable i, not its value
    }()
}
// Expected: prints 1 2 3 (in some order)
// Actual: may print 2 3 3, or 3 3 3, or other combinations
// All goroutines share the same i variable; most run after the loop ends
```

**Correct:**

```go
// Option 1: Create a local copy inside the loop body
for _, i := range s {
    val := i           // New variable created each iteration
    go func() {
        fmt.Print(val) // Captures val, which is fixed per iteration
    }()
}

// Option 2: Pass the value as a function argument (not a closure)
for _, i := range s {
    go func(val int) {
        fmt.Print(val) // val is a function parameter — not a captured variable
    }(i)               // i is evaluated and passed NOW
}

// Note: Go 1.22+ changed loop variable semantics — each iteration
// gets a new variable, so the bug no longer occurs in Go 1.22+
// But for compatibility and clarity, either option above is still recommended
```

Why this matters: Closures in Go capture variables by reference, not by value. In a goroutine closure, the captured variable `i` is the same memory location used by the loop — it changes each iteration. When the goroutines execute (usually after the loop), they all read the current value of `i`, which is typically the last iteration's value or beyond. This produces non-deterministic output. The fix is to either shadow the variable locally (`val := i`) or pass it as a function argument.
