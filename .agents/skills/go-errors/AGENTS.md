# Go Error Management

**Version 0.1.0**  
Agent Skills  
March 2026

> **Note:**  
> This document is mainly for agents and LLMs to follow when maintaining,  
> generating, or refactoring Go error handling, wrapping, sentinel errors, and panic patterns. Humans  
> may also find it useful, but guidance here is optimized for automation  
> and consistency by AI-assisted workflows.

---

## Abstract

Guidelines for robust Go error handling. Covers sentinel errors, error wrapping, type checking, panicking, defer error patterns, and avoiding silent failures.

---

## Table of Contents

1. [Error Management](#1-error-management) — **CRITICAL**
   - 1.1 [Don't Ignore Errors Returned by Deferred Functions](#11-dont-ignore-errors-returned-by-deferred-functions)
   - 1.2 [Handle an Error Only Once — Log OR Return, Not Both](#12-handle-an-error-only-once--log-or-return-not-both)
   - 1.3 [Never Ignore Error Return Values](#13-never-ignore-error-return-values)
   - 1.4 [Use errors.Is and errors.As for Error Type Checking](#14-use-errorsis-and-errorsas-for-error-type-checking)
   - 1.5 [Use errors.Is to Compare Sentinel Errors, Not ==](#15-use-errorsis-to-compare-sentinel-errors-not-)
   - 1.6 [Use Panic Only for Unrecoverable Errors](#16-use-panic-only-for-unrecoverable-errors)
   - 1.7 [Wrap Errors with Context Using %w](#17-wrap-errors-with-context-using-w)

---

## 1. Error Management

**Impact: CRITICAL**

Proper error handling is fundamental to writing reliable Go code. Panicking inappropriately, ignoring errors, improper wrapping, and incorrect error type checking are among the most common sources of production bugs. These patterns ensure robust error handling.

### 1.1 Don't Ignore Errors Returned by Deferred Functions

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

### 1.2 Handle an Error Only Once — Log OR Return, Not Both

**Impact: MEDIUM**

Handling an error twice — logging it and then returning it — causes the same error to appear multiple times in logs at different layers. Choose one: either handle the error (log it, recover from it) and return nil, or add context and return it up the call stack. Never do both.

**Incorrect:**

```go
func GetRoute(srcLat, srcLng, dstLat, dstLng float32) (Route, error) {
    err := validateCoordinates(srcLat, srcLng)
    if err != nil {
        log.Println("failed to validate source coordinates")  // Logged here...
        return Route{}, err                                    // ...AND returned
    }

    err = validateCoordinates(dstLat, dstLng)
    if err != nil {
        log.Println("failed to validate destination coordinates")
        return Route{}, err  // Double handling: log + return
    }

    return computeRoute(srcLat, srcLng, dstLat, dstLng)
}

// Caller also logs the error:
route, err := GetRoute(...)
if err != nil {
    log.Println("failed to get route:", err)  // Logged AGAIN — duplicate noise
}
// Result: same error logged 2+ times at different layers
```

**Correct:**

```go
// Option 1: Add context and return — let the top level log
func GetRoute(srcLat, srcLng, dstLat, dstLng float32) (Route, error) {
    if err := validateCoordinates(srcLat, srcLng); err != nil {
        return Route{}, fmt.Errorf("invalid source coordinates: %w", err)  // Context added
    }
    if err := validateCoordinates(dstLat, dstLng); err != nil {
        return Route{}, fmt.Errorf("invalid destination coordinates: %w", err)
    }
    return computeRoute(srcLat, srcLng, dstLat, dstLng)
}

// Top-level handler logs once with full context:
route, err := GetRoute(...)
if err != nil {
    log.Println("failed to get route:", err)
    // Error message: "failed to get route: invalid source coordinates: lat out of range"
}

// Option 2: Handle fully at the point of occurrence
func processRequest(r *http.Request) {
    if err := validateRequest(r); err != nil {
        log.Printf("invalid request: %v", err)
        // Handle the error: respond, recover, and return nil upward
        return  // Do NOT also return the error
    }
}
```

Why this matters: Each layer should either handle an error completely (log, recover, continue) or add context and propagate it upward — never both. Logging AND returning causes log flooding with duplicate entries, making debugging harder. Use `fmt.Errorf("context: %w", err)` to enrich errors as they propagate, then log them once at the appropriate boundary.

### 1.3 Never Ignore Error Return Values

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

### 1.4 Use errors.Is and errors.As for Error Type Checking

**Impact: HIGH**

Checking error types with `==` or type assertions breaks when errors are wrapped. Use `errors.Is()` for sentinel errors and `errors.As()` for error types to work correctly with wrapped errors.

**Incorrect:**

```go
func HandleError(err error) {
    if err == os.ErrNotExist {  // Fails if error is wrapped!
        // Handle missing file
    }

    if _, ok := err.(*json.SyntaxError); ok {  // Fails with wrapping!
        // Handle JSON error
    }
}
```

**Correct:**

```go
func HandleError(err error) {
    if errors.Is(err, os.ErrNotExist) {  // Works with wrapped errors
        // Handle missing file
    }

    var syntaxErr *json.SyntaxError
    if errors.As(err, &syntaxErr) {  // Works with wrapped errors
        // Handle JSON error, can access syntaxErr fields
    }
}
```

Why this matters: When you wrap errors with `fmt.Errorf("%w", err)`, direct comparisons fail because the error is now a different type. `errors.Is()` unwraps the error chain to find matches. `errors.As()` finds the first error in the chain that matches the target type and assigns it, giving you access to type-specific fields.

### 1.5 Use errors.Is to Compare Sentinel Errors, Not ==

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

### 1.6 Use Panic Only for Unrecoverable Errors

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

### 1.7 Wrap Errors with Context Using %w

**Impact: CRITICAL**

When propagating errors up the call stack, always add context about what operation failed. Use `%w` verb with `fmt.Errorf` to wrap errors while preserving the original error for inspection with `errors.Is()` and `errors.As()`.

**Incorrect:**

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err  // No context about what failed
    }

    var cfg Config
    err = json.Unmarshal(data, &cfg)
    if err != nil {
        return nil, errors.New("unmarshal failed")  // Lost original error
    }

    return &cfg, nil
}
```

**Correct:**

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading config file %q: %w", path, err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config JSON: %w", err)
    }

    return &cfg, nil
}
```

Why this matters: The incorrect version returns cryptic errors like "no such file" without context. The correct version returns "reading config file '/etc/app.conf': no such file or directory", making debugging trivial. The `%w` verb allows callers to check for specific error types: `if errors.Is(err, os.ErrNotExist)`.

Best practice: Include variable values (like file paths) in error messages to aid debugging, but avoid including sensitive data like passwords or tokens.

---

## References

1. [https://go.dev/doc/effective_go](https://go.dev/doc/effective_go)
2. [https://go.dev/blog/error-handling-and-go](https://go.dev/blog/error-handling-and-go)
3. [https://go.dev/blog/go1.13-errors](https://go.dev/blog/go1.13-errors)
