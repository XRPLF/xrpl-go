## Use Idiomatic Getter and Setter Names

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
