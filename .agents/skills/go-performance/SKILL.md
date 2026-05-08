---
name: go-performance
description: "Use this skill when optimizing Go performance, writing benchmarks, using the standard library (HTTP, JSON, SQL, time), or ensuring production-quality testing."
---

# Go Performance & Quality

Best practices for Go standard library usage, effective testing, and performance optimization. Covers HTTP client/server pitfalls, JSON/SQL mistakes, table-driven tests, profiling, memory allocation, CPU caches, and GC tuning.

## When to Use This Skill

- Using net/http (client timeouts, connection pooling)
- Working with encoding/json, database/sql, time package
- Writing benchmarks and table-driven tests
- Profiling and optimizing Go applications
- Understanding escape analysis, inlining, false sharing
- Tuning the garbage collector

## How It Works

1. Identifies performance anti-patterns and stdlib misuse
2. Explains production impact (memory leaks, connection exhaustion, GC pressure)
3. Shows correct patterns with measurable improvements
4. Covers profiling-first optimization workflow

## Categories Covered

1. **Standard Library** — HTTP timeouts, JSON errors, SQL pitfalls, time.After leaks
2. **Testing** — table-driven tests, race detector, benchmarks, test execution modes
3. **Optimizations** — CPU caches, false sharing, stack vs heap, GC tuning, inlining

## Present Results to User

When identifying issues:
```
❌ Found: HTTP client with no timeout
📍 Location: client/api.go:23
⚠️ Impact: HIGH - Connections hang indefinitely in production
✅ Fix: Set Timeout on http.Client

[Show incorrect code]
[Show correct code]
[Explain why the fix matters]
```

## Troubleshooting

**Premature optimization warning:**
- Profile first with `go tool pprof` before applying optimizations
- These patterns prevent common footguns, not every pattern applies to every codebase

**Benchmark reliability:**
- Use `b.ResetTimer()` after setup code
- Use `b.ReportAllocs()` to track allocations
- Run with `-benchmem` flag
