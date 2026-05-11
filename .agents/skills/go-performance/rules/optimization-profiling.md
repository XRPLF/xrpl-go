## Profile and Trace Go Applications Before Optimizing

**Impact: HIGH**

Go provides two essential diagnostics tools: `pprof` for profiling (CPU, heap, goroutines, blocking) and the execution tracer for visualizing runtime behavior. Never guess about performance bottlenecks — profile first to identify what actually needs optimization.

**Enabling pprof in a running service:**

```go
import (
    _ "net/http/pprof"  // Blank import registers /debug/pprof endpoints
    "net/http"
    "log"
)

func main() {
    // Your application code...
    log.Fatal(http.ListenAndServe(":8080", nil))
}
// Access profiles at: http://localhost:8080/debug/pprof/
```

**Profile types and when to use them:**

| Profile | Endpoint | Use when |
|---------|----------|----------|
| CPU | `/debug/pprof/profile?seconds=30` | CPU time is high; find hot functions |
| Heap | `/debug/pprof/heap?debug=0` | Memory is high; find allocation sources |
| Goroutine | `/debug/pprof/goroutine?debug=0` | Goroutine count is high; find leaks |
| Block | `/debug/pprof/block` | Goroutines block too long; find contention |
| Mutex | `/debug/pprof/mutex` | Mutex contention; find lock hot spots |

**Using pprof:**

```bash
# CPU profile during a benchmark
go test -bench=. -cpuprofile profile.out
go tool pprof -http=:8080 profile.out   # Opens browser with call graph

# Heap profile — force GC first for accurate data
curl http://localhost:8080/debug/pprof/heap?gc=1 -o heap1.out
# ... wait a few seconds ...
curl http://localhost:8080/debug/pprof/heap?gc=1 -o heap2.out
go tool pprof -http=:8080 -diff_base heap1.out heap2.out  # Compare for leaks

# Enable block profiling (must enable at runtime — disabled by default)
runtime.SetBlockProfileRate(1)  // 1 = record every blocking event
```

**Execution tracer — understand GC and goroutine scheduling:**

```bash
# Collect trace during a benchmark
go test -bench=. -v -trace=trace.out
go tool trace trace.out   # Opens browser with timeline visualization
```

**Reading pprof signals:**
- `runtime.mallocgc` dominates CPU → too many small heap allocations; use `sync.Pool`
- Channel/mutex operations dominate → contention; reduce lock scope or restructure
- `syscall.Read`/`syscall.Write` dominate → I/O bound; improve buffering

**Custom user-level traces:**

```go
import "runtime/trace"

ctx, task := trace.NewTask(context.Background(), "fibonacci")
trace.WithRegion(ctx, "main", func() {
    v = fibonacci(10)
})
task.End()
// Appears in go tool trace with duration distribution
```

**pprof best practices:**
- Enable only one profile at a time (CPU + heap simultaneously = erroneous results)
- CPU and heap profiling are safe in production (activated only when accessed, not continuous)
- Block and mutex profiling have overhead — enable selectively, use low rates in production
- `GODEBUG=gctrace=1` prints a line to stderr each time the GC runs

Why this matters: Optimizing code without profiling wastes effort and risks making things worse. CPU profiling (sample-based, per function) identifies hot code paths. Heap profiling shows allocation sources. The execution tracer (not sample-based, per goroutine) reveals GC behavior and goroutine scheduling problems invisible to the CPU profiler. Together they provide a complete picture of application performance. Use the execution tracer when concurrency or GC frequency is the bottleneck.
