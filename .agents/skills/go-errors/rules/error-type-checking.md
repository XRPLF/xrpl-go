## Use errors.Is and errors.As for Error Type Checking

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
