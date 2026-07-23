## Use net/http/httptest and testing/iotest for HTTP and I/O Tests

**Impact: MEDIUM**

Go's standard library includes two underused testing utilities: `net/http/httptest` for testing HTTP handlers and clients without a real network, and `testing/iotest` for testing custom `io.Reader`/`io.Writer` implementations and error resilience.

**httptest — testing an HTTP handler:**

```go
func Handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("X-API-VERSION", "1.0")
    b, _ := io.ReadAll(r.Body)
    _, _ = w.Write(append([]byte("hello "), b...))
    w.WriteHeader(http.StatusCreated)
}

func TestHandler(t *testing.T) {
    // Build a fake request — no network needed
    req := httptest.NewRequest(http.MethodGet, "http://localhost", strings.NewReader("foo"))
    // Record the response
    w := httptest.NewRecorder()

    Handler(w, req)

    // Assert on headers, body, and status code
    if got := w.Result().Header.Get("X-API-VERSION"); got != "1.0" {
        t.Errorf("api version: expected 1.0, got %s", got)
    }
    body, _ := io.ReadAll(w.Result().Body)
    if got := string(body); got != "hello foo" {
        t.Errorf("body: expected hello foo, got %s", got)
    }
    if w.Result().StatusCode != http.StatusCreated {
        t.FailNow()
    }
}
```

**httptest — testing an HTTP client:**

```go
func TestDurationClientGet(t *testing.T) {
    // Start a local HTTP server — no Docker, no external services
    srv := httptest.NewServer(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            _, _ = w.Write([]byte(`{"duration": 314}`))
        },
    ))
    defer srv.Close()  // Shuts down the server after the test

    client := NewDurationClient()
    duration, err := client.GetDuration(srv.URL, 51.55, -0.12, 51.57, -0.13)
    if err != nil {
        t.Fatal(err)
    }
    if duration != 314*time.Second {
        t.Errorf("expected 314 seconds, got %v", duration)
    }
}
```

**iotest — testing a custom io.Reader:**

```go
// Test that LowerCaseReader correctly transforms input
func TestLowerCaseReader(t *testing.T) {
    err := iotest.TestReader(
        &LowerCaseReader{reader: strings.NewReader("aBcDeFgHiJ")},
        []byte("abcdefghij"),  // Expected output
    )
    if err != nil {
        t.Fatal(err)
    }
}

// Test that a function is resilient to read errors
func TestFoo_ToleratesReadErrors(t *testing.T) {
    // iotest.TimeoutReader fails on the second Read call, then succeeds
    err := foo(iotest.TimeoutReader(strings.NewReader(randomString(1024))))
    if err != nil {
        t.Fatal(err)
    }
}
```

**Available iotest helpers:**
- `iotest.TestReader` — validates a custom `io.Reader` behaves correctly (byte count, fills slice, etc.)
- `iotest.ErrReader` — returns a specified error on every read
- `iotest.HalfReader` — reads only half the requested bytes each call
- `iotest.OneByteReader` — reads one byte per call
- `iotest.TimeoutReader` — fails on the second read, then succeeds
- `iotest.TruncateWriter` — writes to an `io.Writer` but stops silently after n bytes

Why this matters: Spinning up real HTTP servers or writing ad-hoc test mocks for I/O interfaces wastes effort and is slower than necessary. `httptest.NewServer` starts a real (local) HTTP server in milliseconds and tears it down automatically. `httptest.NewRecorder` lets you inspect headers, body, and status without any network. `iotest.TestReader` validates all the subtle contracts of `io.Reader` (partial reads, EOF handling) that are easy to miss. Use these before reaching for Docker or hand-rolled mocks.
