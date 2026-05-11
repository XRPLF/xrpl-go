# Go Code Organization & Design

**Version 0.1.0**  
Agent Skills  
March 2026

> **Note:**  
> This document is mainly for agents and LLMs to follow when maintaining,  
> generating, or refactoring Go project structure, API design, and function/method patterns. Humans  
> may also find it useful, but guidance here is optimized for automation  
> and consistency by AI-assisted workflows.

---

## Abstract

Guidelines for Go code organization, API design, and function/method design. Covers idiomatic project structure, interface usage, naming conventions, receiver types, named returns, and defer evaluation.

---

## Table of Contents

1. [Code Organization](#1-code-organization) — **MEDIUM**
   - 1.1 [Avoid Deeply Nested Code](#11-avoid-deeply-nested-code)
   - 1.2 [Avoid Generic Utility Package Names](#12-avoid-generic-utility-package-names)
   - 1.3 [Avoid Interface Pollution](#13-avoid-interface-pollution)
   - 1.4 [Avoid Misusing init Functions](#14-avoid-misusing-init-functions)
   - 1.5 [Avoid Overusing the any Type](#15-avoid-overusing-the-any-type)
   - 1.6 [Avoid Package Name Collisions](#16-avoid-package-name-collisions)
   - 1.7 [Avoid Unintended Variable Shadowing](#17-avoid-unintended-variable-shadowing)
   - 1.8 [Be Careful with Type Embedding](#18-be-careful-with-type-embedding)
   - 1.9 [Define Interfaces on the Consumer Side](#19-define-interfaces-on-the-consumer-side)
   - 1.10 [Document All Exported Elements](#110-document-all-exported-elements)
   - 1.11 [Follow Standard Go Project Layout](#111-follow-standard-go-project-layout)
   - 1.12 [Return Structs, Accept Interfaces](#112-return-structs-accept-interfaces)
   - 1.13 [Use Generics Appropriately](#113-use-generics-appropriately)
   - 1.14 [Use Idiomatic Getter and Setter Names](#114-use-idiomatic-getter-and-setter-names)
   - 1.15 [Use Linters to Catch Common Mistakes](#115-use-linters-to-catch-common-mistakes)
   - 1.16 [Use the Functional Options Pattern for Configuration](#116-use-the-functional-options-pattern-for-configuration)
2. [Functions & Methods](#2-functions--methods) — **MEDIUM**
   - 2.1 [Accept io.Reader Instead of Filename as Function Input](#21-accept-ioreader-instead-of-filename-as-function-input)
   - 2.2 [Defer Arguments Are Evaluated Immediately, Not When Called](#22-defer-arguments-are-evaluated-immediately-not-when-called)
   - 2.3 [Returning a Nil Pointer as an Interface Is Not Nil](#23-returning-a-nil-pointer-as-an-interface-is-not-nil)
   - 2.4 [Use Consistent, Short Receiver Names](#24-use-consistent-short-receiver-names)
   - 2.5 [Use Named Result Parameters Sparingly and Purposefully](#25-use-named-result-parameters-sparingly-and-purposefully)
   - 2.6 [Watch for Subtle Bugs with Named Result Parameters](#26-watch-for-subtle-bugs-with-named-result-parameters)

---

## 1. Code Organization

**Impact: MEDIUM**

Well-organized Go code follows idiomatic patterns for project structure, naming conventions, and API design. Poor organization leads to maintenance nightmares, tight coupling, and confused teammates. These patterns cover variable shadowing, init functions, interfaces, getters/setters, and project layout.

### 1.1 Avoid Deeply Nested Code

**Impact: MEDIUM**

Deeply nested code is harder to read and reason about. When a function requires readers to track multiple levels of indentation to understand the flow, bugs hide and maintenance becomes painful. Go's convention is to keep the "happy path" aligned to the left with early returns (guard clauses).

**Incorrect:**

```go
func GetWeather(ctx context.Context, city string) (string, error) {
    if city != "" {
        resp, err := http.Get("https://api.weather.com/" + city)
        if err == nil {
            defer resp.Body.Close()
            body, err := io.ReadAll(resp.Body)
            if err == nil {
                return string(body), nil
            } else {
                return "", err
            }
        } else {
            return "", err
        }
    } else {
        return "", errors.New("city is required")
    }
}
```

**Correct:**

```go
func GetWeather(ctx context.Context, city string) (string, error) {
    if city == "" {
        return "", errors.New("city is required")
    }

    resp, err := http.Get("https://api.weather.com/" + city)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}
```

Why this matters: The guard clause style keeps the happy path at the leftmost indentation level. Each error is handled immediately and the function returns — no else branches needed. The reader can scan down the left margin to follow the main logic without tracking nested conditions.

Pattern: When you find yourself writing `if err == nil { ... }` blocks, invert them to `if err != nil { return ..., err }`. When you have `if condition { ... } else { ... }`, check if one branch is just an early return that lets you eliminate the else entirely.

### 1.2 Avoid Generic Utility Package Names

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

### 1.3 Avoid Interface Pollution

**Impact: MEDIUM**

Creating interfaces speculatively — before you have multiple concrete implementations or a clear need for abstraction — pollutes your codebase with unnecessary indirection. Interfaces should be discovered, not invented. The common Go proverb: "Don't design with interfaces, discover them."

**Incorrect:**

```go
// Created "just in case" — only one implementation exists
type CustomerService interface {
    GetCustomer(id int) (*Customer, error)
    CreateCustomer(name, email string) (*Customer, error)
    DeleteCustomer(id int) error
}

type customerService struct {
    db *sql.DB
}

func NewCustomerService(db *sql.DB) CustomerService {
    return &customerService{db: db}
}
```

**Correct:**

```go
// Return the concrete type — callers can define the interface they need
type CustomerService struct {
    db *sql.DB
}

func NewCustomerService(db *sql.DB) *CustomerService {
    return &CustomerService{db: db}
}

func (s *CustomerService) GetCustomer(id int) (*Customer, error) { ... }
func (s *CustomerService) CreateCustomer(name, email string) (*Customer, error) { ... }
func (s *CustomerService) DeleteCustomer(id int) error { ... }

// Consumer defines what it needs:
type customerReader interface {
    GetCustomer(id int) (*Customer, error)
}
```

Why this matters: Unnecessary interfaces add cognitive overhead — readers must find the concrete type to understand what's happening. They also make refactoring harder by creating artificial contracts. Interfaces are most valuable when: (1) you have multiple concrete implementations, (2) you need to decouple packages, or (3) you need to restrict behavior for testing.

Reference: [https://go.dev/doc/faq#overloading](https://go.dev/doc/faq#overloading)

### 1.4 Avoid Misusing init Functions

**Impact: MEDIUM**

`init` functions run automatically at package initialization, before `main`. They look convenient for setup work, but they have significant drawbacks: they cannot return errors, they force the use of global state, they execute even in tests (causing side effects), and they make dependency injection impossible.

**Incorrect:**

```go
var db *sql.DB

func init() {
    // Cannot return an error — if this fails, only option is panic
    var err error
    db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        panic(err)  // Crashes the entire program on startup
    }
}

func GetUser(id int) (*User, error) {
    // Tests that import this package will trigger DB connection
    return queryUser(db, id)
}
```

**Correct:**

```go
func NewDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("opening db: %w", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("pinging db: %w", err)
    }
    return db, nil
}

func main() {
    db, err := NewDB(os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    server := NewServer(db)
    server.Run()
}
```

Why this matters: Constructor functions return errors, making failures explicit and handleable. They don't run until you call them, so tests that don't need a database won't trigger a connection attempt. Dependencies are passed explicitly, making the code easier to test and understand.

Use `init` only for truly static setup with no error cases: registering codecs, setting up global flag defaults, or initializing lookup tables that can't fail.

### 1.5 Avoid Overusing the any Type

**Impact: MEDIUM**

The `any` type (alias for `interface{}`) can hold any value but provides no type safety. Overusing it means the compiler can't help you catch type errors, you need runtime type assertions everywhere, and the code becomes harder to understand and maintain.

**Incorrect:**

```go
// Store loses all type information
type Cache struct {
    mu    sync.Mutex
    items map[string]any
}

func (c *Cache) Set(key string, value any) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = value
}

func (c *Cache) Get(key string) any {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.items[key]
}

// Caller must type-assert, can panic at runtime
user := cache.Get("user").(User)  // panics if wrong type
```

**Correct:**

```go
// Option 1: Use generics for type-safe containers (Go 1.18+)
type Cache[T any] struct {
    mu    sync.Mutex
    items map[string]T
}

func (c *Cache[T]) Set(key string, value T) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = value
}

func (c *Cache[T]) Get(key string) (T, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
    v, ok := c.items[key]
    return v, ok
}

// Option 2: Use specific types when the domain is clear
type UserCache struct {
    mu    sync.Mutex
    items map[string]User
}
```

Why this matters: `any` parameters in function signatures communicate nothing about what the function expects. Callers must read documentation or source code to understand valid inputs. Type assertions at runtime can panic. Generics provide the flexibility of `any` with compile-time safety.

Use `any` only at true boundaries: JSON unmarshaling into unknown structures, reflection-based utilities, or framework code that genuinely handles arbitrary types.

### 1.6 Avoid Package Name Collisions

**Impact: LOW**

When a local variable or function parameter has the same name as an imported package, the variable shadows the package name within that scope. This leads to confusing code where readers must determine whether a name refers to a package or a variable.

**Incorrect:**

```go
import "net/http"

func HandleRequest(http *http.Request) {  // "http" variable shadows http package!
    // Inside this function, "http" refers to the parameter, not the package
    // This is confusing and error-prone
    client := &http.Client{}  // Compile error: http.Client doesn't exist on *http.Request
}

// Also problematic: local variable shadows package
func Process() {
    context := context.Background()  // shadows "context" package
    _ = context
}
```

**Correct:**

```go
import "net/http"

// Use descriptive parameter names that don't conflict with packages
func HandleRequest(req *http.Request) {
    client := &http.Client{}  // http package is accessible
    resp, err := client.Do(req)
    ...
}

// Or use import aliases when you must disambiguate
import (
    gocontext "context"
    "myapp/internal/context"  // custom context package
)

func Process() {
    ctx := gocontext.Background()
    appCtx := context.New()
    ...
}
```

Why this matters: Shadowed package names cause compilation errors when you try to use the package within the shadowing scope, and they confuse readers who must track whether a name is a package or variable. Common collisions: `context`, `http`, `io`, `os`, `log`, `sync`.

Prevention: Choose parameter names that describe the domain (`req` for requests, `w` for writers, `r` for readers) rather than echoing the type name.

### 1.7 Avoid Unintended Variable Shadowing

**Impact: MEDIUM**

Variable shadowing occurs when you declare a variable in an inner scope with the same name as an outer scope variable. This is legal in Go but often leads to confusing bugs where you think you're modifying one variable but are actually working with a different one.

**Incorrect:**

```go
func LoadConfig() (*Config, error) {
    config := &Config{Timeout: 30}

    if file, err := os.Open("config.json"); err == nil {
        defer file.Close()
        config, err := json.Unmarshal(data, &config)  // Shadows outer 'config'!
        if err != nil {
            return nil, err
        }
        // This config only exists in this block
    }

    return config, nil  // Returns partially initialized config!
}
```

**Correct:**

```go
func LoadConfig() (*Config, error) {
    config := &Config{Timeout: 30}

    file, err := os.Open("config.json")
    if err == nil {
        defer file.Close()
        err = json.Unmarshal(data, config)  // Uses outer 'config'
        if err != nil {
            return nil, err
        }
    }

    return config, nil
}
```

Why this matters: The incorrect version creates a new `config` variable inside the if block that shadows the outer one. Changes to this shadowed variable don't affect the outer `config`, so the function returns the partially initialized default config instead of the loaded one.

Prevention: Use `go vet` which detects some shadowing cases. Consider shorter scopes and unique variable names. The `:=` operator is convenient but dangerous in nested scopes—sometimes explicit `var` declarations make shadowing more obvious.

### 1.8 Be Careful with Type Embedding

**Impact: MEDIUM**

Type embedding promotes all methods of the embedded type to the outer struct. This is powerful but dangerous when you embed types that expose behavior you don't want to be part of your API. Accidentally exposing internal concurrency primitives is the most common mistake.

**Incorrect:**

```go
type InMemoryCache struct {
    sync.Mutex  // Exposes Lock() and Unlock() as public methods!
    items map[string]string
}

func NewInMemoryCache() *InMemoryCache {
    return &InMemoryCache{items: make(map[string]string)}
}

// Callers can now call cache.Lock() directly — this breaks encapsulation
// and allows callers to deadlock your cache
cache := NewInMemoryCache()
cache.Lock()   // Exposed through embedding — this is wrong
cache.Set("key", "value")
```

**Correct:**

```go
// Option 1: Use a named field (composition, not embedding)
type InMemoryCache struct {
    mu    sync.Mutex  // Unexported field — not promoted
    items map[string]string
}

func (c *InMemoryCache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = value
}

// Option 2: Embed only when you want to expose the embedded type's API
type ReadOnlyDB struct {
    *sql.DB  // Intentionally exposing DB's query methods
}

// Override to block writes
func (r *ReadOnlyDB) Exec(query string, args ...any) (sql.Result, error) {
    return nil, errors.New("read-only database")
}
```

Why this matters: Embedding `sync.Mutex` exposes `Lock()` and `Unlock()` as public methods on your struct. This lets callers hold the lock externally and cause deadlocks. The rule: embed when you want to expose, use a named (unexported) field when you want to encapsulate. Ask: "Do I want callers to use the embedded type's methods directly?"

### 1.9 Define Interfaces on the Consumer Side

**Impact: MEDIUM**

In Go, interfaces should live with the package that uses them, not the package that implements them. When producers define interfaces, they force consumers to depend on an abstraction they didn't ask for. When consumers define interfaces, they express exactly the behavior they need — nothing more.

**Incorrect:**

```go
// producer/store.go — producer defines the interface
package store

// Store is defined by the producer
type Store interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
    UpdateUser(user *User) error
    DeleteUser(id int) error
}

type PostgresStore struct { db *sql.DB }
func (s *PostgresStore) GetUser(id int) (*User, error) { ... }
// ... other methods

// consumer/service.go — forced to use producer's interface
import "store"
func NewService(s store.Store) *Service { ... }
```

**Correct:**

```go
// store/store.go — producer returns a concrete type
package store

type PostgresStore struct { db *sql.DB }
func NewPostgresStore(db *sql.DB) *PostgresStore { ... }
func (s *PostgresStore) GetUser(id int) (*User, error) { ... }
func (s *PostgresStore) CreateUser(user *User) error { ... }

// service/service.go — consumer defines only what it needs
package service

// Defined by the consumer, containing only required methods
type userReader interface {
    GetUser(id int) (*User, error)
}

func NewService(r userReader) *Service { ... }
// Works with *store.PostgresStore because it satisfies userReader
```

Why this matters: Consumer-side interfaces keep them small and focused (interface segregation), reduce unnecessary coupling between packages, and make testing trivial — just implement the small interface in tests. Producer-side interfaces tend to grow large and create import cycles.

Reference: [https://go.dev/doc/effective_go#interfaces](https://go.dev/doc/effective_go#interfaces)

### 1.10 Document All Exported Elements

**Impact: MEDIUM**

Go's documentation system generates docs directly from comments. Every exported type, function, method, constant, and variable should have a doc comment. These comments appear in `go doc` output, pkg.go.dev, and editor hover documentation. Undocumented exports force users to read source code.

**Incorrect:**

```go
package auth

// Missing package comment

type Token struct {  // No doc comment
    Value     string
    ExpiresAt time.Time
}

func Validate(token string) error {  // No doc comment
    ...
}

const MaxTokenAge = 24 * time.Hour  // No doc comment
```

**Correct:**

```go
// Package auth provides JWT-based authentication utilities.
package auth

// Token represents an authentication token with its expiry.
type Token struct {
    Value     string
    ExpiresAt time.Time
}

// Validate checks that the token string is properly formatted, signed,
// and not expired. It returns an error describing the validation failure
// if the token is invalid.
func Validate(token string) error {
    ...
}

// MaxTokenAge is the maximum lifetime of an authentication token.
const MaxTokenAge = 24 * time.Hour
```

Why this matters: Good documentation reduces the time teammates spend reading source code to understand how to use an API. Go doc comments follow a convention: start the comment with the name of the element being documented (`// Token represents...`, `// Validate checks...`). This makes `go doc` output read as complete sentences.

Rules:

- Package doc goes above the `package` declaration

- Start each comment with the element name

- Document edge cases, error conditions, and non-obvious behavior

- Use `go doc` or `godoc` to check how your comments render

### 1.11 Follow Standard Go Project Layout

**Impact: MEDIUM**

Inconsistent project structure makes it harder for new contributors to navigate the codebase and for tools to find code. The Go community has converged on a standard layout that separates concerns and signals intent through directory names.

**Incorrect:**

```go
myapp/
├── main.go          # Everything in root
├── server.go
├── db.go
├── utils.go         # Catch-all utility file
├── helpers.go       # Another catch-all
└── models.go
```

**Correct:**

```go
myapp/
├── cmd/
│   └── myapp/
│       └── main.go          # Entry point(s)
├── internal/
│   ├── server/
│   │   └── server.go        # Not importable by external packages
│   ├── store/
│   │   └── postgres.go
│   └── domain/
│       └── user.go
├── pkg/
│   └── client/
│       └── client.go        # Stable public API, importable externally
├── api/
│   └── openapi.yaml         # API specifications
├── scripts/
│   └── migrate.sh
└── go.mod
```

Why this matters: The `internal/` directory enforces that packages inside cannot be imported by code outside the module — a Go compiler guarantee that prevents leaking implementation details. The `cmd/` convention supports multiple binaries in one repo. The `pkg/` directory signals stable, intentionally public packages.

Key rules:

- Use `internal/` for implementation packages not intended for external use

- Use `cmd/{binary-name}/main.go` for entry points

- Keep `main.go` thin — put logic in `internal/`

- Avoid flat structures with dozens of files in the root

Reference: [https://github.com/golang-standards/project-layout](https://github.com/golang-standards/project-layout)

### 1.12 Return Structs, Accept Interfaces

**Impact: MEDIUM**

Functions should return concrete types (structs or pointers) and accept interfaces as parameters. Returning interfaces from constructors forces callers to depend on an abstraction and prevents them from accessing concrete methods. The principle: "be conservative in what you return, be liberal in what you accept."

**Incorrect:**

```go
// Returning an interface from a constructor
type Logger interface {
    Info(msg string)
    Error(msg string, err error)
}

// Forces callers to use Logger interface, hiding concrete type
func NewLogger(level string) Logger {
    return &zapLogger{level: level}
}

// Callers can't access zapLogger-specific methods without type assertion
```

**Correct:**

```go
type Logger interface {
    Info(msg string)
    Error(msg string, err error)
}

type ZapLogger struct {
    level  string
    sugar  *zap.SugaredLogger
}

// Return concrete type — callers can use it as Logger interface or access specific methods
func NewZapLogger(level string) (*ZapLogger, error) {
    // ...
    return &ZapLogger{level: level, sugar: sugar}, nil
}

func (l *ZapLogger) Info(msg string)              { l.sugar.Info(msg) }
func (l *ZapLogger) Error(msg string, err error)  { l.sugar.Errorw(msg, "error", err) }
func (l *ZapLogger) Sync() error                  { return l.sugar.Sync() } // concrete method
```

Why this matters: When you return an interface, callers lose access to concrete methods without unsafe type assertions. They also become tightly coupled to your abstraction choice. Returning concrete types gives callers maximum flexibility — they can use the value as an interface, access concrete methods, or embed it. If a caller needs an interface, they define one themselves.

Exception: Returning `error` (an interface) is idiomatic and correct — it's a well-established Go convention.

### 1.13 Use Generics Appropriately

**Impact: MEDIUM**

Go 1.18 introduced generics. They are most valuable for data structures and functions that operate on slices, maps, or channels of any type. Avoid using generics when interfaces or simple function overloading are clearer — generics add complexity when not needed.

**Incorrect:**

```go
// Don't use generics when an interface suffices
type Stringer[T any] interface {
    String() T
}

// Don't over-engineer simple operations
func Map[T, U any](slice []T, fn func(T) U) []U {
    // Fine for a utility library, but don't reach for generics first
}

// Using generics to avoid a trivial interface
func PrintAny[T fmt.Stringer](v T) {
    fmt.Println(v.String())  // Just use fmt.Stringer directly
}
```

**Correct:**

```go
// Good: Generic data structure avoids code duplication
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    var zero T
    if len(s.items) == 0 {
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

// Good: Utility functions on slices of any comparable type
func Contains[T comparable](slice []T, item T) bool {
    for _, v := range slice {
        if v == item {
            return true
        }
    }
    return false
}
```

Why this matters: Generics add syntactic complexity. Use them when: (1) you're building a data structure that holds arbitrary types, (2) writing utility functions that work on slices/maps/channels of any type, or (3) factoring out duplicated code that differs only in type. Don't use them to avoid a simple interface, and don't force generic parameters that constrain callers unnecessarily.

Reference: [https://go.dev/doc/faq#generics](https://go.dev/doc/faq#generics)

### 1.14 Use Idiomatic Getter and Setter Names

**Impact: LOW**

Go does not follow Java-style getter/setter conventions. Prefixing getters with `Get` is not idiomatic and adds noise without benefit. Follow Go's naming conventions to write APIs that feel natural to other Go developers.

**Incorrect:**

```go
type Account struct {
    balance float64
}

// Java-style: Get prefix is redundant
func (a *Account) GetBalance() float64 {
    return a.balance
}

// SetBalance is fine, but GetBalance is not idiomatic
func (a *Account) SetBalance(amount float64) {
    a.balance = amount
}
```

**Correct:**

```go
type Account struct {
    balance float64
}

// Getter: named after the field, no Get prefix
func (a *Account) Balance() float64 {
    return a.balance
}

// Setter: SetX is acceptable
func (a *Account) SetBalance(amount float64) {
    if amount < 0 {
        panic("balance cannot be negative")
    }
    a.balance = amount
}
```

Why this matters: Go's standard library sets the precedent — `http.Request` has `Cookie()` not `GetCookie()`. Code that follows these conventions integrates naturally with Go tooling and feels idiomatic to experienced Go developers. Deviating from convention creates friction.

More importantly, don't add getters/setters by default. Export the field directly if no validation or encapsulation logic is needed. Only add getters/setters when they provide value: input validation, computed values, future flexibility, or satisfying an interface.

### 1.15 Use Linters to Catch Common Mistakes

**Impact: MEDIUM**

Go's compiler catches many errors but not all subtle bugs. Linters perform static analysis to catch issues the compiler misses: shadowed variables, unchecked errors, complexity hotspots, duplicate code, and security vulnerabilities. Running linters in CI prevents these issues from reaching production.

**Incorrect:**

```go
// These issues compile fine but linters catch them:

func ProcessFile(path string) {
    f, _ := os.Open(path)     // errcheck: error return ignored
    defer f.Close()

    data := make([]byte, 1024)
    n, _ := f.Read(data)      // errcheck: error return ignored

    result := data[:n]
    _ = result
}

func IsPrime(n int) bool {
    // gocyclo: function complexity too high (many nested conditions)
    ...
}
```

**Correct:**

```go
// .golangci.yml — configure linters at the repo root
// Run: golangci-lint run ./...

func ProcessFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("opening %s: %w", path, err)
    }
    defer f.Close()

    data := make([]byte, 1024)
    n, err := f.Read(data)
    if err != nil && err != io.EOF {
        return fmt.Errorf("reading %s: %w", path, err)
    }

    return process(data[:n])
}
```

Why this matters: Linters act as automated code reviewers that run consistently. Key linters to enable:

- `errcheck` — finds ignored error returns

- `gosimple` — suggests simpler code patterns

- `staticcheck` — advanced static analysis

- `govet` — Go's own vet tool (reports suspicious constructs)

- `gocyclo` — flags high cyclomatic complexity

- `gosec` — security-focused analysis

Add `.golangci.yml` to your project root and run `golangci-lint run` in CI. Fix issues before they accumulate — a clean lint baseline is easy to maintain.

Reference: [https://golangci-lint.run/](https://golangci-lint.run/)

### 1.16 Use the Functional Options Pattern for Configuration

**Impact: MEDIUM**

When a constructor has many optional parameters, avoid long parameter lists or config structs that must be zeroed out. The functional options pattern provides a clean, extensible API that's easy to read, optional by default, and backward-compatible when new options are added.

**Incorrect:**

```go
// Growing parameter list — breaks callers when adding new options
func NewServer(host string, port int, timeout time.Duration, maxConns int, tls bool) *Server

// Config struct — callers must know zero values and field names
type ServerConfig struct {
    Host     string
    Port     int
    Timeout  time.Duration
    MaxConns int
    TLS      bool
}
func NewServer(cfg ServerConfig) *Server
```

**Correct:**

```go
type options struct {
    port     int
    timeout  time.Duration
    maxConns int
    tls      bool
}

type Option func(*options) error

func WithPort(port int) Option {
    return func(o *options) error {
        if port < 1 || port > 65535 {
            return fmt.Errorf("invalid port: %d", port)
        }
        o.port = port
        return nil
    }
}

func WithTimeout(d time.Duration) Option {
    return func(o *options) error {
        o.timeout = d
        return nil
    }
}

func NewServer(host string, opts ...Option) (*Server, error) {
    o := &options{
        port:    8080,          // sensible defaults
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        if err := opt(o); err != nil {
            return nil, err
        }
    }
    return &Server{host: host, opts: o}, nil
}

// Usage: clean, self-documenting, options are optional
srv, err := NewServer("localhost",
    WithPort(9090),
    WithTimeout(60*time.Second),
)
```

Why this matters: Adding a new option never breaks existing callers. Each option is self-documenting. Invalid configurations can be caught at construction time with meaningful errors. The pattern scales from 2 to 20+ options without API churn.

---

## 2. Functions & Methods

**Impact: MEDIUM**

Function and method design in Go involves choosing receiver types, understanding named return parameters, and handling defer evaluation. Poor choices lead to confusing APIs and subtle bugs. These patterns ensure clean function design.

### 2.1 Accept io.Reader Instead of Filename as Function Input

**Impact: MEDIUM**

Functions that accept a filename to read from are harder to test and less reusable than functions that accept an `io.Reader`. A filename forces callers to create actual files, while `io.Reader` accepts files, HTTP bodies, strings, byte buffers, or any other data source.

**Incorrect:**

```go
// Tightly coupled to the filesystem — hard to test, hard to reuse
func countEmptyLines(filename string) (int, error) {
    file, err := os.Open(filename)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var count int
    for scanner.Scan() {
        if scanner.Text() == "" {
            count++
        }
    }
    return count, scanner.Err()
}

// Testing requires creating actual files:
// os.WriteFile("test.txt", []byte("foo\n\nbar"), 0o644)
// count, err := countEmptyLines("test.txt")
// os.Remove("test.txt")
```

**Correct:**

```go
// Accepts any data source — testable and reusable
func countEmptyLines(r io.Reader) (int, error) {
    scanner := bufio.NewScanner(r)
    var count int
    for scanner.Scan() {
        if scanner.Text() == "" {
            count++
        }
    }
    return count, scanner.Err()
}

// Testing is trivial — no files needed:
func TestCountEmptyLines(t *testing.T) {
    input := strings.NewReader("foo\n\nbar\n\nbaz")
    count, err := countEmptyLines(input)
    // ...
}

// Works with files:
file, _ := os.Open("data.txt")
defer file.Close()
count, _ := countEmptyLines(file)

// Works with HTTP request body:
count, _ := countEmptyLines(r.Body)
```

Why this matters: `*os.File`, `http.Request.Body`, `strings.NewReader`, and `bytes.NewReader` all implement `io.Reader`. A function accepting `io.Reader` works with all of them transparently. Test cases don't need filesystem setup/teardown. The function is reusable across different data sources without duplication.

Exception: `os.Open` and similar filesystem-specific functions obviously need filenames. But functions that just *read* data should prefer `io.Reader`.

### 2.2 Defer Arguments Are Evaluated Immediately, Not When Called

**Impact: MEDIUM**

`defer` schedules a function call for when the surrounding function returns, but it evaluates the function's arguments **immediately** when the `defer` statement is reached. If you rely on a variable that changes later, the deferred call will use the value at the time of the `defer` — not the final value.

**Incorrect:**

```go
func notify(status string)          { /* send notification */ }
func incrementCounter(status string) { /* update metric */ }

func process() error {
    var status string
    defer notify(status)          // BUG: status="" captured NOW, not later
    defer incrementCounter(status) // BUG: status="" captured NOW

    if err := foo(); err != nil {
        status = "error_foo"
        return err
    }
    if err := bar(); err != nil {
        status = "error_bar"
        return err
    }
    status = "success"
    return nil
    // notify and incrementCounter always receive "" regardless of path
}
```

**Correct:**

```go
// Option 1: Pass a pointer so the deferred function reads the final value
func process() error {
    var status string
    defer notify(&status)           // Pointer to status — reads value at call time
    defer incrementCounter(&status)

    if err := foo(); err != nil {
        status = "error_foo"
        return err
    }
    if err := bar(); err != nil {
        status = "error_bar"
        return err
    }
    status = "success"
    return nil
    // notify and incrementCounter receive the final status value ✓
}

// Option 2: Use a closure — captures variables by reference, not by value
func process() error {
    var status string
    defer func() {
        notify(status)           // Reads status when the closure executes
        incrementCounter(status)
    }()

    // ... same logic
    status = "success"
    return nil
}
```

Why this matters: `defer notify(status)` is equivalent to saving the call as `notify("")` at that exact moment. This is a frequent source of bugs in logging, metrics, and cleanup code. With a **closure**, variables referenced from the outer scope are read when the closure runs (on function return), not when `defer` is called. With a **pointer**, the deferred function dereferences it at call time. Both solutions correctly capture the final state.

### 2.3 Returning a Nil Pointer as an Interface Is Not Nil

**Impact: HIGH**

In Go, an interface holds two things: a type and a value. An interface is `nil` only if both are nil. When you return a nil pointer of a concrete type as an interface, the interface is NOT nil — it has a type but a nil value. This causes `if err != nil` checks to pass even when there's "no error."

**Incorrect:**

```go
type MultiError struct {
    errs []string
}

func (m *MultiError) Error() string {
    return strings.Join(m.errs, "; ")
}

func (c Customer) Validate() error {
    var m *MultiError  // nil pointer to MultiError

    if c.Age < 0 {
        m = &MultiError{}
        m.errs = append(m.errs, "age is negative")
    }

    // BUG: returning a nil *MultiError as error interface
    return m  // Interface has type=*MultiError, value=nil — NOT nil!
}

// Caller is surprised:
err := customer.Validate()
if err != nil {  // This is TRUE even when there are no errors!
    log.Fatal(err)  // Executed even for valid customers
}
```

**Correct:**

```go
func (c Customer) Validate() error {
    var m *MultiError

    if c.Age < 0 {
        m = &MultiError{}
        m.errs = append(m.errs, "age is negative")
    }

    // Option 1: Explicit nil return when no errors
    if m != nil {
        return m
    }
    return nil  // Returns a nil interface, not a nil pointer wrapped in interface

    // Option 2: Return error interface directly
    // var errs []string
    // if ... { errs = append(errs, "...") }
    // if len(errs) > 0 { return fmt.Errorf("%s", strings.Join(errs, "; ")) }
    // return nil
}
```

Why this matters: A nil `*MultiError` is not a nil `error`. The interface value `error{type: *MultiError, value: nil}` is non-nil. Any function returning a concrete pointer type as an interface must explicitly check if the pointer is nil and return `nil` (bare) in that case. This is one of the most counterintuitive behaviors in Go.

### 2.4 Use Consistent, Short Receiver Names

**Impact: MEDIUM**

Receiver names should be short (1-2 characters), consistent across all methods of a type, and reflect the type name. Avoid `this` or `self` which are not idiomatic in Go.

**Incorrect:**

```go
type User struct {
    Name string
}

func (this *User) GetName() string {  // "this" is not idiomatic
    return this.Name
}

func (user *User) SetName(name string) {  // Inconsistent with GetName
    user.Name = name
}

func (u *User) Validate() error {  // Another inconsistent name
    if u.Name == "" {
        return errors.New("name required")
    }
    return nil
}
```

**Correct:**

```go
type User struct {
    Name string
}

func (u *User) GetName() string {
    return u.Name
}

func (u *User) SetName(name string) {
    u.Name = name
}

func (u *User) Validate() error {
    if u.Name == "" {
        return errors.New("name required")
    }
    return nil
}
```

Why this matters: Consistent receiver names make code easier to scan and understand. Go convention is to use the first letter(s) of the type name: `User` → `u`, `HTTPClient` → `c` or `hc`, `ResponseWriter` → `w`. This is immediately recognizable to Go developers.

Guidelines:

- Use 1-2 characters (usually first letter of type)

- Be consistent across ALL methods of the type

- Avoid `this`, `self`, or full words

- For types with same first letter, use 2 characters: `HTTPClient` → `hc`

Code review: If you see varying receiver names in a type's methods, standardize them in one pass.

### 2.5 Use Named Result Parameters Sparingly and Purposefully

**Impact: LOW**

Named result parameters initialize return values to their zero values and enable "naked" returns. They improve readability in specific cases but can harm it in others. Use them when they add genuine clarity, not as a habit.

**Incorrect:**

```go
// Unnecessary: single result, name adds no information
func StoreCustomer(customer Customer) (err error) {
    // "err" name doesn't help the reader
    return db.Save(customer)
}

// Unnecessary: obvious what the return means
func Add(a, b int) (result int) {
    result = a + b
    return  // Naked return — reader must remember what "result" is
}
```

**Correct:**

```go
// Good: disambiguates multiple returns of the same type
type locator interface {
    // Without names: unclear which float32 is lat and which is lng
    // getCoordinates(address string) (float32, float32, error)

    // With names: signature is self-documenting
    getCoordinates(address string) (lat, lng float32, err error)
}

// Good: convenience when named params simplify initialization
func ReadFull(r io.Reader, buf []byte) (n int, err error) {
    // n and err are pre-initialized to 0 and nil
    for len(buf) > 0 && err == nil {
        var nr int
        nr, err = r.Read(buf)
        n += nr
        buf = buf[nr:]
    }
    return  // Naked return is acceptable in short functions
}

// Rule: don't mix naked returns and explicit returns in the same function
```

Why this matters: Named result parameters are initialized to their zero values at function entry. This can cause subtle bugs — see mistake #44. Use them when: multiple returns have the same type and names disambiguate, or when pre-initialization provides a genuine benefit. Avoid naked returns in long functions — readers must scroll back to the signature to understand what's being returned. Keep naked returns to short functions only.

### 2.6 Watch for Subtle Bugs with Named Result Parameters

**Impact: MEDIUM**

Named result parameters are initialized to their zero values when a function begins. This means an early return that doesn't explicitly assign the named parameter will return the zero value — which can be a silent bug, especially for `error`.

**Incorrect:**

```go
// BUG: err is a named result parameter, initialized to nil
func (l loc) getCoordinates(ctx context.Context, address string) (
    lat, lng float32, err error) {

    isValid := l.validateAddress(address)
    if !isValid {
        return 0, 0, errors.New("invalid address")
    }

    if ctx.Err() != nil {
        // BUG: we return "err" but never assigned it!
        // err is still nil (its zero value)
        return 0, 0, err  // Always returns nil — context error is swallowed!
    }

    // Get coordinates...
    return lat, lng, nil
}
```

**Correct:**

```go
// Option 1: Assign before returning
func (l loc) getCoordinates(ctx context.Context, address string) (
    lat, lng float32, err error) {

    isValid := l.validateAddress(address)
    if !isValid {
        return 0, 0, errors.New("invalid address")
    }

    if err = ctx.Err(); err != nil {  // Assign ctx.Err() to err first
        return 0, 0, err
    }

    return lat, lng, nil
}

// Option 2: Avoid named result parameters when they add risk
func (l loc) getCoordinates(ctx context.Context, address string) (float32, float32, error) {
    if !l.validateAddress(address) {
        return 0, 0, errors.New("invalid address")
    }
    if err := ctx.Err(); err != nil {
        return 0, 0, err  // Can't forget to assign — no named param to confuse us
    }
    lat, lng := l.computeCoordinates(address)
    return lat, lng, nil
}
```

Why this matters: The bug compiles cleanly. The named `err` parameter is `nil` by default. The code `return 0, 0, err` looks correct but returns `nil` when the context was cancelled. This is especially dangerous in error handling paths where you intend to propagate an error but accidentally return nil. Named result parameters that shadow error handling are among the subtlest Go bugs.

---

## References

1. [https://go.dev/doc/effective_go](https://go.dev/doc/effective_go)
2. [https://go.dev/wiki/CodeReviewComments](https://go.dev/wiki/CodeReviewComments)
3. [https://github.com/golang/go/wiki/CommonMistakes](https://github.com/golang/go/wiki/CommonMistakes)
