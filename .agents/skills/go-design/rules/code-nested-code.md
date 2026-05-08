## Avoid Deeply Nested Code

**Impact: MEDIUM**

Deeply nested code is harder to read and reason about. When a function requires readers to track multiple levels of indentation to understand the flow, bugs hide and maintenance becomes painful. Go's convention is to keep the "happy path" aligned to the left with early returns (guard clauses).

**Incorrect:**

```go
func GetWeather(ctx context.Context, city string) (string, error) {
    if city != "" {
        resp, err := http.Get("https://api.weather.com/" + city)
        if err == nil {
            defer resp.Body.Close()
            body, err := io.ReadAll(resp.Body)
            if err == nil {
                return string(body), nil
            } else {
                return "", err
            }
        } else {
            return "", err
        }
    } else {
        return "", errors.New("city is required")
    }
}
```

**Correct:**

```go
func GetWeather(ctx context.Context, city string) (string, error) {
    if city == "" {
        return "", errors.New("city is required")
    }

    resp, err := http.Get("https://api.weather.com/" + city)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}
```

Why this matters: The guard clause style keeps the happy path at the leftmost indentation level. Each error is handled immediately and the function returns — no else branches needed. The reader can scan down the left margin to follow the main logic without tracking nested conditions.

Pattern: When you find yourself writing `if err == nil { ... }` blocks, invert them to `if err != nil { return ..., err }`. When you have `if condition { ... } else { ... }`, check if one branch is just an early return that lets you eliminate the else entirely.
