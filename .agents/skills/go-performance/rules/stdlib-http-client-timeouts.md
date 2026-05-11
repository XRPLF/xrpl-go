## Always Set HTTP Client Timeouts

**Impact: CRITICAL**

The default HTTP client has no timeout—it waits forever for responses. This causes goroutine leaks and resource exhaustion when servers are slow or unresponsive.

**Incorrect:**

```go
func FetchData(url string) ([]byte, error) {
    resp, err := http.Get(url)  // Uses default client - no timeout!
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}
```

**Correct:**

```go
var httpClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   5 * time.Second,
        ResponseHeaderTimeout: 5 * time.Second,
    },
}

func FetchData(url string) ([]byte, error) {
    resp, err := httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}
```

Why this matters: Without timeouts, slow servers cause goroutines to hang indefinitely. In production, this leads to goroutine leaks, memory exhaustion, and cascading failures. Set `Client.Timeout` for overall request timeout and `Transport` timeouts for finer control over connection, TLS handshake, and response header phases.
