## Returning a Nil Pointer as an Interface Is Not Nil

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
