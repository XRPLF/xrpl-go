## Prefer Nil Slices Over Empty Slices

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
