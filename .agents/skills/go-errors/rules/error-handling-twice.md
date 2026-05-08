## Handle an Error Only Once — Log OR Return, Not Both

**Impact: MEDIUM**

Handling an error twice — logging it and then returning it — causes the same error to appear multiple times in logs at different layers. Choose one: either handle the error (log it, recover from it) and return nil, or add context and return it up the call stack. Never do both.

**Incorrect:**

```go
func GetRoute(srcLat, srcLng, dstLat, dstLng float32) (Route, error) {
    err := validateCoordinates(srcLat, srcLng)
    if err != nil {
        log.Println("failed to validate source coordinates")  // Logged here...
        return Route{}, err                                    // ...AND returned
    }

    err = validateCoordinates(dstLat, dstLng)
    if err != nil {
        log.Println("failed to validate destination coordinates")
        return Route{}, err  // Double handling: log + return
    }

    return computeRoute(srcLat, srcLng, dstLat, dstLng)
}

// Caller also logs the error:
route, err := GetRoute(...)
if err != nil {
    log.Println("failed to get route:", err)  // Logged AGAIN — duplicate noise
}
// Result: same error logged 2+ times at different layers
```

**Correct:**

```go
// Option 1: Add context and return — let the top level log
func GetRoute(srcLat, srcLng, dstLat, dstLng float32) (Route, error) {
    if err := validateCoordinates(srcLat, srcLng); err != nil {
        return Route{}, fmt.Errorf("invalid source coordinates: %w", err)  // Context added
    }
    if err := validateCoordinates(dstLat, dstLng); err != nil {
        return Route{}, fmt.Errorf("invalid destination coordinates: %w", err)
    }
    return computeRoute(srcLat, srcLng, dstLat, dstLng)
}

// Top-level handler logs once with full context:
route, err := GetRoute(...)
if err != nil {
    log.Println("failed to get route:", err)
    // Error message: "failed to get route: invalid source coordinates: lat out of range"
}

// Option 2: Handle fully at the point of occurrence
func processRequest(r *http.Request) {
    if err := validateRequest(r); err != nil {
        log.Printf("invalid request: %v", err)
        // Handle the error: respond, recover, and return nil upward
        return  // Do NOT also return the error
    }
}
```

Why this matters: Each layer should either handle an error completely (log, recover, continue) or add context and propagate it upward — never both. Logging AND returning causes log flooding with duplicate entries, making debugging harder. Use `fmt.Errorf("context: %w", err)` to enrich errors as they propagate, then log them once at the appropriate boundary.
