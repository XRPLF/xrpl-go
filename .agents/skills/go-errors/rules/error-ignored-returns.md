## Never Ignore Error Return Values

**Impact: CRITICAL**

One of the most dangerous patterns in Go is ignoring error return values. Functions that return errors do so because failure is a real possibility. Ignoring these errors leads to silent failures, corrupted state, and difficult-to-debug production issues.

**Incorrect:**

```go
func SaveUser(user User) {
    data, _ := json.Marshal(user)  // Ignoring error
    file.Write(data)                // Ignoring error
    file.Close()                    // Ignoring error
}
```

**Correct:**

```go
func SaveUser(user User) error {
    data, err := json.Marshal(user)
    if err != nil {
        return fmt.Errorf("failed to marshal user: %w", err)
    }

    if err := file.Write(data); err != nil {
        return fmt.Errorf("failed to write data: %w", err)
    }

    if err := file.Close(); err != nil {
        return fmt.Errorf("failed to close file: %w", err)
    }

    return nil
}
```

Why this matters: The incorrect version silently fails if JSON marshaling fails (e.g., circular reference) or if disk is full. The correct version propagates errors to the caller, allowing proper error handling and recovery. In production, this difference is between mysterious data loss and clear error reporting.

Note: If you genuinely need to ignore an error (rare cases like defer file.Close() when you've already successfully read), use explicit blank assignment `_ = file.Close()` to signal intent to code reviewers.
