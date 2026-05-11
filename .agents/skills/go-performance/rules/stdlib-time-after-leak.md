## time.After in a Loop Leaks Memory — Use time.NewTimer Instead

**Impact: HIGH**

`time.After(d)` creates a new timer each call. The resources (including the channel) are only released when the timer fires. In a loop or frequently-called function, calling `time.After` without the timer ever firing leads to memory accumulation. Use `time.NewTimer` with `Reset()` to reuse the timer.

**Incorrect:**

```go
func consumer(ch <-chan Event) {
    for {
        select {
        case event := <-ch:
            handle(event)
        case <-time.After(time.Hour):  // BUG: creates a new timer every iteration
            log.Println("warning: no messages received")
            // Each iteration that handles an event leaves a 1-hour timer
            // consuming ~200 bytes until it fires
            // At 5M events/hour: consumes 1 GB of memory
        }
    }
}
```

**Correct:**

```go
func consumer(ch <-chan Event) {
    timerDuration := 1 * time.Hour
    timer := time.NewTimer(timerDuration)   // Create once
    defer timer.Stop()                      // Stop on exit

    for {
        timer.Reset(timerDuration)          // Reset each iteration — no allocation
        select {
        case event := <-ch:
            handle(event)
        case <-timer.C:
            log.Println("warning: no messages received")
        }
    }
}
```

Why this matters: `time.After` is implemented as `time.NewTimer(d).C` — it creates a full timer object but returns only the channel. Since the caller has no reference to the timer, it can't stop it. The timer's resources are held for the full duration, even if the channel is never read. In a loop processing millions of events, this creates millions of timers. `time.NewTimer` with `Reset()` reuses a single timer with no new heap allocation. This applies to loops, HTTP handlers (called repeatedly), and Kafka consumer functions.
