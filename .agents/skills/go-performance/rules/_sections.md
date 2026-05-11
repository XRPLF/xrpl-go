# Sections for Go Performance & Quality

## 1. Standard Library (stdlib)

**Impact:** HIGH

**Description:** The Go standard library has subtle behaviors in time handling, HTTP client/server, JSON marshaling, and SQL operations. Misusing time.After, forgetting HTTP timeouts, and SQL connection pooling mistakes impact production. These patterns cover standard library best practices.

## 2. Testing (testing)

**Impact:** MEDIUM

**Description:** Effective testing in Go requires understanding table-driven tests, race detector usage, test execution modes, and benchmarking. Poor testing patterns lead to brittle tests, false confidence, and missed bugs. These patterns ensure robust test suites.

## 3. Optimizations (optimization)

**Impact:** HIGH

**Description:** Go performance optimization requires understanding CPU caches, memory allocation, inlining, escape analysis, and the garbage collector. Premature optimization wastes time, but knowing performance fundamentals prevents costly mistakes. These patterns cover optimization techniques.
