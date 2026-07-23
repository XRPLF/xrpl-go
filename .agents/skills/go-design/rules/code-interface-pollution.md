## Avoid Interface Pollution

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

Reference: [Go interfaces guide](https://go.dev/doc/faq#overloading)
