## Define Interfaces on the Consumer Side

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

Reference: [Effective Go — Interfaces](https://go.dev/doc/effective_go#interfaces)
