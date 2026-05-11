## Reduce Data Hazards With Instruction-Level Parallelism

**Impact: LOW**

Modern CPUs can execute multiple independent instructions simultaneously (instruction-level parallelism, or ILP). When instructions depend on previous results — called *data hazards* — the CPU stalls waiting for the dependency to resolve. Restructuring code to eliminate data hazards can yield measurable speedups in tight loops.

**What a data hazard looks like:**

```go
// This loop has a data hazard on s[0]:
// Each iteration reads s[0], then writes s[0] — next iteration depends on that write
func addInt(s []int) {
    for i := 0; i < n; i++ {
        s[0] += s[i]  // Read s[0], add s[i], write s[0] — must wait for previous write
    }
}

// The CPU cannot start the next iteration until the current iteration's write to s[0]
// completes. Operations are serialized even though the CPU has parallel execution units.
```

**Eliminating the hazard with a local accumulator:**

```go
// BETTER: Accumulate in a local variable, write to s[0] once at the end
func addInt(s []int) {
    v := 0
    for i := 0; i < n; i++ {
        v += s[i]  // v is a CPU register — no memory dependency hazard
    }
    s[0] = v  // Single write at end
}

// The CPU can pipeline these additions much more efficiently.
// On a 1000-element slice, this can be ~20% faster.
```

**Mixed-operation example:**

```go
// BAD: data hazard — v1 used in three consecutive instructions
func doSomething(s []int32) {
    for i := 0; i < n; i++ {
        v1 := s[0]        // v1 depends on memory load
        s[0] = v1 + 1     // depends on v1
        if v1%2 == 0 {    // depends on v1 — but previous instruction also depended on v1
            s[1] += 2
        }
    }
}

// GOOD: split into independent operations the CPU can execute in parallel
func doSomething(s []int32) {
    for i := 0; i < n; i++ {
        v1 := s[0]
        s[0] = v1 + 1   // Write s[0]
        v2 := v1 % 2    // Compute in parallel with the write to s[0]
        if v2 == 0 {
            s[1] += 2
        }
    }
}
// CPU can execute the write to s[0] and the modulo computation simultaneously
// since they both depend on v1 (already in a register) but not on each other.
```

**How to detect ILP opportunities:**
- Look for hot loops where the same variable is read and written every iteration
- Look for consecutive operations that all depend on the same value
- Profile with `pprof` to confirm the function is a bottleneck before optimizing
- Benchmark before and after: `go test -bench=. -count=10 | benchstat`

**When this matters:**
- Tight inner loops processing large slices (numerical computation, string processing)
- The benefit is CPU-architecture specific and compiler/hardware dependent
- Modern compilers sometimes apply this optimization automatically; verify with benchmarks

Why this matters: CPUs have multiple execution units that can handle arithmetic, memory reads, and branching in parallel — but only when instructions are independent. Data hazards force the CPU to serialize operations, leaving execution units idle. Restructuring hot loops to use local accumulator variables or split dependent computations into independent chains lets the CPU utilize its full parallel execution capacity. The effect is most pronounced in tight numerical loops on large data sets.
