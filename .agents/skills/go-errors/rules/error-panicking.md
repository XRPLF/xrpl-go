## Use Panic Only for Unrecoverable Errors

**Impact: CRITICAL**

Panic should be reserved for truly unexpected, unrecoverable situations. Using panic for expected errors crashes your program and is un-idiomatic Go. Return errors instead.

**Incorrect:**

```go
func LoadConfig(path string) *Config {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err)  // Don't panic on expected errors!
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        panic(err)  // Crashes the whole program
    }
    return &cfg
}
```

**Correct:**

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading config: %w", err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }
    return &cfg, nil
}
```

Why this matters: Panic terminates the program (or goroutine) immediately unless recovered. File not found, invalid JSON, and network errors are expected failures that callers should handle. Panic makes your code unusable as a library and prevents graceful error handling.

When to panic: Programming errors (e.g., `nil` pointer bugs you should fix), impossible states that indicate broken invariants, or initialization failures in `init()` where error returns aren't possible. Even then, consider whether `log.Fatal` is more appropriate.
