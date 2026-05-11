## Avoid defer in Tight Loops

**Impact: MEDIUM**

Defer statements are convenient but have overhead. In tight loops processing many items, defer accumulates on the stack and doesn't execute until the function returns, potentially causing resource exhaustion.

**Incorrect:**

```go
func ProcessFiles(paths []string) error {
    for _, path := range paths {
        file, err := os.Open(path)
        if err != nil {
            return err
        }
        defer file.Close()  // All defers execute at function return!
        // Process file...
    }
    return nil
    // All files stay open until here
}
```

**Correct:**

```go
func ProcessFiles(paths []string) error {
    for _, path := range paths {
        if err := processFile(path); err != nil {
            return err
        }
    }
    return nil
}

func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()  // Executes when processFile returns
    // Process file...
    return nil
}
```

**Alternative (immediate cleanup):**

```go
func ProcessFiles(paths []string) error {
    for _, path := range paths {
        file, err := os.Open(path)
        if err != nil {
            return err
        }

        err = processFileContents(file)
        file.Close()  // Close immediately after use

        if err != nil {
            return err
        }
    }
    return nil
}
```

Why this matters: With 10,000 files, the incorrect version keeps all file handles open until function exit, potentially hitting OS limits (typically 1024-4096 open files). The correct version closes each file immediately after processing.

Rule: Use defer for cleanup in normal functions. In loops, either extract to a helper function with its own defer, or close resources explicitly in the loop.
