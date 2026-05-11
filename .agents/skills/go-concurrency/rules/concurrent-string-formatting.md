## String Formatting Can Cause Data Races or Deadlocks in Concurrent Code

**Impact: HIGH**

`fmt.Sprintf` and related functions call the `String()` method on values. If a `String()` method acquires a mutex that is already held in the calling goroutine, it causes a deadlock. If formatting traverses mutable shared data (like a `context.WithValue` chain), it can cause a data race.

**Incorrect (deadlock):**

```go
type Customer struct {
    mu  sync.RWMutex
    id  string
    age int
}

func (c *Customer) UpdateAge(age int) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if age < 0 {
        // BUG: %v calls c.String(), which tries to acquire c.mu.RLock()
        // But c.mu is already write-locked — DEADLOCK
        return fmt.Errorf("age should be positive for customer %v", c)
    }
    c.age = age
    return nil
}

func (c *Customer) String() string {
    c.mu.RLock()           // Tries to lock an already write-locked mutex
    defer c.mu.RUnlock()
    return fmt.Sprintf("id %s, age %d", c.id, c.age)
}
```

**Correct:**

```go
func (c *Customer) UpdateAge(age int) error {
    // Option 1: Validate BEFORE acquiring the lock
    if age < 0 {
        // c.mu is not locked here — String() can acquire RLock safely
        return fmt.Errorf("age should be positive for customer %v", c)
    }

    c.mu.Lock()
    defer c.mu.Unlock()
    c.age = age
    return nil
}

// Option 2: Format specific fields directly, don't call String()
func (c *Customer) UpdateAge(age int) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if age < 0 {
        return fmt.Errorf("age should be positive for customer id %s", c.id)
        // Accessing c.id directly — no String() call, no deadlock
    }
    c.age = age
    return nil
}
```

Why this matters: The `%v` and `%s` format directives call the `String()` method if implemented. If `String()` tries to acquire a lock that the formatting call's goroutine already holds, the goroutine deadlocks. Additionally, `fmt.Sprintf("%v", ctx)` traverses the context chain, which may access mutable values across goroutines — a potential data race. Be especially careful when formatting structs inside locked critical sections.
