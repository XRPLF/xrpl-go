## Avoid Overusing the any Type

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
