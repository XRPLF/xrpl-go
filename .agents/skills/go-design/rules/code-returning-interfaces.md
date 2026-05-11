## Return Structs, Accept Interfaces

**Impact: MEDIUM**

Functions should return concrete types (structs or pointers) and accept interfaces as parameters. Returning interfaces from constructors forces callers to depend on an abstraction and prevents them from accessing concrete methods. The principle: "be conservative in what you return, be liberal in what you accept."

**Incorrect:**

```go
// Returning an interface from a constructor
type Logger interface {
    Info(msg string)
    Error(msg string, err error)
}

// Forces callers to use Logger interface, hiding concrete type
func NewLogger(level string) Logger {
    return &zapLogger{level: level}
}

// Callers can't access zapLogger-specific methods without type assertion
```

**Correct:**

```go
type Logger interface {
    Info(msg string)
    Error(msg string, err error)
}

type ZapLogger struct {
    level  string
    sugar  *zap.SugaredLogger
}

// Return concrete type — callers can use it as Logger interface or access specific methods
func NewZapLogger(level string) (*ZapLogger, error) {
    // ...
    return &ZapLogger{level: level, sugar: sugar}, nil
}

func (l *ZapLogger) Info(msg string)              { l.sugar.Info(msg) }
func (l *ZapLogger) Error(msg string, err error)  { l.sugar.Errorw(msg, "error", err) }
func (l *ZapLogger) Sync() error                  { return l.sugar.Sync() } // concrete method
```

Why this matters: When you return an interface, callers lose access to concrete methods without unsafe type assertions. They also become tightly coupled to your abstraction choice. Returning concrete types gives callers maximum flexibility — they can use the value as an interface, access concrete methods, or embed it. If a caller needs an interface, they define one themselves.

Exception: Returning `error` (an interface) is idiomatic and correct — it's a well-established Go convention.
