## Use Named Result Parameters Sparingly and Purposefully

**Impact: LOW**

Named result parameters initialize return values to their zero values and enable "naked" returns. They improve readability in specific cases but can harm it in others. Use them when they add genuine clarity, not as a habit.

**Incorrect:**

```go
// Unnecessary: single result, name adds no information
func StoreCustomer(customer Customer) (err error) {
    // "err" name doesn't help the reader
    return db.Save(customer)
}

// Unnecessary: obvious what the return means
func Add(a, b int) (result int) {
    result = a + b
    return  // Naked return — reader must remember what "result" is
}
```

**Correct:**

```go
// Good: disambiguates multiple returns of the same type
type locator interface {
    // Without names: unclear which float32 is lat and which is lng
    // getCoordinates(address string) (float32, float32, error)

    // With names: signature is self-documenting
    getCoordinates(address string) (lat, lng float32, err error)
}

// Good: convenience when named params simplify initialization
func ReadFull(r io.Reader, buf []byte) (n int, err error) {
    // n and err are pre-initialized to 0 and nil
    for len(buf) > 0 && err == nil {
        var nr int
        nr, err = r.Read(buf)
        n += nr
        buf = buf[nr:]
    }
    return  // Naked return is acceptable in short functions
}

// Rule: don't mix naked returns and explicit returns in the same function
```

Why this matters: Named result parameters are initialized to their zero values at function entry. This can cause subtle bugs — see mistake #44. Use them when: multiple returns have the same type and names disambiguate, or when pre-initialization provides a genuine benefit. Avoid naked returns in long functions — readers must scroll back to the signature to understand what's being returned. Keep naked returns to short functions only.
