## Wrap Errors with Context Using %w

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
