## time.Duration Is in Nanoseconds, Not Milliseconds

**Impact: MEDIUM**

`time.Duration` is an alias for `int64`, representing a duration in **nanoseconds**. Passing a bare integer like `1000` creates a 1,000 nanosecond (1 microsecond) duration, not 1 second or 1 millisecond as developers from Java/JavaScript backgrounds might expect. Always use time constants.

**Incorrect:**

```go
// WRONG: creates a 1000-nanosecond ticker, not 1-second
ticker := time.NewTicker(1000)  // 1000 ns = 1 microsecond!

// WRONG: 5000 ms? No — 5000 nanoseconds = 5 microseconds
time.Sleep(5000)
```

**Correct:**

```go
// Use time constants for clarity and correctness
ticker := time.NewTicker(time.Second)          // 1 second
ticker := time.NewTicker(500 * time.Millisecond)  // 500 ms
ticker := time.NewTicker(time.Microsecond)     // 1 microsecond (if that's what you want)

time.Sleep(5 * time.Second)
time.Sleep(100 * time.Millisecond)

// Available time constants:
// time.Nanosecond  = 1
// time.Microsecond = 1000
// time.Millisecond = 1000000
// time.Second      = 1000000000
// time.Minute      = 60000000000
// time.Hour        = 3600000000000

// Building durations programmatically:
duration := time.Duration(n) * time.Second  // n seconds
timeout := 30 * time.Second                 // 30 seconds
```

Why this matters: `time.Duration(1000)` is 1000 nanoseconds — 1 microsecond. In Java, `Thread.sleep(1000)` sleeps for 1 second (milliseconds). Go uses nanoseconds as the base unit. The `time` package provides named constants (`time.Second`, `time.Millisecond`, etc.) precisely to avoid this confusion. Always multiply by a time constant rather than passing a raw integer.
