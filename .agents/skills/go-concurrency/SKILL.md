---
name: go-concurrency
description: "Use this skill when writing Go goroutines, channels, mutexes, sync primitives, context cancellation, or any concurrent Go code to avoid data races and goroutine leaks."
---

# Go Concurrency

Critical patterns for safe and correct concurrent Go code. Covers goroutine lifecycle, data races, channels vs mutexes, context propagation, sync package usage, and goroutine leak prevention.

## When to Use This Skill

- Writing goroutines and managing their lifecycle
- Choosing between channels and mutexes
- Using context for cancellation and propagation
- Working with sync.WaitGroup, sync.Mutex, sync.Cond
- Using errgroup for goroutine error handling
- Detecting and fixing data races

## How It Works

1. Identifies critical concurrency anti-patterns
2. Explains data races, deadlocks, and goroutine leaks
3. Shows safe patterns for common concurrent workloads
4. Covers both foundational concepts and practical implementation

## Categories Covered

1. **Concurrency Foundations** — goroutines vs parallelism, workload types, channel vs mutex, context
2. **Concurrency Practice** — goroutine leaks, data races, channels, sync primitives, errgroup

## Present Results to User

When identifying issues:
```
❌ Found: Goroutine with no exit condition (leak)
📍 Location: services/worker.go:78
⚠️ Impact: CRITICAL - Memory grows unbounded in production
✅ Fix: Use context cancellation or done channel

[Show incorrect code]
[Show correct code]
[Explain why the fix matters]
```

## Troubleshooting

**Channel vs mutex:**
- Use channels when transferring ownership of data between goroutines
- Use mutexes when protecting shared state accessed by multiple goroutines

**Race detector:**
- Always run `go test -race` — races that seem harmless in testing cause crashes in production
