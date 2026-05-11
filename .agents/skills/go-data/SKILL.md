---
name: go-data
description: "Use this skill when working with Go slices, maps, integers, strings, range loops, or defer in loops to avoid common data type and control flow pitfalls."
---

# Go Data Types & Control Flow

Correct patterns for Go's built-in data types (slices, maps, integers), control structures (range loops, break, defer), and string handling (runes, UTF-8, conversions).

## When to Use This Skill

- Working with slices and maps (initialization, capacity, memory leaks)
- Iterating with range loops (variable capture, pointer reuse)
- Handling integers (overflow, octal literals, floating point)
- Manipulating strings (runes, trim, substring leaks)
- Using defer inside loops

## How It Works

1. Identifies data type and control flow anti-patterns
2. Explains subtle Go semantics (nil vs empty slice, range variable reuse)
3. Shows correct implementations with working examples
4. Categorized by impact level

## Categories Covered

1. **Data Types** — slices, maps, integers, floating point, nil/empty slice
2. **Control Structures** — range loops, defer in loops, break, map iteration
3. **Strings** — runes, UTF-8, trim functions, substring memory leaks

## Present Results to User

When identifying issues:
```
❌ Found: Range loop variable captured by closure
📍 Location: handlers/batch.go:34
⚠️ Impact: HIGH - All goroutines reference same variable
✅ Fix: Copy loop variable before passing to goroutine

[Show incorrect code]
[Show correct code]
[Explain why the fix matters]
```

## Troubleshooting

**Nil vs empty slice confusion:**
- Both have length 0, but nil slice marshals to `null` in JSON; empty slice marshals to `[]`
- Use `var s []T` for nil slice, `s := []T{}` or `make([]T, 0)` for empty
