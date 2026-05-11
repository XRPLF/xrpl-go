## Substring Operations Can Cause Memory Leaks

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
