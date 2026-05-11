## Break Inside switch/select Only Breaks the Innermost Statement

**Impact: MEDIUM**

A `break` statement terminates the innermost `for`, `switch`, or `select` statement. When a `switch` or `select` is nested inside a `for` loop, `break` exits the `switch`/`select`, not the loop. This is a frequent source of subtle infinite loops.

**Incorrect:**

```go
// Trying to break the for loop from inside a switch
for i := 0; i < 5; i++ {
    fmt.Printf("%d ", i)
    switch i {
    case 2:
        break  // BUG: breaks the switch, not the for loop!
    }
}
// Prints: 0 1 2 3 4  (loop runs all 5 iterations)

// Same problem with select inside a for loop
for {
    select {
    case msg := <-ch:
        process(msg)
    case <-ctx.Done():
        break  // BUG: breaks the select, not the for loop — infinite loop!
    }
}
```

**Correct:**

```go
// Use a labeled break to exit the for loop
loop:
    for i := 0; i < 5; i++ {
        fmt.Printf("%d ", i)
        switch i {
        case 2:
            break loop  // Breaks the for loop labeled "loop"
        }
    }
// Prints: 0 1 2

// Same for select inside a for loop
loop:
    for {
        select {
        case msg := <-ch:
            process(msg)
        case <-ctx.Done():
            break loop  // Breaks the for loop, not the select
        }
    }

// Alternative: use return if inside a function
func process(ctx context.Context, ch <-chan int) {
    for {
        select {
        case msg := <-ch:
            handle(msg)
        case <-ctx.Done():
            return  // Returns from the function entirely
        }
    }
}
```

Why this matters: This is a known Go gotcha. `break` terminates the innermost statement — and `switch`/`select` count as statements. Labels are idiomatic in Go (used in the standard library's `net/http` package) and are the correct tool here. The label name should describe the loop's purpose (e.g., `readlines:`, `loop:`) for clarity.
