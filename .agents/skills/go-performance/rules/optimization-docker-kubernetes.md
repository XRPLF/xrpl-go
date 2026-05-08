## Set GOMAXPROCS to Match Container CPU Quota

**Impact: MEDIUM**

Go sets `GOMAXPROCS` (the number of OS threads running Go code simultaneously) to the number of **host** logical CPUs, not the container's CPU limit. In Kubernetes, a pod with `cpu: 1000m` (1 core) running on an 8-core node will have `GOMAXPROCS=8`, causing CFS throttling and severe latency spikes.

**The problem — GOMAXPROCS defaults to host cores:**

```
Host machine: 8 logical CPUs
Kubernetes pod: cpu: 1000m (1 core limit = 100ms quota per 100ms period)

Go runtime sets GOMAXPROCS = 8 (host cores)
→ 8 OS threads created, all competing for CPU
→ After consuming 100ms of CPU time, CFS throttles the container
→ All 8 threads freeze until next 100ms period begins
→ Result: latency spikes, throughput degradation
```

**Why CFS throttling is so damaging:**

```
Timeline (100ms CFS period):
[0ms]     All 8 threads start running
[12ms]    Container has consumed its 100ms quota (8 threads × 12ms ≈ 96ms)
[12ms]    CFS THROTTLES the container — ALL goroutines freeze
[100ms]   CFS period resets, container unthrottled
[100-112ms] Normal operation resumes
...repeat

Effect: ~87% of each 100ms period the service is frozen.
Even with 1 core of CPU quota, the service behaves as if the host is overloaded.
```

**The fix — use automaxprocs:**

```go
// main.go
import (
    _ "go.uber.org/automaxprocs"  // Blank import: sets GOMAXPROCS at init time
)

func main() {
    // automaxprocs reads the Linux CFS quota from /proc and sets:
    // GOMAXPROCS = ceil(quota / period)
    // e.g., quota=100ms, period=100ms → GOMAXPROCS=1
    // e.g., quota=400ms, period=100ms → GOMAXPROCS=4
    // ...
}
```

```bash
go get go.uber.org/automaxprocs
```

**Manual alternative — set GOMAXPROCS explicitly:**

```go
import (
    "os"
    "runtime"
    "strconv"
)

func init() {
    if val := os.Getenv("GOMAXPROCS"); val != "" {
        if n, err := strconv.Atoi(val); err == nil && n > 0 {
            runtime.GOMAXPROCS(n)
        }
    }
}

// Then set in Kubernetes manifest:
// env:
//   - name: GOMAXPROCS
//     valueFrom:
//       resourceFieldRef:
//         resource: limits.cpu
//         divisor: "1"
```

**Diagnosing the problem:**

```bash
# Check current GOMAXPROCS at runtime
import "runtime"
fmt.Println(runtime.GOMAXPROCS(0))  // 0 = query without changing

# Check if throttling is occurring (Linux, inside container)
cat /sys/fs/cgroup/cpu/cpu.stat | grep throttled
# throttled_time 12345678  ← microseconds of throttling

# Kubernetes: check container CPU throttling metrics
# kubectl top pod --containers
# Or use Prometheus: container_cpu_cfs_throttled_seconds_total
```

**Kubernetes resource configuration reference:**

```yaml
resources:
  requests:
    cpu: "1"       # Scheduling hint: 1 core requested
  limits:
    cpu: "1"       # Hard limit: 1 core = 100ms quota per 100ms period

# With automaxprocs: GOMAXPROCS=1 (matches the limit)
# Without automaxprocs: GOMAXPROCS=8 (or whatever the node has)
```

Why this matters: `GOMAXPROCS` controls how many goroutines run in parallel. When set higher than the container's CPU quota allows, the Linux CFS scheduler throttles the entire container after the quota is exhausted — pausing all goroutines until the next scheduling period. This creates predictable latency spikes (often 300%+ tail latency) that are difficult to diagnose without understanding the CFS/GOMAXPROCS interaction. The fix is a single blank import of `go.uber.org/automaxprocs` in `main.go`, which reads the container CPU quota from `/proc` at startup and sets `GOMAXPROCS` accordingly.
