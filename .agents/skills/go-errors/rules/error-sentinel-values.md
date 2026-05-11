## Use errors.Is to Compare Sentinel Errors, Not ==

**Impact: HIGH**

Sentinel errors are package-level error values (like `sql.ErrNoRows`, `io.EOF`). Comparing them with `==` breaks when errors are wrapped with `fmt.Errorf("%w", err)`. Use `errors.Is` to safely check for sentinel errors through any wrapping chain.

**Incorrect:**

```go
// Direct == comparison breaks with wrapped errors
err := queryDatabase()
if err == sql.ErrNoRows {  // BUG: fails if err is wrapped
    return defaultValue, nil
}

// Example of wrapping that breaks == check:
func queryDatabase() error {
    err := db.QueryRow("SELECT ...").Scan(&val)
    if err != nil {
        return fmt.Errorf("queryDatabase: %w", err)  // wraps sql.ErrNoRows
    }
    return nil
}

// Caller's == check now misses the error:
if err == sql.ErrNoRows {  // FALSE — err is the wrapped error, not sql.ErrNoRows directly
    // This branch never executes!
}
```

**Correct:**

```go
// errors.Is unwraps the chain and checks each level
err := queryDatabase()
if errors.Is(err, sql.ErrNoRows) {  // Works regardless of wrapping depth
    return defaultValue, nil
}

// Custom sentinel errors — define at package level
var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("invalid input")

func findUser(id int) (*User, error) {
    if id <= 0 {
        return nil, fmt.Errorf("findUser: %w", ErrInvalidInput)  // Wrapped
    }
    // ...
}

// Caller can still detect the original error type:
if errors.Is(err, ErrInvalidInput) {  // TRUE even when wrapped
    http.Error(w, "Bad Request", 400)
}
```

Why this matters: `errors.Is` traverses the entire error chain by calling `Unwrap()` repeatedly. A direct `==` comparison only matches the outermost error value. Since idiomatic Go wraps errors with context at every layer, `==` comparisons will silently fail in real code. Always use `errors.Is` for sentinel errors and `errors.As` for type-based error checking.
