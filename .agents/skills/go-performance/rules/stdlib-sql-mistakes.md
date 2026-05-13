## Common SQL Mistakes: Open, Pool, Prepared Statements, Null, Rows.Err

**Impact: HIGH**

Five common mistakes when using `database/sql`: (1) assuming `sql.Open` establishes a connection; (2) not configuring the connection pool; (3) not using prepared statements; (4) mishandling NULL values; (5) not checking `rows.Err()` after iteration.

**Mistake 1: sql.Open doesn't guarantee a connection**

```go
// sql.Open may only validate arguments, not establish a connection
db, err := sql.Open("mysql", dsn)
// err might be nil even if the database is unreachable!

// CORRECT: Call Ping to verify connectivity at startup
db, err := sql.Open("mysql", dsn)
if err != nil { return err }
if err := db.Ping(); err != nil {
    return fmt.Errorf("cannot reach database: %w", err)
}
```

**Mistake 2: Configure the connection pool for production**

```go
// Defaults are inappropriate for production:
// SetMaxOpenConns: unlimited (can overwhelm the database)
// SetMaxIdleConns: 2 (too low for concurrent applications)

db.SetMaxOpenConns(25)                 // Limit total open connections
db.SetMaxIdleConns(25)                 // Keep up to 25 idle connections ready
db.SetConnMaxIdleTime(5 * time.Minute) // Release idle connections after 5 min
db.SetConnMaxLifetime(2 * time.Hour)   // Recycle connections periodically
```

**Mistake 3: Use prepared statements for repeated queries**

```go
// Repeated query without prepared statement: recompiled each time (slow + injection risk)
rows, err := db.Query("SELECT * FROM orders WHERE id = ?", id)

// CORRECT: Prepare once, execute many times
stmt, err := db.Prepare("SELECT * FROM orders WHERE id = ?")
if err != nil { return err }
defer stmt.Close()
rows, err := stmt.Query(id)
```

**Mistake 4: Handle NULL values with sql.NullXXX or pointers**

```go
// NULL column → Scan error: "converting NULL to string is unsupported"
var department string
rows.Scan(&department)  // BUG if column can be NULL

// CORRECT option 1: use pointer
var department *string
rows.Scan(&department)  // nil if NULL

// CORRECT option 2: use sql.NullString
var department sql.NullString
rows.Scan(&department)
if department.Valid {
    fmt.Println(department.String)
}
// Also: sql.NullBool, sql.NullInt32, sql.NullInt64, sql.NullFloat64, sql.NullTime
```

**Mistake 5: Check rows.Err() after iteration**

```go
// rows.Next() stops when done OR when an error occurs — you can't tell which
for rows.Next() {
    // ...
}
// MISSING: check why the loop stopped
if err := rows.Err(); err != nil {  // Required after every rows.Next() loop
    return err
}
```

Why this matters: Each of these mistakes has real production consequences: silent connection failures at startup, database overload from unconstrained connections, performance and security issues from unprepared statements, panics from NULL values, and silent data loss from unchecked iteration errors.
