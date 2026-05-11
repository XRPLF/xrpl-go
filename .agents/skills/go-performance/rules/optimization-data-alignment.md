## Order Struct Fields by Size Descending to Reduce Padding and Memory Usage

**Impact: LOW**

Go aligns struct fields to multiples of their own size. Fields declared in a suboptimal order can cause the compiler to insert padding bytes, wasting memory and hurting cache performance. Sorting fields from largest to smallest eliminates unnecessary padding.

**Incorrect — field order causes padding:**

```go
// Foo uses 24 bytes due to alignment padding
type Foo struct {
    b1 byte    // 1 byte at 0x00
    // 7 bytes padding (compiler inserts to align i to multiple of 8)
    i  int64   // 8 bytes at 0x08
    b2 byte    // 1 byte at 0x10
    // 7 bytes padding (to make struct size a multiple of 8)
}
// Total: 1 + 7 (pad) + 8 + 1 + 7 (pad) = 24 bytes
// 14 of 24 bytes are padding!
```

**Correct — largest fields first:**

```go
// Foo uses only 16 bytes after reordering
type Foo struct {
    i  int64  // 8 bytes at 0x00
    b1 byte   // 1 byte at 0x08
    b2 byte   // 1 byte at 0x09
    // 6 bytes padding to reach next multiple of 8
}
// Total: 8 + 1 + 1 + 6 (pad) = 16 bytes
// 33% memory savings just from moving one field

// Rule of thumb: sort fields largest → smallest
// uint64/int64/float64/complex64/pointer: 8 bytes
// uint32/int32/float32: 4 bytes
// uint16/int16: 2 bytes
// uint8/int8/byte/bool: 1 byte
```

**Alignment guarantees in Go (64-bit architecture):**
| Type | Size | Alignment |
|------|------|-----------|
| byte, uint8, int8 | 1 byte | 1 byte |
| uint16, int16 | 2 bytes | 2 bytes |
| uint32, int32, float32 | 4 bytes | 4 bytes |
| uint64, int64, float64, complex64, pointer | 8 bytes | 8 bytes |
| complex128 | 16 bytes | 16 bytes |

**Checking alignment with unsafe:**

```go
import (
    "fmt"
    "unsafe"
)

type Foo struct {
    b1 byte
    i  int64
    b2 byte
}

fmt.Println(unsafe.Sizeof(Foo{}))   // 24 — includes padding
fmt.Println(unsafe.Alignof(Foo{}))  // 8 — alignment of the struct

type FooOpt struct {
    i  int64
    b1 byte
    b2 byte
}

fmt.Println(unsafe.Sizeof(FooOpt{}))  // 16 — no wasted padding
```

Why this matters: Padding bytes add up across millions of struct instances. A 24-byte struct vs. a 16-byte struct is a 33% memory increase; for 1 million instances, that's 8 MB of wasted memory. Beyond memory, larger structs mean fewer fit in a cache line, which increases cache misses when iterating over a slice. The compiler cannot reorder fields automatically (that would break binary compatibility). Use `betteralign` or `fieldalignment` from `golang.org/x/tools/go/analysis/passes/fieldalignment` to detect suboptimal struct layouts.
