## Always Close HTTP Response Bodies and Other Transient Resources

**Impact: HIGH**

HTTP response bodies must be closed after use, even if you don't read them. Failing to close leaks memory and may prevent TCP connection reuse. Always `defer resp.Body.Close()` immediately after a successful HTTP response, and read the body before closing if you want keep-alive connections.

**Incorrect:**

```go
func getBody(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    // BUG: resp.Body is never closed — memory leak, TCP connection never reused
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}
```

**Correct:**

```go
func getBody(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer func() {
        if err := resp.Body.Close(); err != nil {
            log.Printf("failed to close response body: %v\n", err)
        }
    }()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}

// Even when you don't need the body, close it AND read it for connection reuse:
func getStatusCode(url string, body io.Reader) (int, error) {
    resp, err := http.Post(url, "application/json", body)
    if err != nil {
        return 0, err
    }
    defer func() {
        if err := resp.Body.Close(); err != nil {
            log.Printf("failed to close: %v\n", err)
        }
    }()

    // Read body to enable TCP connection reuse (even if we discard it)
    _, _ = io.Copy(io.Discard, resp.Body)
    return resp.StatusCode, nil
}
```

Why this matters: `http.Response.Body` implements `io.ReadCloser`. Closing the body returns the underlying TCP connection to the pool for reuse. If you close without reading, the default transport may close the connection entirely (preventing keep-alive). If you close after reading, the connection can be reused. Use `io.Copy(io.Discard, resp.Body)` to efficiently drain the body without storing it. Apply the same pattern to `sql.Rows` (defer `rows.Close()`), files (defer `file.Close()`), and any other `io.Closer`.
