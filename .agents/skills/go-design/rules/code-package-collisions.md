## Avoid Package Name Collisions

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
