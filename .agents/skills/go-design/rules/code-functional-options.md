## Use the Functional Options Pattern for Configuration

**Impact: MEDIUM**

When a constructor has many optional parameters, avoid long parameter lists or config structs that must be zeroed out. The functional options pattern provides a clean, extensible API that's easy to read, optional by default, and backward-compatible when new options are added.

**Incorrect:**

```go
// Growing parameter list — breaks callers when adding new options
func NewServer(host string, port int, timeout time.Duration, maxConns int, tls bool) *Server

// Config struct — callers must know zero values and field names
type ServerConfig struct {
    Host     string
    Port     int
    Timeout  time.Duration
    MaxConns int
    TLS      bool
}
func NewServer(cfg ServerConfig) *Server
```

**Correct:**

```go
type options struct {
    port     int
    timeout  time.Duration
    maxConns int
    tls      bool
}

type Option func(*options) error

func WithPort(port int) Option {
    return func(o *options) error {
        if port < 1 || port > 65535 {
            return fmt.Errorf("invalid port: %d", port)
        }
        o.port = port
        return nil
    }
}

func WithTimeout(d time.Duration) Option {
    return func(o *options) error {
        o.timeout = d
        return nil
    }
}

func NewServer(host string, opts ...Option) (*Server, error) {
    o := &options{
        port:    8080,          // sensible defaults
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        if err := opt(o); err != nil {
            return nil, err
        }
    }
    return &Server{host: host, opts: o}, nil
}

// Usage: clean, self-documenting, options are optional
srv, err := NewServer("localhost",
    WithPort(9090),
    WithTimeout(60*time.Second),
)
```

Why this matters: Adding a new option never breaks existing callers. Each option is self-documenting. Invalid configurations can be caught at construction time with meaningful errors. The pattern scales from 2 to 20+ options without API churn.
