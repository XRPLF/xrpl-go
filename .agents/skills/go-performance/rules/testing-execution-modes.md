## Use t.Parallel() for Parallel Tests and -shuffle to Detect Order Dependencies

**Impact: MEDIUM**

By default, Go runs tests within a package sequentially. Slow test suites benefit from parallelism via `t.Parallel()`. Additionally, tests that pass only in a specific order hide shared state bugs — use `-shuffle=on` to detect these dependencies.

**Using t.Parallel():**

```go
func TestProcessOrder(t *testing.T) {
    t.Parallel()  // This test can run concurrently with other parallel tests

    result := processOrder(Order{ID: 1, Amount: 100})
    if result.Status != "processed" {
        t.Errorf("expected processed, got %s", result.Status)
    }
}

// Table-driven tests: parallelize each subtest
func TestValidate(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  bool
    }{
        {"empty", "", false},
        {"valid", "hello@example.com", true},
        {"invalid", "not-an-email", false},
    }

    for _, tt := range tests {
        tt := tt  // Capture range variable (required before Go 1.22)
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // Subtests run concurrently
            got := validate(tt.input)
            if got != tt.want {
                t.Errorf("validate(%q) = %v, want %v", tt.input, got, tt.want)
            }
        })
    }
}
```

**Important: parallel tests and shared resources**

```go
// WRONG: shared mutable state in parallel tests causes data races
var counter int  // package-level

func TestIncrement(t *testing.T) {
    t.Parallel()
    counter++  // Data race!
}

// CORRECT: each test owns its state
func TestIncrement(t *testing.T) {
    t.Parallel()
    counter := 0  // Local variable
    counter++
    // ...
}
```

**Using -shuffle to detect order-dependent tests:**

```bash
# Randomize test execution order
go test -shuffle=on ./...

# Use a specific seed to reproduce a failure
go test -shuffle=on -v ./...
# Output includes: -test.shuffle 1673506577441440997

# Reproduce with same seed
go test -shuffle=1673506577441440997 ./...
```

```go
// This test hides an order dependency — it only passes if TestInit runs first
func TestProcess(t *testing.T) {
    // BUG: relies on globalState set by TestInit
    if globalState == nil {
        t.Fatal("globalState is nil")  // Fails when order is shuffled
    }
}

// CORRECT: each test initializes what it needs
func TestProcess(t *testing.T) {
    state := newState()  // No dependency on other tests
    result := process(state)
    // ...
}
```

Why this matters: Sequential tests in a large package can take minutes; parallel tests with `t.Parallel()` can reduce that to seconds on modern hardware. `-shuffle=on` is essential for finding tests that silently depend on execution order or shared global state — these tests will pass in CI but randomly fail when the suite grows or order changes. Run with `-shuffle=on` regularly to keep your test suite isolated and reliable.
