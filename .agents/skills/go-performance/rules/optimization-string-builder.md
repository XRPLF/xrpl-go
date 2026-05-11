## Use strings.Builder for Efficient String Concatenation

**Impact: MEDIUM**

Concatenating strings with `+` creates a new string each time because strings are immutable. For loops building strings, this causes O(n²) complexity and excessive allocations. Use `strings.Builder` instead.

**Incorrect:**

```go
func BuildQuery(fields []string) string {
    query := "SELECT "
    for i, field := range fields {
        if i > 0 {
            query += ", "  // New allocation every iteration!
        }
        query += field
    }
    query += " FROM table"
    return query
}
```

**Correct:**

```go
func BuildQuery(fields []string) string {
    var b strings.Builder
    b.WriteString("SELECT ")
    for i, field := range fields {
        if i > 0 {
            b.WriteString(", ")
        }
        b.WriteString(field)
    }
    b.WriteString(" FROM table")
    return b.String()
}
```

Why this matters: String concatenation with `+` copies both strings into a new allocation. With n fields, the incorrect version makes O(n²) allocations and copies. `strings.Builder` grows a single buffer, resulting in O(n) performance. For large strings or loops, this is significantly faster and reduces GC pressure.

When to use: Any loop concatenating strings, building responses, or constructing large strings. For 2-3 simple concatenations, `+` is fine and more readable.
