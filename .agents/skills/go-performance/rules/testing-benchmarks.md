## Write Accurate Benchmarks: Reset Timer, Prevent Compiler Optimizations

**Impact: MEDIUM**

Go benchmarks are easy to write incorrectly. Three common mistakes produce misleading results: (1) not resetting the timer after expensive setup; (2) allowing the compiler to inline and eliminate the function under test; (3) ignoring environmental variance in micro-benchmarks.

**Mistake 1: Not resetting or pausing the timer**

```go
// WRONG: setup time is included in benchmark measurement
func BenchmarkFoo(b *testing.B) {
    expensiveSetup()  // Counted in the result!
    for i := 0; i < b.N; i++ {
        functionUnderTest()
    }
}

// CORRECT: reset timer after one-time setup
func BenchmarkFoo(b *testing.B) {
    expensiveSetup()
    b.ResetTimer()  // Zero elapsed time and allocation counters
    for i := 0; i < b.N; i++ {
        functionUnderTest()
    }
}

// CORRECT: pause/resume timer for per-iteration setup
func BenchmarkFoo(b *testing.B) {
    for i := 0; i < b.N; i++ {
        b.StopTimer()      // Pause the benchmark clock
        expensiveSetup()
        b.StartTimer()     // Resume the benchmark clock
        functionUnderTest()
    }
}
```

**Mistake 2: Compiler inlines the function, making the benchmark empty**

```go
// WRONG: compiler may inline popcnt and then eliminate the call as having no side effects
func BenchmarkPopcnt1(b *testing.B) {
    for i := 0; i < b.N; i++ {
        popcnt(uint64(i))  // Result is unused — compiler may remove entirely
    }
}
// Result: 0.28 ns/op (suspiciously fast — near one clock cycle!)

// CORRECT: assign result to local variable, then to a package-level variable
var global uint64

func BenchmarkPopcnt2(b *testing.B) {
    var v uint64
    for i := 0; i < b.N; i++ {
        v = popcnt(uint64(i))  // Assign to local (keeps function call)
    }
    global = v  // Assign to global (prevents compiler from eliminating local)
}
// Result: 1.99 ns/op (accurate)
```

**Mistake 3: Micro-benchmark order affects results**

```go
// Benchmark order can change results due to CPU caching, thermal throttling, etc.
// Running Int32 first may show it as faster; running Int64 first may show the opposite.

// SOLUTION: use -count and benchstat for statistical analysis
// $ go test -bench=. -count=10 | tee stats.txt
// $ benchstat stats.txt
// name                    time/op
// AtomicStoreInt32-4      5.10ns ± 1%
// AtomicStoreInt64-4      5.10ns ± 1%
// (Both are equivalent — random order was misleading)
```

**Running benchmarks:**

```bash
# Run benchmarks (not regular tests)
go test -bench=. ./...

# Increase benchmark time for more stable results
go test -bench=. -benchtime=5s ./...

# Run N times for statistical comparison
go test -bench=. -count=10 | tee stats.txt
benchstat stats.txt  # golang.org/x/perf/cmd/benchstat
```

Why this matters: (1) Setup code before the benchmark loop inflates results if the timer isn't reset. (2) Go's compiler can inline small functions and then eliminate calls with no observable side effects — benchmarks that look fast may actually be benchmarking nothing. Assigning to a local variable prevents elimination; assigning the local to a global prevents the local from being optimized away. (3) Micro-benchmark results can flip based on which benchmark runs first due to CPU state. Use `benchstat` with `-count=10` to compute stable statistics and detect when apparent differences are within noise.
