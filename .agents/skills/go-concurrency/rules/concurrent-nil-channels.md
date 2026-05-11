## Understand Nil Channel Behavior

**Impact: MEDIUM**

Nil channels block forever on send and receive operations. This is useful for disabling cases in select statements but causes deadlocks if misunderstood.

**Incorrect:**

```go
func merge(ch1, ch2 <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for {
            select {
            case v := <-ch1:  // If ch1 is nil, blocks forever!
                out <- v
            case v := <-ch2:
                out <- v
            }
        }
    }()
    return out
}
```

**Correct:**

```go
func merge(ch1, ch2 <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for ch1 != nil || ch2 != nil {
            select {
            case v, ok := <-ch1:
                if !ok {
                    ch1 = nil  // Disable this case
                    continue
                }
                out <- v
            case v, ok := <-ch2:
                if !ok {
                    ch2 = nil  // Disable this case
                    continue
                }
                out <- v
            }
        }
    }()
    return out
}
```

Why this matters: Setting a channel to `nil` makes select ignore that case, allowing graceful shutdown when one channel closes. Without this, select would continuously receive zero values from the closed channel.
