## Tune the GC With GOGC to Reduce GC Pressure Under Load

**Impact: MEDIUM**

Go's garbage collector (concurrent mark-and-sweep) is triggered by the `GOGC` environment variable. The default `GOGC=100` means a GC runs when the heap doubles. Understanding this lets you tune GC frequency for your workload, reducing stop-the-world pauses during sudden load spikes.

**How the GC is triggered:**

```bash
# GOGC=100 (default): GC runs when heap size doubles
# If heap is 128 MB after last GC, next GC triggers at 256 MB

# Lower GOGC = more frequent GC, less memory usage
GOGC=50 ./myapp   # GC triggers when heap grows 50% (more GC cycles, less peak memory)

# Higher GOGC = less frequent GC, more memory usage
GOGC=200 ./myapp  # GC triggers when heap doubles + 100% more (fewer GCs, higher peak)

# Disable GC entirely (useful for batch jobs, never for long-running services)
GOGC=off ./myapp

# Print GC traces to stderr
GODEBUG=gctrace=1 go test -bench=. -v
```

**Pre-allocate a minimum heap to reduce GC frequency at startup:**

```go
// For services with a known peak heap, pre-allocate to raise the GC trigger baseline
// Uses virtual memory (lazy allocation via mmap) — won't consume physical RAM until accessed
var min = make([]byte, 1_000_000_000)  // 1 GB reservation

// With GOGC=100 and 1 GB baseline:
// GC won't trigger until heap reaches 2 GB, instead of triggering at 256 MB
// This is effective when:
//   1. You know the expected peak heap size
//   2. Traffic patterns cause rapid heap growth (sudden spike scenario)
//   3. You want to reduce stop-the-world pauses during load spikes
```

**When to tune GOGC:**

```
Scenario 1: Steady gradual load increase
→ Keep GOGC=100 (default); GC frequency stays moderate

Scenario 2: Sudden traffic spike (0 → 1M users in minutes)
→ Bump GOGC to 200-400 to reduce GC cycles during the spike
→ Or pre-allocate a minimum heap matching expected peak

Scenario 3: Memory-constrained environment
→ Lower GOGC (50-80) to collect more aggressively

Scenario 4: Batch job (runs once, exits)
→ GOGC=off; GC overhead eliminated entirely
```

**GC concepts:**
- Mark stage: traverses all heap objects, marks live ones
- Sweep stage: deallocates unmarked objects
- Two stop-the-world phases per GC cycle (brief), then concurrent operation resumes
- Go GC can use up to 25% of available CPU capacity during concurrent phase
- `debug.FreeOSMemory()` forces a GC and returns free memory to OS (rarely needed)

**Detecting GC pressure:**

```bash
# Print GC events: shows heap size before/after, pause duration
GODEBUG=gctrace=1 ./myapp 2>&1 | grep "^gc"

# Use pprof heap profile to find what's allocating
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# In benchmarks, report allocs per operation
func BenchmarkFoo(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        // ...
    }
}
```

Why this matters: The Go GC is designed to be low-latency, but it still consumes CPU during collection and introduces brief stop-the-world pauses. For services experiencing sudden traffic spikes, frequent GC cycles (because the heap doubles many times rapidly) cause cascading latency increases visible to users. Tuning `GOGC` upward during known high-traffic periods, or pre-allocating a minimum heap baseline, can eliminate these cycles. Profile with `GODEBUG=gctrace=1` first to confirm GC frequency is actually the bottleneck.
