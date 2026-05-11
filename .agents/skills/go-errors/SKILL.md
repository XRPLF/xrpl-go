---
name: go-errors
description: "Use this skill when writing, reviewing, or debugging Go error handling — wrapping errors, sentinel errors, error type checking, panicking, or handling errors in defer."
---

# Go Error Management

Critical patterns for robust Go error handling. Covers sentinel errors, error wrapping with context, type/value checking, defer error patterns, and when to use panic vs return.

## When to Use This Skill

- Writing error handling for any Go function
- Reviewing code for ignored or swallowed errors
- Using `errors.Is` / `errors.As` correctly
- Deciding whether to wrap or not wrap an error
- Handling errors inside defer statements

## How It Works

1. Identifies error handling anti-patterns (CRITICAL impact)
2. Explains the production consequences of each mistake
3. Shows correct error propagation and wrapping patterns
4. Covers the full error handling lifecycle

## Categories Covered

1. **Error Management** — sentinel values, wrapping, type checking, panicking, defer errors, ignored returns

## Present Results to User

When identifying issues:
```
❌ Found: Ignoring error return value
📍 Location: handlers/user.go:45
⚠️ Impact: CRITICAL - Silent failures in production
✅ Fix: Check and handle the error explicitly

[Show incorrect code]
[Show correct code]
[Explain why the fix matters]
```

## Troubleshooting

**When is it safe to ignore errors:**
- `Close()` in defer after a successful write is sometimes acceptable — but log it
- Use explicit `_ = f.Close()` to show the ignore is intentional, never silent

**errors.Is vs errors.As:**
- `errors.Is` checks for a specific error value (sentinel)
- `errors.As` checks for a specific error type and unwraps the chain
