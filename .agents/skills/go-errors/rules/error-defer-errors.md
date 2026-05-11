## Don't Ignore Errors Returned by Deferred Functions

**Impact: MEDIUM**

Deferred calls like `rows.Close()` or `file.Close()` can return errors that indicate data loss or resource problems. Silently ignoring these errors with bare `defer f.Close()` hides real failures. Propagate or log defer errors explicitly.

**Incorrect:**

```go
func getBalance(db *sql.DB, clientID string) (float32, error) {
    rows, err := db.Query("SELECT balance FROM accounts WHERE id = ?", clientID)
    if err != nil {
        return 0, err
    }
    defer rows.Close()  // BUG: error from Close() is silently discarded

    // Process rows...
    return balance, nil
}
```

**Correct:**

```go
// Option 1: Use named result params + closure to propagate the error
func getBalance(db *sql.DB, clientID string) (balance float32, err error) {
    rows, err := db.Query("SELECT balance FROM accounts WHERE id = ?", clientID)
    if err != nil {
        return 0, err
    }
    defer func() {
        closeErr := rows.Close()
        if err == nil {         // Only overwrite err if there isn't already one
            err = closeErr      // Propagate close error when no other error exists
        }
    }()

    // Process rows...
    return balance, nil
}

// Option 2: Log the defer error (appropriate when propagation isn't practical)
func getBalance(db *sql.DB, clientID string) (float32, error) {
    rows, err := db.Query("SELECT balance FROM accounts WHERE id = ?", clientID)
    if err != nil {
        return 0, err
    }
    defer func() {
        if err := rows.Close(); err != nil {
            log.Printf("failed to close rows: %v", err)
        }
    }()

    return balance, nil
}
```

Why this matters: `rows.Close()`, `file.Close()`, and similar cleanup calls can return errors signaling that buffered data wasn't flushed, locks weren't released, or other critical failures occurred. In Option 1, the named result `err` is shared with the deferred closure; if the function returned successfully (`err == nil`), the close error replaces it. If the function already has an error, the original error takes priority. Choose propagation for critical resources and logging for best-effort cleanup.
