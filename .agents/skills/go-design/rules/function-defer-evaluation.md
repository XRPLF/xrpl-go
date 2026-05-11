## Defer Arguments Are Evaluated Immediately, Not When Called

**Impact: MEDIUM**

`defer` schedules a function call for when the surrounding function returns, but it evaluates the function's arguments **immediately** when the `defer` statement is reached. If you rely on a variable that changes later, the deferred call will use the value at the time of the `defer` — not the final value.

**Incorrect:**

```go
func notify(status string)          { /* send notification */ }
func incrementCounter(status string) { /* update metric */ }

func process() error {
    var status string
    defer notify(status)          // BUG: status="" captured NOW, not later
    defer incrementCounter(status) // BUG: status="" captured NOW

    if err := foo(); err != nil {
        status = "error_foo"
        return err
    }
    if err := bar(); err != nil {
        status = "error_bar"
        return err
    }
    status = "success"
    return nil
    // notify and incrementCounter always receive "" regardless of path
}
```

**Correct:**

```go
// Option 1: Pass a pointer so the deferred function reads the final value
func process() error {
    var status string
    defer notify(&status)           // Pointer to status — reads value at call time
    defer incrementCounter(&status)

    if err := foo(); err != nil {
        status = "error_foo"
        return err
    }
    if err := bar(); err != nil {
        status = "error_bar"
        return err
    }
    status = "success"
    return nil
    // notify and incrementCounter receive the final status value ✓
}

// Option 2: Use a closure — captures variables by reference, not by value
func process() error {
    var status string
    defer func() {
        notify(status)           // Reads status when the closure executes
        incrementCounter(status)
    }()

    // ... same logic
    status = "success"
    return nil
}
```

Why this matters: `defer notify(status)` is equivalent to saving the call as `notify("")` at that exact moment. This is a frequent source of bugs in logging, metrics, and cleanup code. With a **closure**, variables referenced from the outer scope are read when the closure runs (on function return), not when `defer` is called. With a **pointer**, the deferred function dereferences it at call time. Both solutions correctly capture the final state.
