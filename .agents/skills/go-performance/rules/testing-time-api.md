## Make Time-Dependent Code Testable by Injecting or Passing Time

**Impact: MEDIUM**

Functions that call `time.Now()` directly are hard to test deterministically: the result changes with the real wall clock. Make time a dependency — either as a function field on a struct, or as a parameter passed by callers — so tests can inject a fixed, predictable time.

**Incorrect — calls time.Now() directly:**

```go
func (c *Cache) TrimOlderThan(since time.Duration) {
    t := time.Now().Add(-since)  // Flaky: depends on real clock; test may fail on slow machine
    for i := 0; i < len(c.events); i++ {
        if c.events[i].Timestamp.After(t) {
            c.events = c.events[i:]
            return
        }
    }
}

// Test is flaky because it relies on real time offsets
func TestCache_TrimOlderThan(t *testing.T) {
    events := []Event{
        {Timestamp: time.Now().Add(-20 * time.Millisecond)},
        {Timestamp: time.Now().Add(-10 * time.Millisecond)},
        {Timestamp: time.Now().Add(10 * time.Millisecond)},
    }
    cache := &Cache{}
    cache.Add(events)
    cache.TrimOlderThan(15 * time.Millisecond)  // Flaky on a loaded machine
    // ...
}
```

**Correct option 1 — inject time as a struct field:**

```go
type now func() time.Time

type Cache struct {
    mu     sync.RWMutex
    events []Event
    now    now  // Unexported; set via factory function
}

func NewCache() *Cache {
    return &Cache{
        events: make([]Event, 0),
        now:    time.Now,  // Production: use real clock
    }
}

func (c *Cache) TrimOlderThan(since time.Duration) {
    t := c.now().Add(-since)  // Uses injected clock
    // ...
}

// Test injects a fixed time — fully deterministic
func TestCache_TrimOlderThan(t *testing.T) {
    fixedTime := parseTime(t, "2020-01-01T12:00:00.06Z")
    events := []Event{
        {Timestamp: parseTime(t, "2020-01-01T12:00:00.04Z")},
        {Timestamp: parseTime(t, "2020-01-01T12:00:00.05Z")},
        {Timestamp: parseTime(t, "2020-01-01T12:00:00.06Z")},
    }
    cache := &Cache{now: func() time.Time { return fixedTime }}
    cache.Add(events)
    cache.TrimOlderThan(15 * time.Millisecond)
    // Result is deterministic regardless of machine speed
}
```

**Correct option 2 — pass current time as a parameter (simpler, but changes the API):**

```go
// Caller provides the current time — no internal clock dependency
func (c *Cache) TrimOlderThan(t time.Time) {
    for i := 0; i < len(c.events); i++ {
        if c.events[i].Timestamp.After(t) {
            c.events = c.events[i:]
            return
        }
    }
}

// Production usage:
cache.TrimOlderThan(time.Now().Add(-since))

// Test usage — completely deterministic:
cache.TrimOlderThan(parseTime(t, "2020-01-01T12:00:00.06Z").Add(-15 * time.Millisecond))
```

Why this matters: When a function embeds `time.Now()`, it couples business logic to the real wall clock. Tests must add sleeps or timing offsets that are inherently fragile on CI. Injecting time as a dependency (function field or parameter) lets tests pin the clock to a known value. Prefer option 2 (pass time explicitly) when it doesn't make the API awkward — it requires no stubs and no unexported fields. Use option 1 (struct field) when the function is called many times and passing time each time would be impractical.
