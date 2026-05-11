## Use Purposeful Channel Sizes — Default to 1 for Buffered Channels

**Impact: LOW**

Choosing a buffered channel size arbitrarily is a common mistake. The size affects backpressure, memory usage, and synchronization semantics. When in doubt, start with a size of 1 or use an unbuffered channel, and use other sizes only when there's a specific, documented reason.

**Incorrect:**

```go
// Magic number with no justification — why 40? Why not 50 or 1000?
ch := make(chan int, 40)

// Unbuffered channel when decoupling sender/receiver is needed:
ch := make(chan int)  // Sender blocks until receiver is ready — may not be intended
```

**Correct:**

```go
// Unbuffered: provides synchronization — sender blocks until receiver is ready
// Use when: you need guaranteed delivery or want to know when work was received
ch := make(chan int)

// Buffered size 1: minimal decoupling — allows sender to proceed without waiting
// Use as the default buffered channel size when unsure
ch := make(chan int, 1)

// Buffered size = pool size: for worker pool pattern
poolSize := runtime.GOMAXPROCS(0)
taskCh := make(chan Task, poolSize)  // Tied to the number of workers

// Buffered size = rate limit: for rate-limiting scenarios
const maxConcurrentRequests = 10
semaphore := make(chan struct{}, maxConcurrentRequests)

// Document any other size with a comment explaining the rationale
// Size 256: empirically determined from benchmark on production workload
ch := make(chan Event, 256)
```

Why this matters: An unbuffered channel provides synchronization guarantees (sender blocks until receiver receives). A buffered channel decouples sender and receiver but can lead to obscure deadlocks if the buffer fills and no one is reading. The minimum useful buffer size is 1. Larger sizes should be tied to concrete values (worker pool size, rate limits) or determined via benchmarks. Magic numbers like `make(chan int, 40)` are a code smell — always comment the rationale.
