## Watch for Subtle Bugs with Named Result Parameters

**Impact: MEDIUM**

Named result parameters are initialized to their zero values when a function begins. This means an early return that doesn't explicitly assign the named parameter will return the zero value — which can be a silent bug, especially for `error`.

**Incorrect:**

```go
// BUG: err is a named result parameter, initialized to nil
func (l loc) getCoordinates(ctx context.Context, address string) (
    lat, lng float32, err error) {

    isValid := l.validateAddress(address)
    if !isValid {
        return 0, 0, errors.New("invalid address")
    }

    if ctx.Err() != nil {
        // BUG: we return "err" but never assigned it!
        // err is still nil (its zero value)
        return 0, 0, err  // Always returns nil — context error is swallowed!
    }

    // Get coordinates...
    return lat, lng, nil
}
```

**Correct:**

```go
// Option 1: Assign before returning
func (l loc) getCoordinates(ctx context.Context, address string) (
    lat, lng float32, err error) {

    isValid := l.validateAddress(address)
    if !isValid {
        return 0, 0, errors.New("invalid address")
    }

    if err = ctx.Err(); err != nil {  // Assign ctx.Err() to err first
        return 0, 0, err
    }

    return lat, lng, nil
}

// Option 2: Avoid named result parameters when they add risk
func (l loc) getCoordinates(ctx context.Context, address string) (float32, float32, error) {
    if !l.validateAddress(address) {
        return 0, 0, errors.New("invalid address")
    }
    if err := ctx.Err(); err != nil {
        return 0, 0, err  // Can't forget to assign — no named param to confuse us
    }
    lat, lng := l.computeCoordinates(address)
    return lat, lng, nil
}
```

Why this matters: The bug compiles cleanly. The named `err` parameter is `nil` by default. The code `return 0, 0, err` looks correct but returns `nil` when the context was cancelled. This is especially dangerous in error handling paths where you intend to propagate an error but accidentally return nil. Named result parameters that shadow error handling are among the subtlest Go bugs.
