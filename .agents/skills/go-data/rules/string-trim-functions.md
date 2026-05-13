## Know the Difference Between TrimRight and TrimSuffix

**Impact: LOW**

`strings.TrimRight` and `strings.TrimSuffix` look similar but work completely differently. Confusing them produces unexpected results that are hard to debug because both functions compile and run without errors.

**Incorrect:**

```go
// Trying to remove the suffix "xo" from "123oxo"
result := strings.TrimRight("123oxo", "xo")
fmt.Println(result)  // "123" — not "123o"!

// Developer expected TrimRight to behave like TrimSuffix
// but TrimRight removed ALL trailing 'x' and 'o' characters

// Similarly confused on the left side:
result2 := strings.TrimLeft("oxo123", "ox")
fmt.Println(result2)  // "123" (removes all leading 'o' and 'x')
// vs
result3 := strings.TrimPrefix("oxo123", "ox")
fmt.Println(result3)  // "o123" (removes only the prefix "ox" once)
```

**Correct:**

```go
// TrimRight/TrimLeft: removes all trailing/leading runes that are IN the set
strings.TrimRight("123oxo", "xo")    // "123"   — removes o, then x, then o
strings.TrimLeft("oxo123", "ox")     // "123"   — removes o, then x, then o
strings.Trim("oxo123oxo", "ox")      // "123"   — TrimLeft + TrimRight

// TrimSuffix/TrimPrefix: removes exactly the given substring (once)
strings.TrimSuffix("123oxo", "xo")  // "123o"  — removes "xo" suffix once
strings.TrimSuffix("123oxo", "*xo")  // "123oxo" — no match, unchanged
strings.TrimPrefix("oxo123", "ox")  // "o123"  — removes "ox" prefix once

// Rule of thumb:
// TrimRight/Left = remove characters from a SET
// TrimSuffix/Prefix = remove an exact SUBSTRING
```

Why this matters: `TrimRight("123oxo", "xo")` iterates backward and removes any rune that appears in the set `{'x', 'o'}` until it finds one that doesn't — giving `"123"`. `TrimSuffix("123oxo", "xo")` checks if the string ends with exactly `"xo"` and removes it once — giving `"123o"`. These are fundamentally different operations despite similar names.
