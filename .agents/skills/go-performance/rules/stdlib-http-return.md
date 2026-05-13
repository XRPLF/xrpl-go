## Always Return After http.Error — It Does Not Stop Handler Execution

**Impact: HIGH**

`http.Error` writes an error response and sets the status code, but it does **not** stop the handler from running. If you omit `return` after `http.Error`, the handler continues executing and may write a second response, leading to garbled output, incorrect status codes, or "superfluous response.WriteHeader call" warnings.

**Incorrect:**

```go
func handler(w http.ResponseWriter, r *http.Request) {
    if err := validate(r); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        // BUG: execution continues — handler writes a second response body
    }
    // This runs even after the error response was sent
    fmt.Fprintln(w, "OK")
}
```

**Correct:**

```go
func handler(w http.ResponseWriter, r *http.Request) {
    if err := validate(r); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return  // Stop execution after sending the error response
    }

    fmt.Fprintln(w, "OK")
}

// Same applies when redirecting:
func redirectHandler(w http.ResponseWriter, r *http.Request) {
    if !authorized(r) {
        http.Redirect(w, r, "/login", http.StatusFound)
        return  // Must return — http.Redirect does not stop execution
    }

    renderDashboard(w, r)
}
```

Why this matters: `http.Error`, `http.Redirect`, and `http.NotFound` all write to the `ResponseWriter` and set headers, but they are ordinary function calls — they don't panic or halt the goroutine. The HTTP response body is the concatenation of everything written to `w`. Without `return`, you send two responses into the same connection, causing protocol errors or double-writing. Always follow any response-sending function with an immediate `return` (or `else` block) to guard subsequent code.
