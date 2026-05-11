## Check for Integer Overflow in Critical Operations

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
