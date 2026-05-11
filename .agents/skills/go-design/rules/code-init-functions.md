## Avoid Misusing init Functions

**Impact: MEDIUM**

`init` functions run automatically at package initialization, before `main`. They look convenient for setup work, but they have significant drawbacks: they cannot return errors, they force the use of global state, they execute even in tests (causing side effects), and they make dependency injection impossible.

**Incorrect:**

```go
var db *sql.DB

func init() {
    // Cannot return an error — if this fails, only option is panic
    var err error
    db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        panic(err)  // Crashes the entire program on startup
    }
}

func GetUser(id int) (*User, error) {
    // Tests that import this package will trigger DB connection
    return queryUser(db, id)
}
```

**Correct:**

```go
func NewDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("opening db: %w", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("pinging db: %w", err)
    }
    return db, nil
}

func main() {
    db, err := NewDB(os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    server := NewServer(db)
    server.Run()
}
```

Why this matters: Constructor functions return errors, making failures explicit and handleable. They don't run until you call them, so tests that don't need a database won't trigger a connection attempt. Dependencies are passed explicitly, making the code easier to test and understand.

Use `init` only for truly static setup with no error cases: registering codecs, setting up global flag defaults, or initializing lookup tables that can't fail.
