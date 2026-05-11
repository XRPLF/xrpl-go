## Avoid time.Sleep in Tests — Use Retry or Channel Synchronization Instead

**Impact: HIGH**

Using `time.Sleep` in tests to wait for goroutines or async operations makes tests flaky: there is no guarantee the sleep duration will be sufficient on a loaded machine. Replace sleeps with either a **retry loop** (poll until condition or timeout) or **channel synchronization** (block until the goroutine signals completion).

**Incorrect — passive sleep:**

```go
func TestGetBestFoo(t *testing.T) {
    mock := &publisherMock{}
    h := Handler{publisher: mock, n: 2}

    foo := h.getBestFoo(42)
    _ = foo

    time.Sleep(10 * time.Millisecond)  // Flaky: may not be enough on a slow machine
    published := mock.Get()
    if len(published) != 2 {
        t.Fatalf("expected 2, got %d", len(published))
    }
}
```

**Correct option 1 — retry loop:**

```go
// assert polls the assertion function up to maxRetry times, sleeping waitTime between tries.
func assert(t *testing.T, assertion func() bool, maxRetry int, waitTime time.Duration) {
    t.Helper()
    for i := 0; i < maxRetry; i++ {
        if assertion() {
            return
        }
        time.Sleep(waitTime)
    }
    t.Fail()
}

func TestGetBestFoo(t *testing.T) {
    mock := &publisherMock{}
    h := Handler{publisher: mock, n: 2}
    _ = h.getBestFoo(42)

    // Poll up to 30 times, 1ms apart — fast when it succeeds, bounded when it fails
    assert(t, func() bool {
        return len(mock.Get()) == 2
    }, 30, time.Millisecond)
}
```

**Correct option 2 — channel synchronization (preferred):**

```go
type publisherMock struct {
    ch chan []Foo
}

func (p *publisherMock) Publish(got []Foo) {
    p.ch <- got  // Signal the test goroutine immediately
}

func TestGetBestFoo(t *testing.T) {
    mock := &publisherMock{ch: make(chan []Foo, 1)}
    defer close(mock.ch)

    h := Handler{publisher: mock, n: 2}
    _ = h.getBestFoo(42)

    // Block until the goroutine publishes, with a timeout to avoid hanging
    select {
    case got := <-mock.ch:
        if len(got) != 2 {
            t.Fatalf("expected 2, got %d", len(got))
        }
    case <-time.After(time.Second):
        t.Fatal("timeout waiting for Publish")
    }
}
```

Why this matters: A passive `time.Sleep` introduces a fixed wait that is both too long (slows the test suite on fast machines) and too short (flaky on slow CI machines). A retry loop checks the condition as soon as it's ready, reducing wait time when things go right and bounding total wait when they go wrong. Channel synchronization is even better: it makes the test fully deterministic by blocking exactly until the work is done, with a timeout guard. Use the testing library `testify/assert`'s `Eventually` function as a ready-made retry helper.
