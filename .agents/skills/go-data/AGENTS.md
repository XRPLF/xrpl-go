# Go Data Types & Control Flow

**Version 0.1.0**  
Agent Skills  
March 2026

> **Note:**  
> This document is mainly for agents and LLMs to follow when maintaining,  
> generating, or refactoring Go slices, maps, integers, strings, range loops, and control structures. Humans  
> may also find it useful, but guidance here is optimized for automation  
> and consistency by AI-assisted workflows.

---

## Abstract

Guidelines for Go data type handling, control flow patterns, and string manipulation. Covers slices, maps, integers, range loops, defer in loops, runes, and UTF-8 encoding.

---

## Table of Contents

1. [Data Types](#1-data-types) — **HIGH**
   - 1.1 [Beware of Append Side Effects with Shared Backing Arrays](#11-beware-of-append-side-effects-with-shared-backing-arrays)
   - 1.2 [Check for Integer Overflow in Critical Operations](#12-check-for-integer-overflow-in-critical-operations)
   - 1.3 [Check Slice Emptiness with len(), Not nil Comparison](#13-check-slice-emptiness-with-len-not-nil-comparison)
   - 1.4 [Handle Floating-Point Precision Correctly](#14-handle-floating-point-precision-correctly)
   - 1.5 [Initialize Maps with Expected Size](#15-initialize-maps-with-expected-size)
   - 1.6 [Know When == Works and When to Use reflect.DeepEqual](#16-know-when--works-and-when-to-use-reflectdeepequal)
   - 1.7 [Prefer Nil Slices Over Empty Slices](#17-prefer-nil-slices-over-empty-slices)
   - 1.8 [Prevent Memory Leaks When Slicing Large Slices](#18-prevent-memory-leaks-when-slicing-large-slices)
   - 1.9 [Understand Slice Length vs Capacity](#19-understand-slice-length-vs-capacity)
   - 1.10 [Understand That Maps Never Shrink](#110-understand-that-maps-never-shrink)
   - 1.11 [Use Clear Octal Literal Syntax](#111-use-clear-octal-literal-syntax)
   - 1.12 [Use copy Correctly: Destination Must Have Length Set](#112-use-copy-correctly-destination-must-have-length-set)
2. [Control Structures](#2-control-structures) — **MEDIUM**
   - 2.1 [Avoid defer in Tight Loops](#21-avoid-defer-in-tight-loops)
   - 2.2 [Break Inside switch/select Only Breaks the Innermost Statement](#22-break-inside-switchselect-only-breaks-the-innermost-statement)
   - 2.3 [Don't Store Pointers to Range Loop Variables](#23-dont-store-pointers-to-range-loop-variables)
   - 2.4 [Never Assume Map Iteration Order or Stability](#24-never-assume-map-iteration-order-or-stability)
   - 2.5 [Range Expression Is Evaluated Only Once](#25-range-expression-is-evaluated-only-once)
   - 2.6 [Range Loop Values Are Copies — Mutations Don't Affect the Original](#26-range-loop-values-are-copies--mutations-dont-affect-the-original)
3. [Strings](#3-strings) — **MEDIUM**
   - 3.1 [Avoid Unnecessary String-to-Byte-Slice Conversions](#31-avoid-unnecessary-string-to-byte-slice-conversions)
   - 3.2 [Know the Difference Between TrimRight and TrimSuffix](#32-know-the-difference-between-trimright-and-trimsuffix)
   - 3.3 [Substring Operations Can Cause Memory Leaks](#33-substring-operations-can-cause-memory-leaks)
   - 3.4 [Understand Runes: len() Returns Bytes, Not Characters](#34-understand-runes-len-returns-bytes-not-characters)

---

## 1. Data Types

**Impact: HIGH**

Understanding Go's data types prevents subtle bugs with slices, maps, integers, and strings. Common mistakes include integer overflow, slice capacity misuse, nil slice confusion, and map initialization. These patterns ensure correct handling of Go's built-in types.

### 1.1 Beware of Append Side Effects with Shared Backing Arrays

**Impact: HIGH**

Slices created by slicing a larger slice share the same backing array. Appending to one may modify the other if the slice hasn't grown beyond the original capacity. This leads to hard-to-debug data corruption.

**Incorrect:**

```go
func main() {
    original := []int{1, 2, 3, 4, 5}

    // Both share the same backing array
    first := original[:3]   // [1, 2, 3], cap=5
    second := original[3:]  // [4, 5], cap=2

    // Appending to first when len < cap modifies original's backing array!
    first = append(first, 99)
    fmt.Println(original)  // [1, 2, 3, 99, 5] -- original is mutated!
    fmt.Println(first)     // [1, 2, 3, 99]

    // Functions receiving sub-slices can mutate caller's data
    func modifySlice(s []int) {
        s[0] = 999  // Modifies backing array — visible to caller!
    }
}
```

**Correct:**

```go
func main() {
    original := []int{1, 2, 3, 4, 5}

    // Option 1: Use full slice expression to limit capacity
    // s[low:high:max] — cap of result = max - low
    first := original[:3:3]  // cap=3, append will allocate new array
    first = append(first, 99)
    fmt.Println(original)    // [1, 2, 3, 4, 5] -- original unchanged ✓
    fmt.Println(first)       // [1, 2, 3, 99]

    // Option 2: Copy the slice to get independent backing array
    firstCopy := make([]int, 3)
    copy(firstCopy, original[:3])
    firstCopy = append(firstCopy, 99)
    fmt.Println(original)    // [1, 2, 3, 4, 5] -- unchanged ✓

    // Option 3: append idiom for independent copy
    firstCopy2 := append([]int(nil), original[:3]...)
}

// When returning sub-slices from functions, always copy
func GetFirst3(data []int) []int {
    if len(data) < 3 {
        return append([]int(nil), data...)
    }
    return append([]int(nil), data[:3]...)  // Independent copy
}
```

Why this matters: Sharing backing arrays is an optimization in Go's slice design, but it creates a hidden coupling between slices. When you pass a sub-slice to a function or store it in a struct, mutations can propagate unexpectedly. The full slice expression `s[low:high:max]` caps the capacity so that the next append triggers a new allocation and breaks the sharing.

### 1.2 Check for Integer Overflow in Critical Operations

**Impact: HIGH**

Integer overflow occurs when arithmetic operations produce results outside the type's range. Go doesn't panic on overflow—it silently wraps around. This causes severe bugs in financial calculations, resource limits, and security checks.

**Incorrect:**

```go
func AllocateBuffer(count int32, size int32) []byte {
    total := count * size  // Can overflow!
    return make([]byte, total)
}

// AllocateBuffer(1000000, 5000) overflows int32, wraps to negative,
// then make() panics with negative size
```

**Correct:**

```go
func AllocateBuffer(count int32, size int32) ([]byte, error) {
    if count > 0 && size > math.MaxInt32/count {
        return nil, fmt.Errorf("buffer size overflow: %d * %d", count, size)
    }

    total := int64(count) * int64(size)
    if total > math.MaxInt32 {
        return nil, fmt.Errorf("buffer too large: %d bytes", total)
    }

    return make([]byte, total), nil
}
```

Why this matters: Overflow bugs cause security vulnerabilities (buffer overflows), financial errors (incorrect totals), and resource exhaustion (negative or wrapped values). These bugs are subtle because Go doesn't warn you.

Detection: Use `int64` for calculations involving multiplication or addition of user input. Check bounds before operations when using smaller integer types. Consider using `math/big` for arbitrary precision when needed.

### 1.3 Check Slice Emptiness with len(), Not nil Comparison

**Impact: LOW**

Testing `s == nil` to check if a slice is empty misses non-nil empty slices (`make([]T, 0)`). Use `len(s) == 0` to reliably check if a slice has no elements, regardless of whether it's nil or just empty.

**Incorrect:**

```go
func ProcessItems(items []string) {
    if items == nil {
        fmt.Println("no items")
        return
    }
    // items might still be empty! (non-nil, length 0)
    for _, item := range items {
        process(item)
    }
}

// This works fine but...
ProcessItems(nil)          // "no items" ✓
ProcessItems([]string{})   // Enters the loop — no output ✓ (loop runs 0 times)
// Actually range on empty slice is fine, but the intent check is wrong:
func HasItems(items []string) bool {
    return items != nil   // Returns false for nil, but also false for []string{}!
}
```

**Correct:**

```go
func ProcessItems(items []string) {
    if len(items) == 0 {
        fmt.Println("no items")
        return
    }
    for _, item := range items {
        process(item)
    }
}

func HasItems(items []string) bool {
    return len(items) > 0  // Works for both nil and non-nil empty slices
}

// Both cases handled correctly:
HasItems(nil)          // false ✓
HasItems([]string{})   // false ✓
HasItems([]string{"a"}) // true ✓
```

Why this matters: `len(nil)` returns 0 in Go, so `len(s) == 0` works correctly for both nil and empty slices. Using `s == nil` introduces a distinction that usually doesn't matter for the caller's intent — they typically want to know "are there elements?", not "was this explicitly initialized?". Using `len` is both correct and communicates intent clearly.

### 1.4 Handle Floating-Point Precision Correctly

**Impact: HIGH**

Floating-point numbers are approximations. They cannot represent all decimal values exactly, arithmetic operations accumulate rounding errors, and comparing float results with `==` produces unexpected results. Ignoring this leads to subtle bugs in financial calculations, physics simulations, and any domain requiring precise numeric results.

**Incorrect:**

```go
// Direct equality comparison of floats is unreliable
a := 0.1 + 0.2
if a == 0.3 {  // false! a is 0.30000000000000004
    fmt.Println("equal")
}

// Order of operations affects precision
func Sum(prices []float64) float64 {
    var total float64
    for _, p := range prices {
        total += p  // Adding many small values to a large total loses precision
    }
    return total
}

// Naive financial calculations
price := 1.005
rounded := math.Round(price*100) / 100  // May not give 1.01 as expected
```

**Correct:**

```go
// Use epsilon comparison for floats
const epsilon = 1e-9

func FloatEquals(a, b float64) bool {
    return math.Abs(a-b) < epsilon
}

a := 0.1 + 0.2
if FloatEquals(a, 0.3) {
    fmt.Println("equal")  // Works correctly
}

// For financial calculations, use integer arithmetic (cents) or decimal library
type Money struct {
    cents int64  // Store as integer cents to avoid float issues
}

func (m Money) Add(other Money) Money {
    return Money{cents: m.cents + other.cents}
}

// Pairwise summation or sorting before summing improves precision
func SumSorted(values []float64) float64 {
    sorted := make([]float64, len(values))
    copy(sorted, values)
    sort.Float64s(sorted)  // Sort ascending: add small values first
    var total float64
    for _, v := range sorted {
        total += v
    }
    return total
}
```

Why this matters: IEEE 754 float64 has 53 bits of mantissa, giving about 15-16 significant decimal digits. Operations like `0.1 + 0.2` don't yield exact results because 0.1 and 0.2 cannot be represented exactly in binary. For money: use integer cents. For comparisons: use epsilon. For summing many values: sort first or use compensated summation.

### 1.5 Initialize Maps with Expected Size

**Impact: HIGH**

Like slices, maps can be pre-sized at creation time. When a map grows beyond its load factor (6.5 elements per bucket), Go allocates new buckets and rehashes all keys — an O(n) operation. Providing an initial size hint avoids repeated growth operations and is significantly faster.

**Incorrect:**

```go
// Map starts with 1 bucket and grows repeatedly as elements are added
func buildIndex(words []string) map[string]int {
    index := make(map[string]int)  // No size hint
    for i, word := range words {
        index[word] = i
    }
    return index
}

// Benchmark: inserting 1 million elements
// Without size: ~227 ms/op (repeated bucket allocations and rehashing)
```

**Correct:**

```go
// Provide the expected number of elements as a size hint
func buildIndex(words []string) map[string]int {
    index := make(map[string]int, len(words))  // Pre-size with expected count
    for i, word := range words {
        index[word] = i
    }
    return index
}

// Benchmark: inserting 1 million elements
// With size: ~91 ms/op — about 60% faster

// Note: the size hint is not a maximum — you can always add more elements
// The hint just tells the runtime how many buckets to pre-allocate
```

Why this matters: A map grows by doubling its bucket count when the average bucket load exceeds ~6.5. Each growth requires rehashing all existing keys. For large maps, this happens logarithmically many times during construction, causing significant allocation overhead. Providing a size hint with `make(map[K]V, n)` pre-allocates enough buckets, reducing or eliminating growth operations.

Unlike slices, maps only accept a single size argument (no separate capacity). The size is a hint — if you provide `n`, Go allocates enough buckets to hold roughly `n` elements without growing.

### 1.6 Know When == Works and When to Use reflect.DeepEqual

**Impact: MEDIUM**

Go's `==` operator only works on *comparable* types. Slices and maps are not comparable — using `==` on structs containing them causes a compile error, and comparing `any`-typed values containing them causes a runtime panic. Know the alternatives.

**Incorrect:**

```go
type Customer struct {
    ID         string
    Operations []float64  // slice — not comparable
}

cust1 := Customer{ID: "x", Operations: []float64{1.0}}
cust2 := Customer{ID: "x", Operations: []float64{1.0}}

// Compile error: invalid operation: cust1 == cust2
// (struct containing []float64 cannot be compared)
fmt.Println(cust1 == cust2)

// Worse: compiles but panics at runtime when using any
var a any = cust1
var b any = cust2
fmt.Println(a == b)  // panic: runtime error: comparing uncomparable type
```

**Correct:**

```go
// Option 1: reflect.DeepEqual — works on anything, ~100x slower than ==
fmt.Println(reflect.DeepEqual(cust1, cust2))  // true
// Note: DeepEqual distinguishes nil from empty slices

// Option 2: Custom comparison method — fast, explicit, correct
func (a Customer) Equal(b Customer) bool {
    if a.ID != b.ID {
        return false
    }
    if len(a.Operations) != len(b.Operations) {
        return false
    }
    for i := range a.Operations {
        if a.Operations[i] != b.Operations[i] {
            return false
        }
    }
    return true
}
// Custom method is ~96x faster than reflect.DeepEqual

// Option 3: Use go-cmp or testify in tests
// assert.Equal(t, cust1, cust2)  // testify handles deep equality
```

Comparable types (can use `==`): booleans, numbers, strings, pointers, channels, interfaces, arrays (of comparable elements), structs (where all fields are comparable).

Non-comparable types (cannot use `==`): slices, maps, functions.

Use `reflect.DeepEqual` for correctness in tests where performance isn't critical. Implement custom `Equal` methods for production code where performance matters. Check `bytes.Compare` and similar standard library functions before implementing your own.

### 1.7 Prefer Nil Slices Over Empty Slices

**Impact: MEDIUM**

Go has two "empty" slice states: `nil` (zero value, no backing array) and an empty non-nil slice (`make([]T, 0)` or `[]T{}`). These behave identically for `append` and `len`/`cap`, but differ in JSON serialization and some reflect operations. Prefer the nil slice unless you have a specific reason for non-nil.

**Incorrect:**

```go
func GetActiveUsers() []User {
    users := make([]User, 0)  // Pre-allocated empty slice, not nil

    rows, err := db.Query(...)
    if err != nil {
        return users  // Returns [] when serialized to JSON
    }
    // ...
    return users
}

// JSON result when no users: []
// This differs from the nil case: null

// Also problematic:
func CollectErrors(items []Item) []string {
    errs := []string{}  // Non-nil empty slice
    for _, item := range items {
        if err := validate(item); err != nil {
            errs = append(errs, err.Error())
        }
    }
    return errs  // Returns [] even if no errors — hard to check "were there any?"
}
```

**Correct:**

```go
func GetActiveUsers() []User {
    var users []User  // nil slice — zero value

    rows, err := db.Query(...)
    if err != nil {
        return users  // Returns null in JSON — signals "no data" clearly
    }
    // ...
    return users
}

func CollectErrors(items []Item) []string {
    var errs []string  // nil slice
    for _, item := range items {
        if err := validate(item); err != nil {
            errs = append(errs, err.Error())
        }
    }
    return errs  // nil if no errors — callers can check errs != nil
}
```

Why this matters: `encoding/json` marshals `nil` slices as `null` and empty non-nil slices as `[]`. API clients receive different signals. Callers checking `if errs != nil` get useful information from nil slices. Use `make([]T, 0, cap)` only when you know the capacity upfront for performance, or when a non-nil empty slice is explicitly required by an API.

### 1.8 Prevent Memory Leaks When Slicing Large Slices

**Impact: HIGH**

When you slice a large slice or array, the result shares the original backing array. Even if the original slice is no longer referenced, the GC cannot reclaim the unused elements because the sub-slice still holds a reference to the entire backing array. This causes significant memory leaks in long-running programs.

**Incorrect:**

```go
// Receives large messages (e.g., 1 million bytes each)
func getMessageType(msg []byte) []byte {
    return msg[:5]  // 5-byte result holds reference to entire 1M backing array!
}

func consumeMessages() {
    var messageTypes [][]byte
    for {
        msg := receiveMessage()  // 1 million bytes
        // Each stored type holds a 1M-byte backing array in memory
        messageTypes = append(messageTypes, getMessageType(msg))
        // After 1,000 messages: ~1 GB held in memory instead of ~5 KB
    }
}

// Also leaks with structs containing pointer fields
func keepFirstTwo(foos []Foo) []Foo {
    return foos[:2]  // foos[2:] elements with pointer fields aren't GC'd
}
```

**Correct:**

```go
// Solution 1: Copy the sub-slice to break the backing array reference
func getMessageType(msg []byte) []byte {
    msgType := make([]byte, 5)
    copy(msgType, msg)
    return msgType  // Independent 5-byte slice; original msg can be GC'd
}

// Solution 2: For structs with pointer fields, nil out unused elements
func keepFirstTwo(foos []Foo) []Foo {
    // Nil out pointer fields so GC can collect referenced memory
    for i := 2; i < len(foos); i++ {
        foos[i].data = nil  // Or foos[i] = Foo{} to zero the whole struct
    }
    return foos[:2]
}

// Note: full slice expression msg[:5:5] does NOT fix the leak —
// the backing array is still referenced and won't be GC'd
```

Why this matters: A sub-slice's capacity retains the entire original backing array in memory. With 1,000 messages of 1 MB each, storing only the first 5 bytes via slicing uses ~1 GB instead of ~5 KB. Always copy when you need to retain a small portion of a large slice. For slices of structs with pointer fields, nil out the excluded elements' pointer fields.

### 1.9 Understand Slice Length vs Capacity

**Impact: HIGH**

Slices have both length (`len`) and capacity (`cap`). Length is the number of elements; capacity is the size of the underlying array. Confusing these leads to unexpected slice behavior, data loss, and shared backing arrays causing mysterious mutations.

**Incorrect:**

```go
func ProcessBatch(items []int) []int {
    results := make([]int, 10)  // Length 10, capacity 10
    for i, item := range items {
        results[i] = item * 2  // Panics if len(items) > 10!
    }
    return results
}
```

**Correct:**

```go
func ProcessBatch(items []int) []int {
    results := make([]int, 0, len(items))  // Length 0, capacity len(items)
    for _, item := range items {
        results = append(results, item*2)  // Safe, preallocated
    }
    return results
}
```

Why this matters: `make([]T, n)` creates a slice with `n` zero elements. `make([]T, 0, n)` creates an empty slice with space for `n` elements. The first form causes index-out-of-bounds panics or overwrites zeros. Slicing operations also preserve the underlying array—slicing `s[0:5]` from a 100-element slice still references all 100 elements, preventing garbage collection.

Pattern: Use `make([]T, 0, n)` when building slices with append. Use `make([]T, n)` when setting elements by index. Be aware that `slice[i:j]` shares the backing array with the original slice.

### 1.10 Understand That Maps Never Shrink

**Impact: HIGH**

Go maps can only grow — they never shrink their bucket count. When you add 1 million elements and then delete them all, the map retains all its allocated buckets. This causes permanent high memory consumption in applications with spiky load patterns.

**Incorrect:**

```go
// Cache that grows during peak traffic and never releases memory
type SessionCache struct {
    mu      sync.Mutex
    sessions map[int][128]byte
}

func (c *SessionCache) Add(id int, data [128]byte) {
    c.mu.Lock()
    c.sessions[id] = data
    c.mu.Unlock()
}

func (c *SessionCache) Remove(id int) {
    c.mu.Lock()
    delete(c.sessions, id)  // Frees the value but NOT the bucket
    c.mu.Unlock()
}
// After Black Friday peak: map held 2M entries (461 MB)
// After removing all entries and GC: still 293 MB allocated (buckets remain!)
```

**Correct:**

```go
// Solution 1: Periodically recreate the map to release buckets
func (c *SessionCache) Compact() {
    c.mu.Lock()
    defer c.mu.Unlock()
    newMap := make(map[int][128]byte, len(c.sessions))
    for k, v := range c.sessions {
        newMap[k] = v
    }
    c.sessions = newMap
    // Old map is now GC-eligible; bucket count matches current size
}

// Solution 2: Store pointers instead of values
// map[int]*[128]byte uses much less bucket space when values are large
// After removing all entries: only pointer-sized slots remain in buckets
type SessionCache struct {
    mu       sync.Mutex
    sessions map[int]*[128]byte
}

// Comparison for 1M elements then GC:
// map[int][128]byte:  add=461MB, remove+GC=293MB (still large)
// map[int]*[128]byte: add=182MB, remove+GC=38MB  (buckets freed)
```

Why this matters: Go's map implementation uses a `B` field tracking the number of buckets as a power of 2. After adding 1M elements, `B=18` (262,144 buckets). Deleting all elements zeroes the bucket slots but `B` stays at 18. The bucket array itself is never freed. For maps with large values, storing pointers reduces both peak memory and post-deletion memory significantly.

### 1.11 Use Clear Octal Literal Syntax

**Impact: LOW**

In Go, an integer literal starting with `0` (zero) is an octal (base-8) number. This is easy to miss at a glance and causes subtle bugs when you expect decimal values. Go 1.13 introduced the `0o` prefix specifically to make octal intent explicit.

**Incorrect:**

```go
// 0755 looks like decimal 755 but is octal 493
os.Mkdir("secrets", 0755)     // permission 755 octal = rwxr-xr-x

// Reading this constant, would you know it's octal?
const permissions = 0600      // Actually 384 decimal

// Easy to make a mistake
timeout := 010                // "ten seconds"? No — this is 8!
```

**Correct:**

```go
// 0o prefix makes octal explicit and unmistakable
os.Mkdir("secrets", 0o755)    // Clearly octal file permissions

const permissions = 0o600     // Reader immediately knows: octal

// Never use leading zero for decimal integers
timeout := 10                 // This is decimal 10

// Binary and hex also have clear prefixes
flags := 0b10110011           // binary
mask := 0xFF                  // hexadecimal
```

Why this matters: `010` and `10` look nearly identical but have different values (8 vs 10). This is especially dangerous with file permissions where `0777`, `0755`, `0644` are common values — developers may type them as `777`, `755`, `644` (decimal) and get completely wrong permission bits.

The Go vet tool does not warn about leading-zero octals, so there's no automatic safety net. Use `0o` for all octal literals and code reviewers will immediately recognize the intent.

### 1.12 Use copy Correctly: Destination Must Have Length Set

**Impact: MEDIUM**

The built-in `copy(dst, src)` copies `min(len(dst), len(src))` elements. If the destination slice has zero length (even if it has non-zero capacity), nothing is copied. This is a frequent source of bugs when developers allocate capacity but forget to set length.

**Incorrect:**

```go
src := []int{1, 2, 3, 4, 5}

// Bug: make([]int, 0, len(src)) creates len=0, cap=5
dst := make([]int, 0, len(src))
n := copy(dst, src)
fmt.Println(n, dst)  // Prints: 0 []  -- nothing was copied!

// Bug: nil dst
var dst2 []int
copy(dst2, src)  // Copies nothing, no error
```

**Correct:**

```go
src := []int{1, 2, 3, 4, 5}

// Correct: make([]int, len(src)) sets both length and capacity
dst := make([]int, len(src))
n := copy(dst, src)
fmt.Println(n, dst)  // Prints: 5 [1 2 3 4 5] ✓

// Alternative: append to copy (idiomatic)
dst2 := append([]int(nil), src...)
fmt.Println(dst2)   // [1 2 3 4 5] ✓

// Partial copy: copy first 3 elements
dst3 := make([]int, 3)
copy(dst3, src)
fmt.Println(dst3)   // [1 2 3]

// Copy into a sub-slice
dst4 := make([]int, 10)
copy(dst4[2:], src)  // Copies src into dst4 starting at index 2
fmt.Println(dst4)    // [0 0 1 2 3 4 5 0 0 0]
```

Why this matters: The `copy` function doesn't resize the destination — that's `append`'s job. `make([]int, 0, cap)` creates a slice with length 0, not `cap`. Always use `make([]T, length)` when you want to `copy` into it. If you want a defensive copy of a slice, `append([]T(nil), src...)` is idiomatic and harder to get wrong.

---

## 2. Control Structures

**Impact: MEDIUM**

Go's control structures have nuances that catch developers off guard. Range loop variable reuse, defer in loops, break statement behavior, and map iteration ordering can cause bugs. These patterns help you avoid control flow pitfalls.

### 2.1 Avoid defer in Tight Loops

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

**Alternative: immediate cleanup**

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

### 2.2 Break Inside switch/select Only Breaks the Innermost Statement

**Impact: MEDIUM**

A `break` statement terminates the innermost `for`, `switch`, or `select` statement. When a `switch` or `select` is nested inside a `for` loop, `break` exits the `switch`/`select`, not the loop. This is a frequent source of subtle infinite loops.

**Incorrect:**

```go
// Trying to break the for loop from inside a switch
for i := 0; i < 5; i++ {
    fmt.Printf("%d ", i)
    switch i {
    case 2:
        break  // BUG: breaks the switch, not the for loop!
    }
}
// Prints: 0 1 2 3 4  (loop runs all 5 iterations)

// Same problem with select inside a for loop
for {
    select {
    case msg := <-ch:
        process(msg)
    case <-ctx.Done():
        break  // BUG: breaks the select, not the for loop — infinite loop!
    }
}
```

**Correct:**

```go
// Use a labeled break to exit the for loop
loop:
    for i := 0; i < 5; i++ {
        fmt.Printf("%d ", i)
        switch i {
        case 2:
            break loop  // Breaks the for loop labeled "loop"
        }
    }
// Prints: 0 1 2

// Same for select inside a for loop
loop:
    for {
        select {
        case msg := <-ch:
            process(msg)
        case <-ctx.Done():
            break loop  // Breaks the for loop, not the select
        }
    }

// Alternative: use return if inside a function
func process(ctx context.Context, ch <-chan int) {
    for {
        select {
        case msg := <-ch:
            handle(msg)
        case <-ctx.Done():
            return  // Returns from the function entirely
        }
    }
}
```

Why this matters: This is a known Go gotcha. `break` terminates the innermost statement — and `switch`/`select` count as statements. Labels are idiomatic in Go (used in the standard library's `net/http` package) and are the correct tool here. The label name should describe the loop's purpose (e.g., `readlines:`, `loop:`) for clarity.

### 2.3 Don't Store Pointers to Range Loop Variables

**Impact: CRITICAL**

Range loop variables are reused across iterations. Taking their address creates pointers that all point to the same memory location, causing all your stored pointers to reference the last element.

**Incorrect:**

```go
func CollectUsers(names []string) []*User {
    var users []*User
    for _, name := range names {
        user := User{Name: name}
        users = append(users, &user)  // All point to same variable!
    }
    return users  // All users have the last name!
}
```

**Correct:**

```go
func CollectUsers(names []string) []*User {
    users := make([]*User, 0, len(names))
    for _, name := range names {
        user := User{Name: name}
        users = append(users, &user)  // Creates new user each iteration
    }
    return users
}

// Or better - use index:
func CollectUsers(names []string) []*User {
    users := make([]*User, len(names))
    for i := range names {
        users[i] = &User{Name: names[i]}
    }
    return users
}
```

Why this matters: This bug is subtle and common. The range loop reuses the same `user` variable for each iteration. Taking `&user` gives you the address of this shared variable, so all your pointers end up pointing to it. After the loop, it contains the last value, making all pointers reference the same data.

Solution: Declare variables inside the loop body (as shown in Correct version 1), use the index to access the original slice, or in Go 1.22+, the loop variable is implicitly copied per iteration. Always use the `-race` flag during testing to catch these issues.

### 2.4 Never Assume Map Iteration Order or Stability

**Impact: MEDIUM**

Go maps have intentionally non-deterministic iteration order. The order changes between runs and even between iterations of the same map in the same program. Additionally, inserting elements into a map during iteration may or may not be visited — the behavior is unspecified.

**Incorrect:**

```go
// Assuming alphabetical or insertion order
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k := range m {
    fmt.Print(k)  // Output varies: "abc", "bac", "cab", etc.
}

// Assuming stable order across iterations
for i := 0; i < 2; i++ {
    for k := range m {
        fmt.Print(k)
    }
    fmt.Println()
}
// Might print:
// zdyaec
// czyade  (different order each time)

// Assuming inserted elements are visited
m2 := map[int]bool{0: true, 1: false, 2: true}
for k, v := range m2 {
    if v {
        m2[10+k] = true  // May or may not be visited — non-deterministic
    }
}
// Result varies between runs
```

**Correct:**

```go
// If order matters, collect keys and sort them
keys := make([]string, 0, len(m))
for k := range m {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
    fmt.Printf("%s: %d\n", k, m[k])
}

// If you need to update a map based on iteration, use a copy
m2Copy := copyMap(m2)
for k, v := range m2 {
    if v {
        m2Copy[10+k] = true  // Update the copy, iterate the original
    }
}
```

Why this matters: Go intentionally randomizes map iteration to prevent developers from relying on ordering (a design choice made explicit in the spec). An element added during iteration "may be produced during the iteration or skipped" — both are valid. Never rely on map ordering for correctness; use sorted slices or ordered data structures when order matters.

### 2.5 Range Expression Is Evaluated Only Once

**Impact: MEDIUM**

The expression in a `range` statement is evaluated exactly once, before the loop begins — it's copied to a temporary variable. Changes to the original slice/channel/array during the loop are not seen by the range iterator, because it's working on the copy.

**Incorrect:**

```go
// Trying to process all elements including dynamically added ones
s := []int{0, 1, 2}
for range s {
    s = append(s, 10)  // BUG: adding elements doesn't extend the loop
}
fmt.Println(s)  // [0 1 2 10 10 10] — loop ran exactly 3 times, not forever

// Trying to switch channels mid-loop
ch := ch1
for v := range ch {     // range evaluates ch once — copies ch1's value
    fmt.Println(v)
    ch = ch2            // Changing ch has no effect — loop still reads ch1
}
```

**Correct:**

```go
// If you need to process dynamically-added elements, track length manually
s := []int{0, 1, 2}
for i := 0; i < len(s); i++ {   // len(s) re-evaluated each iteration
    s = append(s, 10)            // Infinite loop! len grows each iteration
    // Don't actually do this — but this is the difference
}

// For arrays: range over pointer to avoid copying large array
a := [3]int{0, 1, 2}
a[2] = 10
for i, v := range &a {  // Range over pointer — sees live array values
    if i == 2 {
        fmt.Println(v)   // Prints 10, not 2
    }
}

// For channels: don't reassign the channel variable to switch targets
// Use select with multiple channels instead
```

Why this matters: `range` copies its expression once. For slices this means a snapshot of (ptr, len, cap). Appending during the loop may reallocate the backing array, but the range iterator still uses the original length. For arrays, range copies the entire array. Use `range &arr` to avoid the copy and see live updates. For channels, reassigning the variable doesn't switch what's being ranged over.

### 2.6 Range Loop Values Are Copies — Mutations Don't Affect the Original

**Impact: MEDIUM**

In a `range` loop, the value variable is a copy of the element, not a reference to it. Mutating the value variable does not modify the original slice or map element. This is one of the most common Go surprises for developers coming from other languages.

**Incorrect:**

```go
type Account struct {
    Balance float64
}

accounts := []Account{
    {Balance: 100},
    {Balance: 200},
    {Balance: 300},
}

// BUG: a is a copy of each element — mutations don't affect the slice
for _, a := range accounts {
    a.Balance += 1000  // Modifies the copy, not accounts[i]
}

fmt.Println(accounts)  // [{100} {200} {300}] — unchanged!
```

**Correct:**

```go
// Option 1: Use the index to access the original element
for i := range accounts {
    accounts[i].Balance += 1000
}

// Option 2: Classic for loop
for i := 0; i < len(accounts); i++ {
    accounts[i].Balance += 1000
}

// Option 3: Slice of pointers (pointer copy points to same struct)
accounts := []*Account{{Balance: 100}, {Balance: 200}, {Balance: 300}}
for _, a := range accounts {
    a.Balance += 1000  // a is a pointer copy — but still points to original
}
// Note: pointer slices can be less cache-friendly (CPU prefetching)
```

Why this matters: Go's rule is simple — everything is a copy when assigned. The range value variable is assigned a copy each iteration. For structs you want to mutate, use the index or a pointer. The same applies to maps: `for k, v := range m { v.Field = x }` won't modify the map values if they're structs.

---

## 3. Strings

**Impact: MEDIUM**

String handling in Go requires understanding runes, UTF-8 encoding, and efficient concatenation. Mistakes with string iteration, trim functions, and conversions lead to bugs and performance issues. These patterns cover proper string manipulation.

### 3.1 Avoid Unnecessary String-to-Byte-Slice Conversions

**Impact: MEDIUM**

Most I/O in Go works with `[]byte`, not `string`. Converting `[]byte` → `string` → `[]byte` just to use string functions is wasteful — each conversion allocates a new copy. The `bytes` package mirrors the `strings` package and operates directly on `[]byte`.

**Incorrect:**

```go
// Reading from io.Reader gives []byte, but we convert to string to use strings package
func sanitize(reader io.Reader) ([]byte, error) {
    b, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }

    // Unnecessary: []byte → string → []byte (2 allocations!)
    s := string(b)
    s = strings.TrimSpace(s)
    return []byte(s), nil
}

// Also wasteful: converting just to check a condition
func hasPrefix(data []byte, prefix string) bool {
    return strings.HasPrefix(string(data), prefix)  // Extra allocation
}
```

**Correct:**

```go
// Use bytes package — same operations, works on []byte directly
func sanitize(reader io.Reader) ([]byte, error) {
    b, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }

    return bytes.TrimSpace(b), nil  // No extra allocations
}

func hasPrefix(data []byte, prefix string) bool {
    return bytes.HasPrefix(data, []byte(prefix))
}

// bytes package mirrors strings package:
// strings.Contains  → bytes.Contains
// strings.Count     → bytes.Count
// strings.Split     → bytes.Split
// strings.TrimSpace → bytes.TrimSpace
// strings.Index     → bytes.Index
// strings.Replace   → bytes.Replace
```

Why this matters: Converting `[]byte` to `string` always copies the data (strings are immutable in Go, so a new allocation is required). If your entire workflow can stay in `[]byte`, you avoid these copies. When working with `io.Reader`, HTTP bodies, file contents, or any I/O, prefer `bytes` package operations over converting to string unnecessarily.

### 3.2 Know the Difference Between TrimRight and TrimSuffix

**Impact: LOW**

`strings.TrimRight` and `strings.TrimSuffix` look similar but work completely differently. Confusing them produces unexpected results that are hard to debug because both functions compile and run without errors.

**Incorrect:**

```go
// Trying to remove the suffix "xo" from "123oxo"
result := strings.TrimRight("123oxo", "xo")
fmt.Println(result)  // "123" — not "123o"!

// Developer expected TrimRight to behave like TrimSuffix
// but TrimRight removed ALL trailing 'x' and 'o' characters

// Similarly confused on the left side:
result2 := strings.TrimLeft("oxo123", "ox")
fmt.Println(result2)  // "123" (removes all leading 'o' and 'x')
// vs
result3 := strings.TrimPrefix("oxo123", "ox")
fmt.Println(result3)  // "o123" (removes only the prefix "ox" once)
```

**Correct:**

```go
// TrimRight/TrimLeft: removes all trailing/leading runes that are IN the set
strings.TrimRight("123oxo", "xo")    // "123"   — removes o, then x, then o
strings.TrimLeft("oxo123", "ox")     // "123"   — removes o, then x, then o
strings.Trim("oxo123oxo", "ox")      // "123"   — TrimLeft + TrimRight

// TrimSuffix/TrimPrefix: removes exactly the given substring (once)
strings.TrimSuffix("123oxo", "xo")  // "123o"  — removes "xo" suffix once
strings.TrimSuffix("123oxo", "*xo")  // "123oxo" — no match, unchanged
strings.TrimPrefix("oxo123", "ox")  // "o123"  — removes "ox" prefix once

// Rule of thumb:
// TrimRight/Left = remove characters from a SET
// TrimSuffix/Prefix = remove an exact SUBSTRING
```

Why this matters: `TrimRight("123oxo", "xo")` iterates backward and removes any rune that appears in the set `{'x', 'o'}` until it finds one that doesn't — giving `"123"`. `TrimSuffix("123oxo", "xo")` checks if the string ends with exactly `"xo"` and removes it once — giving `"123o"`. These are fundamentally different operations despite similar names.

### 3.3 Substring Operations Can Cause Memory Leaks

**Impact: HIGH**

String substring operations (`s[low:high]`) share the same backing byte array as the original string. If you store a small substring extracted from a large string, the entire large string's backing array stays in memory as long as the substring is referenced.

**Incorrect:**

```go
// Log messages are large (potentially thousands of bytes)
// We only want to keep the 36-byte UUID prefix
func (s *store) handleLog(log string) error {
    if len(log) < 36 {
        return errors.New("log is not correctly formatted")
    }

    // BUG: uuid shares backing array with the full log string!
    uuid := log[:36]
    s.store(uuid)  // Stores uuid, but keeps entire log in memory
    return nil
}

// After caching 1,000 UUIDs from 10KB log messages:
// Expected memory: ~36 KB
// Actual memory: ~10 MB (entire log strings kept alive)
```

**Correct:**

```go
// Option 1: Force a copy using []byte round-trip (works in all Go versions)
func (s *store) handleLog(log string) error {
    if len(log) < 36 {
        return errors.New("log is not correctly formatted")
    }

    uuid := string([]byte(log[:36]))  // Independent copy — 36 bytes only
    s.store(uuid)
    return nil
}

// Option 2: Use strings.Clone (Go 1.20+) — cleaner, same effect
func (s *store) handleLog(log string) error {
    if len(log) < 36 {
        return errors.New("log is not correctly formatted")
    }

    uuid := strings.Clone(log[:36])  // Independent copy
    s.store(uuid)
    return nil
}
```

Why this matters: In Go's implementation, a substring creates a new string header (pointer + length) pointing into the original backing array. The GC cannot free the original array while any substring holds a reference. For long-lived substrings extracted from short-lived large strings, always make an explicit copy. IDEs may warn that `string([]byte(s))` is redundant, but it has a real effect: it forces a new allocation.

### 3.4 Understand Runes: len() Returns Bytes, Not Characters

**Impact: MEDIUM**

In Go, a `rune` is a Unicode code point (alias for `int32`). Strings are sequences of bytes, not characters. `len(s)` returns the number of bytes, not the number of characters. Multi-byte characters (like Chinese, emoji, or accented letters) can span 2-4 bytes each, causing off-by-one errors and corrupted output when you treat strings as byte arrays.

**Incorrect:**

```go
s := "hêllo"  // ê is a 2-byte UTF-8 character

fmt.Println(len(s))    // 6, not 5! (ê takes 2 bytes)
fmt.Println(s[1])      // 195 (raw byte), not 'ê'

// Iterating by index gives bytes, not runes
for i := 0; i < len(s); i++ {
    fmt.Printf("%c", s[i])  // Prints garbled output: hÃllo
}

// Truncating at byte index corrupts multi-byte characters
short := s[:3]  // Cuts ê in the middle — invalid UTF-8!
```

**Correct:**

```go
s := "hêllo"

// Get rune count (not byte count)
fmt.Println(utf8.RuneCountInString(s))  // 5 ✓

// Iterate over runes using range — i is byte position, r is the rune
for i, r := range s {
    fmt.Printf("position %d: %c\n", i, r)
}
// position 0: h
// position 1: ê  (starts at byte 1, takes 2 bytes)
// position 3: l  (starts at byte 3)
// ...

// Access the nth rune: convert to []rune first
runes := []rune(s)
fmt.Printf("%c\n", runes[1])  // ê ✓

// Safe substring by rune index
sub := string([]rune(s)[:3])  // "hêl" ✓
```

Why this matters: `len(s)` and indexing `s[i]` operate on bytes. A `range` loop over a string produces rune values with their starting byte positions. When working with text that may contain non-ASCII characters (names, addresses, international content), always use rune-based operations. Use `unicode/utf8` package for byte-level UTF-8 manipulation.

---

## References

1. [https://go.dev/doc/effective_go](https://go.dev/doc/effective_go)
2. [https://go.dev/wiki/CodeReviewComments](https://go.dev/wiki/CodeReviewComments)
3. [https://github.com/golang/go/wiki/CommonMistakes](https://github.com/golang/go/wiki/CommonMistakes)
