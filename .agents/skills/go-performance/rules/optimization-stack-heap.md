## Understand Stack vs. Heap: Returning Pointers Forces Heap Allocation

**Impact: MEDIUM**

Go allocates variables on the stack (fast, self-cleaning) or the heap (requires GC, 10x+ slower). The compiler decides via **escape analysis**: if a variable's address outlives the function, it *escapes* to the heap. Unnecessary heap allocations pressure the GC and can dominate application performance.

**Stack vs. heap characteristics:**
- **Stack**: per-goroutine (2 KB initial, grows as needed), LIFO, self-cleaning, no GC involved, ~1 ns alloc
- **Heap**: shared across goroutines, requires GC to reclaim, GC can use 25% of CPU and pause application

**Sharing up → heap; sharing down → stack:**

```go
// Sharing UP: returning a pointer to a local variable — z escapes to the heap
//go:noinline
func sumPtr(x, y int) *int {
    z := x + y
    return &z  // z's address outlives sumPtr → compiler allocates z on heap
}

// Sharing DOWN: passing pointers from parent to child — a and b stay on the stack
//go:noinline
func sum(x, y *int) int {
    return *x + *y  // x and y are owned by caller — no escape
}

func main() {
    a := 3
    b := 2
    c := sum(&a, &b)  // a and b stay on main's stack frame
    _ = c
}
```

**Benchmark comparison:**

```go
var globalValue int
var globalPtr *int

func BenchmarkSumValue(b *testing.B) {
    b.ReportAllocs()
    var local int
    for i := 0; i < b.N; i++ {
        local = sumValue(i, i)  // Stack allocation: 0 allocs/op
    }
    globalValue = local
}

func BenchmarkSumPtr(b *testing.B) {
    b.ReportAllocs()
    var local *int
    for i := 0; i < b.N; i++ {
        local = sumPtr(i, i)  // Heap allocation: 1 alloc/op
    }
    globalValue = *local
}
// BenchmarkSumValue: 1.26 ns/op   0 allocs/op
// BenchmarkSumPtr:  14.84 ns/op   1 allocs/op  ← ~10x slower
```

**Variables that escape to the heap:**

```go
// 1. Returned pointer (sharing up)
func newFoo() *Foo { return &Foo{} }  // Escapes

// 2. Global variables (accessible by all goroutines)
var g *Foo
func set() { g = &Foo{} }  // Escapes

// 3. Pointer sent to a channel
ch <- &Foo{}  // Escapes

// 4. Variable too large for the stack
s := make([]int, n)  // Escapes if n is a variable (size unknown at compile time)
s := make([]int, 10) // May stay on stack (size known)

// 5. Backing array reallocated by append (may escape)
```

**Inspect escape analysis decisions:**

```bash
# See what escapes and why
go build -gcflags "-m=2" ./...
# Example output:
# ./main.go:12:2: z escapes to heap

# Use b.ReportAllocs() in benchmarks to count heap allocs per operation
```

Why this matters: Returning a pointer (e.g., `return &localVar`) is often written to "avoid a copy," but it actually forces a heap allocation that is 10x more expensive than a stack allocation. The GC must eventually collect all heap allocations, using up to 25% of available CPU. In data-intensive hot paths, GC pressure from unnecessary allocations can dominate performance. Favor value semantics unless sharing is semantically required. Use `go build -gcflags "-m=2"` and `b.ReportAllocs()` to audit heap allocations before optimizing.
