## Use Generics Appropriately

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

Reference: [Go generics proposal](https://go.dev/doc/faq#generics)
