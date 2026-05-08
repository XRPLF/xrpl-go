## select With Multiple Ready Channels Chooses Randomly, Not by Order

**Impact: MEDIUM**

When multiple `case` clauses in a `select` statement are ready simultaneously, Go chooses one **at random** — not in source order. Code that assumes the first `case` has priority will have non-deterministic behavior. Implement explicit prioritization when channel ordering matters.

**Incorrect:**

```go
// WRONG assumption: messageCh case runs first because it appears first
for {
    select {
    case v := <-messageCh:    // Assumed to have priority — it doesn't
        fmt.Println(v)
    case <-disconnectCh:
        fmt.Println("disconnected")
        return
    }
}
// If both channels have data, Go may pick disconnectCh first,
// causing some messages to be missed
```

**Correct:**

```go
// For a single producer: use inner select + default to drain messageCh first
for {
    select {
    case v := <-messageCh:
        fmt.Println(v)
    case <-disconnectCh:
        // Drain remaining messages before returning
        for {
            select {
            case v := <-messageCh:
                fmt.Println(v)
            default:
                fmt.Println("disconnected")
                return  // default fires when messageCh is empty
            }
        }
    }
}

// Alternative: use nil channels to remove a case once its source is done
func merge(ch1, ch2 <-chan int) <-chan int {
    ch := make(chan int, 1)
    go func() {
        for ch1 != nil || ch2 != nil {
            select {
            case v, open := <-ch1:
                if !open { ch1 = nil; break }  // Remove ch1 from select
                ch <- v
            case v, open := <-ch2:
                if !open { ch2 = nil; break }  // Remove ch2 from select
                ch <- v
            }
        }
        close(ch)
    }()
    return ch
}
```

Why this matters: The Go specification states that when multiple `select` cases are ready, one is chosen via uniform pseudo-random selection. This prevents starvation of a slow channel by a fast one, but it breaks any assumption about ordering. The `default` case inside an inner `for/select` provides a way to drain one channel before returning. Setting a channel to `nil` elegantly removes it from a `select` when it's closed, since receiving from a nil channel blocks forever.
