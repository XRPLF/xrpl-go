## Follow Standard Go Project Layout

**Impact: MEDIUM**

Inconsistent project structure makes it harder for new contributors to navigate the codebase and for tools to find code. The Go community has converged on a standard layout that separates concerns and signals intent through directory names.

**Incorrect:**

```
myapp/
├── main.go          # Everything in root
├── server.go
├── db.go
├── utils.go         # Catch-all utility file
├── helpers.go       # Another catch-all
└── models.go
```

**Correct:**

```
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

Reference: [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
