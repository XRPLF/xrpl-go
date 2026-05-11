## Avoid Unnecessary String-to-Byte-Slice Conversions

**Impact: MEDIUM**

Most I/O in Go works with `[]byte`, not `string`. Converting `[]byte` → `string` → `[]byte` just to use string functions is wasteful — each conversion allocates a new copy. The `bytes` package mirrors the `strings` package and operates directly on `[]byte`.

**Incorrect:**

```go
// Reading from io.Reader gives []byte, but we convert to string to use strings package
func sanitize(reader io.Reader) ([]byte, error) {
    b, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }

    // Unnecessary: []byte → string → []byte (2 allocations!)
    s := string(b)
    s = strings.TrimSpace(s)
    return []byte(s), nil
}

// Also wasteful: converting just to check a condition
func hasPrefix(data []byte, prefix string) bool {
    return strings.HasPrefix(string(data), prefix)  // Extra allocation
}
```

**Correct:**

```go
// Use bytes package — same operations, works on []byte directly
func sanitize(reader io.Reader) ([]byte, error) {
    b, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }

    return bytes.TrimSpace(b), nil  // No extra allocations
}

func hasPrefix(data []byte, prefix string) bool {
    return bytes.HasPrefix(data, []byte(prefix))
}

// bytes package mirrors strings package:
// strings.Contains  → bytes.Contains
// strings.Count     → bytes.Count
// strings.Split     → bytes.Split
// strings.TrimSpace → bytes.TrimSpace
// strings.Index     → bytes.Index
// strings.Replace   → bytes.Replace
```

Why this matters: Converting `[]byte` to `string` always copies the data (strings are immutable in Go, so a new allocation is required). If your entire workflow can stay in `[]byte`, you avoid these copies. When working with `io.Reader`, HTTP bodies, file contents, or any I/O, prefer `bytes` package operations over converting to string unnecessarily.
