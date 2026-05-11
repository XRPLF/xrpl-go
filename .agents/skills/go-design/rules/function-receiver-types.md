## Use Consistent, Short Receiver Names

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
