## Avoid Generic Utility Package Names

**Impact: MEDIUM**

Package names like `util`, `common`, `base`, `helpers`, or `misc` are anti-patterns in Go. They accumulate unrelated code over time and give callers no information about what the package provides. A package name should describe what it contains, not how it's categorized.

**Incorrect:**

```go
// util/util.go — What does this package do?
package util

func FormatDate(t time.Time) string { ... }
func HashPassword(pw string) string { ... }
func ValidateEmail(email string) bool { ... }
func RetryWithBackoff(fn func() error, attempts int) error { ... }

// Callers see: util.FormatDate, util.HashPassword — package name adds nothing
```

**Correct:**

```go
// timeutil/format.go — specific, descriptive
package timeutil

func Format(t time.Time) string { ... }

// auth/hash.go
package auth

func HashPassword(pw string) string { ... }

// validation/email.go
package validation

func ValidateEmail(email string) bool { ... }

// retry/retry.go
package retry

func WithBackoff(fn func() error, attempts int) error { ... }

// Callers see: timeutil.Format, auth.HashPassword — package name is informative
```

Why this matters: Package names in Go are part of the API — callers write `package.Function()`. When the package is named `util`, the call site is `util.HashPassword()` which tells the reader nothing. When named `auth`, the call site is `auth.HashPassword()` which is self-documenting.

If you find yourself creating a utility package, ask: "What do these functions have in common?" The answer is usually the right package name.
