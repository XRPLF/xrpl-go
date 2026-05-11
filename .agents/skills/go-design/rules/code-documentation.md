## Document All Exported Elements

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
