---
name: go-design
description: "Use this skill when designing Go APIs, organizing packages, writing functions or methods, choosing receiver types, or structuring a Go project."
---

# Go Code Organization & Design

Idiomatic Go patterns for project structure, API design, interface usage, and function/method design. Helps avoid anti-patterns that lead to tight coupling, confusing APIs, and maintenance issues.

## When to Use This Skill

- Structuring a new Go project or package
- Designing interfaces, types, and exported APIs
- Choosing between value vs pointer receivers
- Using named return values or defer
- Avoiding variable shadowing and init() abuse

## How It Works

1. Identifies anti-patterns in code organization and API design
2. Explains why each pattern is problematic
3. Shows correct, idiomatic Go implementations
4. Categorized by impact level (CRITICAL, HIGH, MEDIUM, LOW)

## Categories Covered

1. **Code Organization** — project structure, interfaces, naming, linters
2. **Functions & Methods** — receivers, named returns, defer, nil receivers

## Present Results to User

When identifying issues:
```
❌ Found: Interface defined on producer side
📍 Location: pkg/store/store.go:12
⚠️ Impact: MEDIUM - Tight coupling, hard to test
✅ Fix: Define interface on the consumer side

[Show incorrect code]
[Show correct code]
[Explain why the fix matters]
```

## Troubleshooting

**Receiver type choice unclear:**
- Use pointer receiver when the method mutates state or the type is large
- Use value receiver for small, immutable types
