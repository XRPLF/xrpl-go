## Use Linters to Catch Common Mistakes

**Impact: MEDIUM**

Go's compiler catches many errors but not all subtle bugs. Linters perform static analysis to catch issues the compiler misses: shadowed variables, unchecked errors, complexity hotspots, duplicate code, and security vulnerabilities. Running linters in CI prevents these issues from reaching production.

**Incorrect:**

```go
// These issues compile fine but linters catch them:

func ProcessFile(path string) {
    f, _ := os.Open(path)     // errcheck: error return ignored
    defer f.Close()

    data := make([]byte, 1024)
    n, _ := f.Read(data)      // errcheck: error return ignored

    result := data[:n]
    _ = result
}

func IsPrime(n int) bool {
    // gocyclo: function complexity too high (many nested conditions)
    ...
}
```

**Correct:**

```go
// .golangci.yml — configure linters at the repo root
// Run: golangci-lint run ./...

func ProcessFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("opening %s: %w", path, err)
    }
    defer f.Close()

    data := make([]byte, 1024)
    n, err := f.Read(data)
    if err != nil && err != io.EOF {
        return fmt.Errorf("reading %s: %w", path, err)
    }

    return process(data[:n])
}
```

Why this matters: Linters act as automated code reviewers that run consistently. Key linters to enable:
- `errcheck` — finds ignored error returns
- `gosimple` — suggests simpler code patterns
- `staticcheck` — advanced static analysis
- `govet` — Go's own vet tool (reports suspicious constructs)
- `gocyclo` — flags high cyclomatic complexity
- `gosec` — security-focused analysis

Add `.golangci.yml` to your project root and run `golangci-lint run` in CI. Fix issues before they accumulate — a clean lint baseline is easy to maintain.

Reference: [golangci-lint](https://golangci-lint.run/)
